package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	config = ReadConfig()
	var err error

	// Откроем все нужные нам для работы базюльки
	var options pebble.Options
	// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными настройками
	// порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук, что нас не
	// устраивает, поэтому немного снизим эту цифру до более приемлемых значений
	options.L0CompactionFileThreshold = 8

	pebbleDB.monday.name = "monday"
	pebbleDB.monday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.monday.name, &options)
	if err != nil {
		log.Errorf("Unable to open monday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.tuesday.name = "tuesday"
	pebbleDB.tuesday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.tuesday.name, &options)
	if err != nil {
		log.Errorf("Unable to open tuesday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.wednesday.name = "wednesday"
	pebbleDB.wednesday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.wednesday.name, &options)
	if err != nil {
		log.Errorf("Unable to open wednesday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.thursday.name = "thursday"
	pebbleDB.thursday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.thursday.name, &options)
	if err != nil {
		log.Errorf("Unable to open thursday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.friday.name = "friday"
	pebbleDB.friday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.friday.name, &options)
	if err != nil {
		log.Errorf("Unable to open friday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.saturday.name = "saturday"
	pebbleDB.saturday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.saturday.name, &options)
	if err != nil {
		log.Errorf("Unable to open saturday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.sunday.name = "sunday"
	pebbleDB.sunday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.sunday.name, &options)
	if err != nil {
		log.Errorf("Unable to open sunday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.count.name = "count_db"
	pebbleDB.count.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.count.name, &options)
	if err != nil {
		log.Errorf("Unable to open count_db: %s\n", err)
		os.Exit(1)
	}

	file := config.DataDir + "/friday_src/friday.txt"
	bytesRead, err := os.ReadFile(file)

	if err != nil {
		log.Errorf("Unable to read %s: %s\n", file, err)
		os.Exit(1)
	}

	fileContent := string(bytesRead)

	// Для каждой базюльки нам нужен отдельный счётчик
	var counter = []int{0, 0, 0, 0, 0, 0, 0}

	re := regexp.MustCompile("(\n\r|\r\n|\n|\r)")

	for _, line := range re.Split(fileContent, -1) {
		data := strings.Split(line, " || ")

		if data[0] == "" || data[1] == "" {
			continue
		}

		for _, day := range strings.Split(data[1], "") {
			dayNum, err := strconv.Atoi(day)

			if err != nil {
				log.Errorf("Unable to seed phrase to db phrase is: %s, error is: %s\n", data[0], err)
				continue
			}

			if dayNum > 7 || dayNum < 1 {
				log.Errorf("Unable to seed phrase to db phrase is: %s, daynum must be in range 1..7\n", data[0])
				continue
			}

			dayNum -= 1
			database := fmt.Sprintf("friday_db/%s", dow[dayNum])
			counter[dayNum] += 1
			key := fmt.Sprintf("%d", counter[dayNum])
			value := strings.Trim(data[0], "\n\r\t ")

			switch dayNum {
			case 0:
				if err := StoreKV(pebbleDB.monday.db, pebbleDB.monday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 1:
				if err := StoreKV(pebbleDB.tuesday.db, pebbleDB.tuesday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 2:
				if err := StoreKV(pebbleDB.wednesday.db, pebbleDB.wednesday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 3:
				if err := StoreKV(pebbleDB.thursday.db, pebbleDB.thursday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 4:
				if err := StoreKV(pebbleDB.friday.db, pebbleDB.friday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 5:
				if err := StoreKV(pebbleDB.saturday.db, pebbleDB.saturday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 6:
				if err := StoreKV(pebbleDB.sunday.db, pebbleDB.sunday.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			}
		}
	}

	// Не забудем позакрывать все БД к лешему.
	err = pebbleDB.monday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.monday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.tuesday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.tuesday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.wednesday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.wednesday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.thursday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.thursday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.friday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.friday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.saturday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.saturday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.sunday.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.sunday.name, err)
		os.Exit(1)
	}

	err = pebbleDB.count.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.count.name, err)
		os.Exit(1)
	}

	for i, amount := range counter {
		log.Printf("Seeded %d phrases in friday module %s database.\n", amount, dow[i])
	}
}
