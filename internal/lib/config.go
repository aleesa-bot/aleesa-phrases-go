package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hjson/hjson-go"
	log "github.com/sirupsen/logrus"
)

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

		buf, err := os.ReadFile(location)

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
			sampleConfig.ForwardsMax = ForwardMax
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

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
