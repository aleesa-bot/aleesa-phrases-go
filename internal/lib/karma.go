package lib

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

// Изменим карму фразе
func changeKarma(chatID string, phrase string, action string) string {
	var chatHash = sha256.Sum256([]byte(chatID))
	var database = fmt.Sprintf("karma_db/%x", chatHash)
	var karmaInt int
	var answer string
	var err error

	if phrase = strings.TrimSpace(phrase); phrase == "" {
		phrase = " "
	}

	if _, ok := KarmaDB[database]; !ok {
		var options pebble.Options
		// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными
		// настройками порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук.
		options.L0CompactionFileThreshold = 300

		KarmaDB[database], err = pebble.Open(Config.DataDir+"/"+database, &options)

		if err != nil {
			log.Errorf("Unable to open karma db, %s: %s\n", database, err)

			if phrase == " " {
				return "Карма пустоты равна 0"
			} else {
				return fmt.Sprintf("Карма %s равна 0", phrase)
			}
		}
	}

	karmaString, err := FetchV(KarmaDB[database], phrase)

	if err != nil {
		// Такого ключа в базе нет
		if errors.Is(err, pebble.ErrNotFound) {
			karmaInt = 0
			// Такой бд нет
		} else if errors.Is(err, fs.ErrNotExist) {
			karmaInt = 0
		} else if errors.Is(err, oserror.ErrNotExist) {
			karmaInt = 0
			// Какая-то ещё ошибка
		} else {
			log.Errorf("Unable to Get karma of %s in database %s: %s", phrase, database, err)

			if phrase == "" {
				return fmt.Sprint("Карма пустоты равна 0")
			} else {
				return fmt.Sprintf("Карма %s равна 0", phrase)
			}
		}
	} else {
		if karmaInt, err = strconv.Atoi(karmaString); err != nil {
			karmaInt = 0
			log.Errorf("Got garbage instead of karma for %s in database %s: %s", phrase, database, err)
		}
	}

	switch action {
	case "++":
		karmaInt += 1
	case "--":
		karmaInt -= 1
	default:
		log.Errorf("Function changeKarma got strange action neither ++ nor --: %s", action)

		if phrase == "" {
			return fmt.Sprint("Карма пустоты равна 0")
		} else {
			return fmt.Sprintf("Карма %s равна 0", phrase)
		}
	}

	// Поменяли карму, теперь пора превратить значение int в string :)
	karmaString = fmt.Sprintf("%d", karmaInt)

	if err := StoreKV(KarmaDB[database], "", phrase, karmaString, false); err != nil {
		log.Errorf("Unable to Set karma of %s in database %s: %s", phrase, database, err)
	}

	// Каждые -5 пунктов пробивают дно :)
	if karmaInt < -5 && math.Mod(float64(karmaInt), 5) == -1 {
		answer = "Зарегистрировано пробитие дна: к"
	} else {
		answer = "K"
	}

	if phrase == " " {
		return fmt.Sprintf("%sарма пустоты равна %s", answer, karmaString)
	} else {
		return fmt.Sprintf("%sарма %s равна %s", answer, phrase, karmaString)
	}
}

// Достанем карму фразы
func getKarma(chatID string, phrase string) string {
	var err error

	chatHash := sha256.Sum256([]byte(chatID))
	database := fmt.Sprintf("karma_db/%x", chatHash)

	if phrase = strings.Trim(phrase, "\n\r\t "); phrase == "" {
		phrase = " "
	}

	if _, ok := KarmaDB[database]; !ok {
		var options pebble.Options
		// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными
		// настройками порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук.
		options.L0CompactionFileThreshold = 300
		KarmaDB[database], err = pebble.Open(Config.DataDir+"/"+database, &options)

		if err != nil {
			log.Errorf("Unable to open karma db, %s: %s\n", database, err)

			if phrase == " " {
				return "Карма пустоты равна 0"
			} else {
				return fmt.Sprintf("Карма %s равна 0", phrase)
			}
		}
	}

	answer, err := FetchV(KarmaDB[database], phrase)

	// Если из базы ничего не вынулось, по каким-то причинам, то просто соврём, что карма равна нулю
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			log.Debugf("Unable to get karma for %s: no record found in db %s", phrase, database)
		} else if errors.Is(err, fs.ErrNotExist) {
			log.Debugf("Unable to get karma for %s: db dir %s does not exist", phrase, database)
		} else if errors.Is(err, oserror.ErrNotExist) {
			log.Debugf("Unable to get karma for %s: db dir %s does not exist", phrase, database)
		} else {
			log.Errorf("Unable to get karma for %s in db dir %s: %s", phrase, database, err)
		}

		if phrase == " " {
			return "Карма пустоты равна 0"
		} else {
			return fmt.Sprintf("Карма %s равна 0", phrase)
		}
	}

	if phrase == " " {
		return fmt.Sprintf("Карма пустоты равна %s", answer)
	} else {
		return fmt.Sprintf("Карма %s равна %s", phrase, answer)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
