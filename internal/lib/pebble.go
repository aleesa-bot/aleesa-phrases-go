package lib

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"

	"github.com/cockroachdb/pebble"
)

// StoreKV сохраняет в указанной бд ключ и значение
func StoreKV(db *pebble.DB, dbName string, key string, value string, useCounter bool) error {
	var kArray = []byte(key)
	var vArray = []byte(value)

	err := db.Set(kArray, vArray, pebble.Sync)

	if useCounter {
		// У pebble нет в арсенале функции count(), так что "изобретём" её
		if err == nil {
			amount, err := FetchV(PebbleDB.Count.DB, dbName)

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

			err = PebbleDB.Count.DB.Set([]byte(dbName), []byte(amount), pebble.Sync)

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
	amount, err := FetchV(PebbleDB.Count.DB, dbName)

	if err != nil {
		// Такого ключа в базе нет
		if errors.Is(err, pebble.ErrNotFound) {
			return answer, fmt.Errorf("looks like %s is not seeded or count_db is missing", dbName)
			// Такой базы нет
		} else if errors.Is(err, fs.ErrNotExist) {
			return answer, fmt.Errorf("unable to get random value from %s: count_db is missing", dbName)
		} else {
			return answer, fmt.Errorf("something going wrong with %s: %w", dbName, err)
		}
	}

	amountNum, err := strconv.Atoi(amount)

	// Хуйня в базе, данные должны проходить конвертацию в int
	if err != nil {
		return answer, fmt.Errorf("garbage in count_db while getting info about %s: %w", dbName, err)
	}

	phraseNum := Random.Intn(amountNum + 1)
	phraseItem := fmt.Sprintf("%d", phraseNum)

	answer, err = FetchV(db, phraseItem)

	if err != nil {
		return "", fmt.Errorf("an error occured while getting answer from %s: %s", dbName, err)
	}

	return answer, nil
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
