package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

func main() {
	config = ReadConfig()
	srcDir := config.DataDir + "/fortune_src"

	// Вычитывать сразу весь список файлов - плохая практика, но мы предполагаем, что тут их немного.
	files, err := os.ReadDir(srcDir)

	if err != nil {
		log.Errorf("Unable to read dir %s: %s\n", srcDir, err)
		os.Exit(1)
	}

	// Откроем все наши БД
	var options pebble.Options
	// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными настройками
	// порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук, что нас не
	// устраивает, поэтому немного снизим эту цифру до более приемлемых значений
	options.L0CompactionFileThreshold = 8

	pebbleDB.fortune.name = "fortune_db"
	pebbleDB.fortune.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.fortune.name, &options)
	if err != nil {
		log.Errorf("Unable to open %s db: %s\n", pebbleDB.fortune.name, err)
		os.Exit(1)
	}

	pebbleDB.count.name = "count_db"
	pebbleDB.count.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.count.name, &options)
	if err != nil {
		log.Errorf("Unable to open %s db: %s\n", pebbleDB.count.name, err)
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
			counter += 1
			key := fmt.Sprintf("%d", counter)
			value := strings.Trim(line, "\n\r\t ")

			if err := StoreKV(pebbleDB.fortune.db, pebbleDB.fortune.name, key, value, true); err != nil {
				log.Errorf("Unable to store key %s with value %s in fortune_db: %s", key, value, err)
				os.Exit(1)
			}
		}
	}

	err = pebbleDB.fortune.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.fortune.name, err)
		os.Exit(1)
	}

	err = pebbleDB.count.db.Close()
	if err != nil {
		log.Errorf("Unable to close %s: %s", pebbleDB.count.name, err)
		os.Exit(1)
	}

	log.Printf("Seeded %d fortunes in database.\n", counter)
}
