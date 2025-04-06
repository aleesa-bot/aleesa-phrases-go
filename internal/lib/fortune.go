package lib

import (
	log "github.com/sirupsen/logrus"
)

// Фортунка
func fortune() string {
	var answer = "There is no knowledge that is not power."
	phrase, err := FetchRandomV(PebbleDB.Fortune.DB, PebbleDB.Fortune.Name)

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
