package lib

import (
	log "github.com/sirupsen/logrus"
)

// Пословица
func proverb() string {
	var answer = "There is no knowledge that is not power."
	phrase, err := FetchRandomV(PebbleDB.Proverb.DB, PebbleDB.Proverb.Name)

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
