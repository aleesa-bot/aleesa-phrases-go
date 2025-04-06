package lib

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// А не пятница ли сейчас?
func friday() string {
	var answer = fmt.Sprintf("Today is %s", time.Now().Weekday().String())
	var phrase string
	var err error

	// Но нам-то интересно какое-то более доругое высказывание относительно текущего дня недели...
	dayNum := int(time.Now().Weekday())

	switch dayNum {
	case 0:
		phrase, err = FetchRandomV(PebbleDB.Sunday.DB, PebbleDB.Sunday.Name)
	case 1:
		phrase, err = FetchRandomV(PebbleDB.Monday.DB, PebbleDB.Monday.Name)
	case 2:
		phrase, err = FetchRandomV(PebbleDB.Tuesday.DB, PebbleDB.Tuesday.Name)
	case 3:
		phrase, err = FetchRandomV(PebbleDB.Wednesday.DB, PebbleDB.Wednesday.Name)
	case 4:
		phrase, err = FetchRandomV(PebbleDB.Thursday.DB, PebbleDB.Thursday.Name)
	case 5:
		phrase, err = FetchRandomV(PebbleDB.Friday.DB, PebbleDB.Friday.Name)
	case 6:
		phrase, err = FetchRandomV(PebbleDB.Saturday.DB, PebbleDB.Saturday.Name)
	}

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
