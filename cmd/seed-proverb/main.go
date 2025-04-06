package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"aleesa-phrases-go/internal/lib"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	lib.Config = lib.ReadConfig()
	file := lib.Config.DataDir + "/proverb_src/proverb.txt"

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

	lib.PebbleDB.Proverb.Name = "proverb_db"
	lib.PebbleDB.Proverb.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Proverb.Name, &options)
	if err != nil {
		log.Errorf("Unable to open %s: %s", lib.PebbleDB.Proverb.Name, err)
		os.Exit(1)
	}

	lib.PebbleDB.Count.Name = "count_db"
	lib.PebbleDB.Count.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Count.Name, &options)
	if err != nil {
		log.Errorf("Unable to open %s: %s", lib.PebbleDB.Count.Name, err)
		os.Exit(1)
	}

	for {
		lib.Counter += 1
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				key := fmt.Sprintf("%d", lib.Counter)
				value := strings.Trim(line, "\n\r\t ")

				if err := lib.StoreKV(lib.PebbleDB.Proverb.DB, lib.PebbleDB.Proverb.Name, key, value, true); err != nil {
					log.Errorf("Unable to store key %s with value %s in proverb_db: %s", key, value, err)
					os.Exit(1)
				}
				break
			}

			log.Errorf("Unable to read string from %s: %s\n", file, err)
			os.Exit(1)
		}

		key := fmt.Sprintf("%d", lib.Counter)
		value := strings.Trim(line, "\n\r\t ")

		if err := lib.StoreKV(lib.PebbleDB.Proverb.DB, lib.PebbleDB.Proverb.Name, key, value, true); err != nil {
			log.Errorf("Unable to store key %s with value %s in proverb_db: %s", key, value, err)
			os.Exit(1)
		}
	}

	if err = lib.PebbleDB.Proverb.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Proverb.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Proverb.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Proverb.Name, err)
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

	log.Printf("Seeded %d proverbs in database.\n", lib.Counter)
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
