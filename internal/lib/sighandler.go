package lib

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Хэндлер сигналов закрывает все бд и сваливает из приложения
func SigHandler() {
	var err error

	for {
		var s = <-SigChan
		switch s {
		case syscall.SIGINT:
			log.Infoln("Got SIGINT, quitting")
		case syscall.SIGTERM:
			log.Infoln("Got SIGTERM, quitting")
		case syscall.SIGQUIT:
			log.Infoln("Got SIGQUIT, quitting")

		// Заходим на новую итерацию, если у нас "неинтересный" сигнал
		default:
			continue
		}

		// Чтобы не срать в логи ошибками от редиски, проставим shutdown state приложения в true
		Shutdown = true

		// Отпишемся от всех каналов и закроем коннект к редиске
		if err = Subscriber.Unsubscribe(Ctx); err != nil {
			log.Errorf("Unable to unsubscribe from redis channels cleanly: %s", err)
		}

		if err = Subscriber.Close(); err != nil {
			log.Errorf("Unable to close redis connection cleanly: %s", err)
		}

		if err := PebbleDB.Fortune.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Fortune.Name, err)
		}

		// Закроем статичные базы с фразами, тут их многовато :(
		if err = PebbleDB.Fortune.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Fortune.Name, err)
		}

		if err := PebbleDB.Proverb.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Proverb.Name, err)
		}

		if err = PebbleDB.Proverb.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Proverb.Name, err)
		}

		if err := PebbleDB.Count.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Count.Name, err)
		}

		if err = PebbleDB.Count.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Count.Name, err)
		}

		if err := PebbleDB.Monday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Monday.Name, err)
		}

		if err = PebbleDB.Monday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Monday.Name, err)
		}

		if err := PebbleDB.Tuesday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Tuesday.Name, err)
		}

		if err = PebbleDB.Tuesday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Tuesday.Name, err)
		}

		if err := PebbleDB.Wednesday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Wednesday.Name, err)
		}

		if err = PebbleDB.Wednesday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Wednesday.Name, err)
		}

		if err := PebbleDB.Thursday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Thursday.Name, err)
		}

		if err = PebbleDB.Thursday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Thursday.Name, err)
		}

		if err := PebbleDB.Friday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Friday.Name, err)
		}

		if err = PebbleDB.Friday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Friday.Name, err)
		}

		if err := PebbleDB.Saturday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Saturday.Name, err)
		}

		if err = PebbleDB.Saturday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Saturday.Name, err)
		}

		if err := PebbleDB.Sunday.DB.Flush(); err != nil {
			log.Errorf("Unable to flush %s db: %s", PebbleDB.Sunday.Name, err)
		}

		if err = PebbleDB.Sunday.DB.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", PebbleDB.Sunday.Name, err)
		}

		// Закроем базы с кармой
		for dbName, karma := range KarmaDB {
			if err := karma.Flush(); err != nil {
				log.Errorf("Unable to flush %s karma db: %s", dbName, err)
			}

			if err := karma.Close(); err != nil {
				log.Errorf("Unable to close %s karma db: %s", dbName, err)
			}
		}

		os.Exit(0)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
