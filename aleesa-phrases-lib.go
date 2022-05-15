package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/hjson/hjson-go"
	log "github.com/sirupsen/logrus"
)

// MyConfig описывает структурку, получаемую при загрузке и распарсивании конфига
type MyConfig struct {
	Server      string `json:"server,omitempty"`
	Port        int    `json:"port,omitempty"`
	LogLevel    string `json:"loglevel,omitempty"`
	Log         string `json:"log,omitempty"`
	Channel     string `json:"channel,omitempty"`
	DataDir     string `json:"data_dir,omitempty"`
	CSign       string `json:"csign,omitempty"`
	ForwardsMax int64  `json:"forwards_max,omitempty"`
}

// To break circular message forwarding we must set some sane default, it can be overridden via config
var forwardMax int64 = 5

// Config - это у нас глобальная штука :)
var config MyConfig

// Просто счётчик, использован в seed-программах
var counter = 0

var dow = []string{
	"monday",
	"tuesday",
	"wednesday",
	"thursday",
	"friday",
	"saturday",
	"sunday",
}

// Struct с дескрипторами БД
var pebbleDB struct {
	// Счётчик записей для каждой БД
	count struct {
		db   *pebble.DB
		name string
	}
	// База модуля пословиц
	proverb struct {
		db   *pebble.DB
		name string
	}
	// База модуля fortune
	fortune struct {
		db   *pebble.DB
		name string
	}
	// 7 баз по дням недели модуля friday
	monday struct {
		db   *pebble.DB
		name string
	}
	tuesday struct {
		db   *pebble.DB
		name string
	}
	wednesday struct {
		db   *pebble.DB
		name string
	}
	thursday struct {
		db   *pebble.DB
		name string
	}
	friday struct {
		db   *pebble.DB
		name string
	}
	saturday struct {
		db   *pebble.DB
		name string
	}
	sunday struct {
		db   *pebble.DB
		name string
	}
}

// ReadConfig читает и валидирует конфиг, а также выставляет некоторые default-ы, если значений для параметров в конфиге нет
func ReadConfig() MyConfig {
	configLoaded := false
	var config MyConfig
	executablePath, err := os.Executable()

	if err != nil {
		log.Errorf("Unable to get current executable path: %s", err)
	}

	configJSONPath := fmt.Sprintf("%s/data/config.json", filepath.Dir(executablePath))

	locations := []string{
		"~/.aleesa-phrases-go.json",
		"~/aleesa-phrases-go.json",
		"/etc/aleesa-phrases-go.json",
		configJSONPath,
	}

	for _, location := range locations {
		fileInfo, err := os.Stat(location)

		// Предполагаем, что файла либо нет, либо мы не можем его прочитать, второе надо бы логгировать, но пока забьём
		if err != nil {
			continue
		}

		// Конфиг-файл длинноват для конфига, попробуем следующего кандидата
		if fileInfo.Size() > 65535 {
			log.Warnf("Config file %s is too long for config, skipping", location)
			continue
		}

		buf, err := ioutil.ReadFile(location)

		// Не удалось прочитать, попробуем следующего кандидата
		if err != nil {
			log.Warnf("Skip reading config file %s: %s", location, err)
			continue
		}

		// Исходя из документации, hjson какбы умеет парсить "кривой" json, но парсит его в map-ку.
		// Интереснее на выходе получить структурку: то есть мы вначале конфиг преобразуем в map-ку, затем эту map-ку
		// сериализуем в json, а потом json преврщааем в стркутурку. Не очень эффективно, но он и не часто требуется.
		var sampleConfig MyConfig
		var tmp map[string]interface{}
		err = hjson.Unmarshal(buf, &tmp)

		// Не удалось распарсить - попробуем следующего кандидата
		if err != nil {
			log.Warnf("Skip parsing config file %s: %s", location, err)
			continue
		}

		tmpJSON, err := json.Marshal(tmp)

		// Не удалось преобразовать map-ку в json
		if err != nil {
			log.Warnf("Skip parsing config file %s: %s", location, err)
			continue
		}

		if err := json.Unmarshal(tmpJSON, &sampleConfig); err != nil {
			log.Warnf("Skip parsing config file %s: %s", location, err)
			continue
		}

		// Валидируем значения из конфига
		if sampleConfig.Server == "" {
			sampleConfig.Server = "localhost"
		}

		if sampleConfig.Port == 0 {
			sampleConfig.Port = 6379
		}

		if sampleConfig.LogLevel == "" {
			sampleConfig.LogLevel = "info"
		}

		// sampleConfig.Log = "" if not set
		// Если путь не задан или является относительным, вычисляем его относительно бинарника
		if sampleConfig.DataDir == "" || sampleConfig.DataDir[0:1] != "/" {
			binaryPath, err := os.Executable()

			if err != nil {
				log.Errorf("Unable to get binary name: %s", err)
				os.Exit(1)
			}

			if sampleConfig.DataDir == "" {
				sampleConfig.DataDir = fmt.Sprintf("%s/data", filepath.Dir(binaryPath))
				log.Infof("Field data_dir is not set in config, using %s", sampleConfig.DataDir)
			} else {
				sampleConfig.DataDir = fmt.Sprintf("%s/%s", filepath.Dir(binaryPath), sampleConfig.DataDir)
			}
		}

		if sampleConfig.Channel == "" {
			log.Errorf("Channel field in config file %s must be set", location)
		}

		if sampleConfig.CSign == "" {
			log.Errorf("CSign field in config file %s must be set", location)
		}

		if sampleConfig.ForwardsMax == 0 {
			sampleConfig.ForwardsMax = forwardMax
		}

		config = sampleConfig
		configLoaded = true
		log.Infof("Using %s as config file", location)
		break
	}

	if !configLoaded {
		log.Error("Config was not loaded! Refusing to start.")
		os.Exit(1)
	}

	return config
}

// StoreKV сохраняет в указанной бд ключ и значение
func StoreKV(db *pebble.DB, dbName string, key string, value string, useCounter bool) error {
	var kArray = []byte(key)
	var vArray = []byte(value)

	err := db.Set(kArray, vArray, pebble.Sync)

	if useCounter {
		// У pebble нет в арсенале функции count(), так что "изобретём" её
		if err == nil {
			amount, err := FetchV(pebbleDB.count.db, dbName)

			if err != nil {
				// Такого ключа в базе нет
				if errors.Is(err, pebble.ErrNotFound) {
					amount = fmt.Sprintf("%d", 0)
					// Такой базы нет
				} else if errors.Is(err, fs.ErrNotExist) {
					amount = fmt.Sprintf("%d", 0)
				} else {
					return err
				}
			}

			amountNum, err := strconv.Atoi(amount)

			// Хуйня в базе, данные должны проходить конвертацию в int
			if err != nil {
				return err
			}

			amountNum += 1
			amount = fmt.Sprintf("%d", amountNum)

			err = pebbleDB.count.db.Set([]byte(dbName), []byte(amount), pebble.Sync)

			if err != nil {
				return err
			}
		}
	}

	return err
}

// FetchV достаёт значение по ключу
func FetchV(db *pebble.DB, key string) (string, error) {
	var kArray = []byte(key)

	var vArray []byte
	var valueString = ""

	vArray, closer, err := db.Get(kArray)

	if err != nil {
		return valueString, err
	}

	valueString = string(vArray)
	err = closer.Close()

	return valueString, err
}

// FetchRandomV достаёт рандомное значение из указанной базы (count_db для указанной db должна содержать количество записей)
func FetchRandomV(db *pebble.DB, dbName string) (string, error) {
	answer := ""
	amount, err := FetchV(pebbleDB.count.db, dbName)

	if err != nil {
		// Такого ключа в базе нет
		if errors.Is(err, pebble.ErrNotFound) {
			return answer, fmt.Errorf("looks like %s is not seeded or count_db is missing", dbName)
			// Такой базы нет
		} else if errors.Is(err, fs.ErrNotExist) {
			return answer, fmt.Errorf("unable to get random value from %s: count_db is missing", dbName)
		} else {
			return answer, fmt.Errorf("something going wrong with %s: %s", dbName, err)
		}
	}

	amountNum, err := strconv.Atoi(amount)

	// Хуйня в базе, данные должны проходить конвертацию в int
	if err != nil {
		return answer, fmt.Errorf("garbage in count_db while getting info about %s: %s", dbName, err)
	}

	phraseNum := 1 + rand.Intn(amountNum)
	phraseItem := fmt.Sprintf("%d", phraseNum)

	answer, err = FetchV(db, phraseItem)

	if err != nil {
		return "", fmt.Errorf("an error occured while getting answer from %s: %s", dbName, err)
	}

	return answer, nil
}
