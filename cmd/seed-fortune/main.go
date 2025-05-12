package main

import (
	"fmt"
	"os"
	"strings"

	"aleesa-phrases-go/internal/lib"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	lib.Config = lib.ReadConfig()
	srcDir := lib.Config.DataDir + "/fortune_src"

	// Вычитывать сразу весь список файлов - плохая практика, но мы предполагаем, что тут их немного.
	files, err := os.ReadDir(srcDir)

	if err != nil {
		log.Errorf("Unable to read dir %s: %s\n", srcDir, err)
		os.Exit(1)
	}

	// Откроем все наши БД
	var options pebble.Options
	// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными настройками
	// порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук.
	options.L0CompactionFileThreshold = 300

	lib.PebbleDB.Fortune.Name = "fortune_db"
	lib.PebbleDB.Fortune.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Fortune.Name, &options)
	if err != nil {
		log.Errorf("Unable to open %s db: %s\n", lib.PebbleDB.Fortune.Name, err)
		os.Exit(1)
	}

	lib.PebbleDB.Count.Name = "count_db"
	lib.PebbleDB.Count.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Count.Name, &options)
	if err != nil {
		log.Errorf("Unable to open %s db: %s\n", lib.PebbleDB.Count.Name, err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "." || file.Name() == ".." {
			continue
		}

		// "Заглатывать" файл целиком - плохая практика, но мы предополагаем, что файлы будут небольшие.
		fortuneFile := fmt.Sprintf("%s/%s", srcDir, file.Name())
		bytesRead, err := os.ReadFile(fortuneFile)

		if err != nil {
			log.Errorf("Unable to read %s: %s\n", file.Name(), err)
			continue
		}

		fileContent := string(bytesRead)
		lines := strings.Split(fileContent, "\n%\n")

		for _, line := range lines {
			lib.Counter += 1
			key := fmt.Sprintf("%d", lib.Counter)
			value := strings.Trim(line, "\n\r\t ")

			if err := lib.StoreKV(lib.PebbleDB.Fortune.DB, lib.PebbleDB.Fortune.Name, key, value, true); err != nil {
				log.Errorf("Unable to store key %s with value %s in fortune_db: %s", key, value, err)
				os.Exit(1)
			}
		}
	}

	if err = lib.PebbleDB.Fortune.DB.Flush(); err != nil {
		log.Errorf("Unable to flush %s: %s", lib.PebbleDB.Fortune.Name, err)
		os.Exit(1)
	}

	if err = lib.PebbleDB.Fortune.DB.Close(); err != nil {
		log.Errorf("Unable to close %s: %s", lib.PebbleDB.Fortune.Name, err)
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

	log.Printf("Seeded %d fortunes in database.\n", lib.Counter)
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
