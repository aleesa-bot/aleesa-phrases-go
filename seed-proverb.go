package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	config = ReadConfig()
	file := config.DataDir + "/proverb_src/proverb.txt"

	fh, err := os.Open(file)

	if err != nil {
		log.Errorf("Unable to open file %s: %s\n", file, err)
		os.Exit(1)
	}

	defer func(fh *os.File) {
		err := fh.Close()

		if err != nil {
			log.Errorf("Unable to close %s cleanly: %s", file, err)
		}
	}(fh)

	reader := bufio.NewReader(fh)

	// Откроем все наши БД
	var options pebble.Options
	// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными настройками
	// порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук, что нас не
	// устраивает, поэтому немного снизим эту цифру до более приемлемых значений
	options.L0CompactionFileThreshold = 8

	pebbleDB.proverb.name = "proverb_db"
	pebbleDB.proverb.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.proverb.name, &options)
	if err != nil {
		log.Errorf("Unable to open %s: %s", pebbleDB.proverb.name, err)
		os.Exit(1)
	}

	pebbleDB.count.name = "count_db"
	pebbleDB.count.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.count.name, &options)
	if err != nil {
		log.Errorf("Unable to open %s: %s", pebbleDB.count.name, err)
		os.Exit(1)
	}

	for {
		counter += 1
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				key := fmt.Sprintf("%d", counter)
				value := strings.Trim(line, "\n\r\t ")

				if err := StoreKV(pebbleDB.proverb.db, pebbleDB.proverb.name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in proverb_db: %s", key, value, err)
					os.Exit(1)
				}
				break
			}

			log.Errorf("Unable to read string from %s: %s\n", file, err)
			os.Exit(1)
		}

		key := fmt.Sprintf("%d", counter)
		value := strings.Trim(line, "\n\r\t ")

		if err := StoreKV(pebbleDB.proverb.db, pebbleDB.proverb.name, key, value, true); err != nil {
			log.Errorf("Unable to store key %s with value %s in proverb_db: %s", key, value, err)
			os.Exit(1)
		}
	}

	err = pebbleDB.proverb.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.proverb.name, err)
		os.Exit(1)
	}

	err = pebbleDB.count.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.count.name, err)
		os.Exit(1)
	}

	log.Printf("Seeded %d proverbs in database.\n", counter)
}
