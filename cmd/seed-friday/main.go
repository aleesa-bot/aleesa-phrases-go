package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"aleesa-phrases-go/internal/lib"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	lib.Config = lib.ReadConfig()
	var err error

	// Откроем все нужные нам для работы базюльки
	var options pebble.Options
	// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными настройками
	// порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук.
	options.L0CompactionFileThreshold = 300

	lib.PebbleDB.Monday.Name = "monday"
	lib.PebbleDB.Monday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Monday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open monday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Tuesday.Name = "tuesday"
	lib.PebbleDB.Tuesday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Tuesday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open tuesday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Wednesday.Name = "wednesday"
	lib.PebbleDB.Wednesday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Wednesday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open wednesday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Thursday.Name = "thursday"
	lib.PebbleDB.Thursday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Thursday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open thursday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Friday.Name = "friday"
	lib.PebbleDB.Friday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Friday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open friday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Saturday.Name = "saturday"
	lib.PebbleDB.Saturday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Saturday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open saturday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Sunday.Name = "sunday"
	lib.PebbleDB.Sunday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Sunday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open sunday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Count.Name = "count_db"
	lib.PebbleDB.Count.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Count.Name, &options)
	if err != nil {
		log.Errorf("Unable to open count_db: %s\n", err)
		os.Exit(1)
	}

	file := lib.Config.DataDir + "/friday_src/friday.txt"
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
			database := fmt.Sprintf("friday_db/%s", lib.Dow[dayNum])
			counter[dayNum] += 1
			key := fmt.Sprintf("%d", counter[dayNum])
			value := strings.Trim(data[0], "\n\r\t ")

			switch dayNum {
			case 0:
				if err := lib.StoreKV(lib.PebbleDB.Monday.DB, lib.PebbleDB.Monday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 1:
				if err := lib.StoreKV(lib.PebbleDB.Tuesday.DB, lib.PebbleDB.Tuesday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 2:
				if err := lib.StoreKV(lib.PebbleDB.Wednesday.DB, lib.PebbleDB.Wednesday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 3:
				if err := lib.StoreKV(lib.PebbleDB.Thursday.DB, lib.PebbleDB.Thursday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 4:
				if err := lib.StoreKV(lib.PebbleDB.Friday.DB, lib.PebbleDB.Friday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 5:
				if err := lib.StoreKV(lib.PebbleDB.Saturday.DB, lib.PebbleDB.Saturday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			case 6:
				if err := lib.StoreKV(lib.PebbleDB.Sunday.DB, lib.PebbleDB.Sunday.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in %s: %s", key, value, database, err)
					os.Exit(1)
				}
			}
		}
	}

	// Не забудем позакрывать все БД к лешему.
	if err = lib.PebbleDB.Monday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Monday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Monday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Monday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Tuesday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Tuesday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Tuesday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Tuesday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Wednesday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Wednesday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Wednesday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Wednesday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Thursday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Thursday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Thursday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Thursday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Friday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Friday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Friday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Friday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Saturday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Saturday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Saturday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Saturday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Sunday.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Sunday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Sunday.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Sunday.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Count.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Count.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Count.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Count.Name, err)
		os.Exit(1)
	}

	for i, amount := range counter {
		log.Printf("Seeded %d phrases in friday module %s database.\n", amount, lib.Dow[i])
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
