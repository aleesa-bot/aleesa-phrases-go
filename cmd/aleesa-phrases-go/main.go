package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aleesa-phrases-go/internal/lib"

	"github.com/cockroachdb/pebble"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// Производит некоторую инициализацию перед запуском main()
func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:           true,
		DisableLevelTruncation: false,
		DisableColors:          true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
	})

	// Запустим наш рандомайзер
	rngSeed := rand.NewSource(time.Now().UnixNano())
	lib.Random = rand.New(rngSeed)

	lib.Config = lib.ReadConfig()

	// no panic, no trace
	switch lib.Config.LogLevel {
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Надеюсь, это заловит все ошибки и паники в лог при падении программы
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Exception: %v\n", err)
			os.Exit(1)
		}
	}()

	var err error
	// Самое гнусное - открыть базки со статичными фразами, конкретно тут - в RO-режиме
	var options pebble.Options
	options.ReadOnly = true

	lib.PebbleDB.Count.Name = "count_db"
	lib.PebbleDB.Count.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Count.Name, &options)
	if err != nil {
		log.Errorf("Unable to open count_db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Fortune.Name = "fortune_db"
	lib.PebbleDB.Fortune.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Fortune.Name, &options)
	if err != nil {
		log.Errorf("Unable to open fortune db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Proverb.Name = "proverb_db"
	lib.PebbleDB.Proverb.DB, err = pebble.Open(lib.Config.DataDir+"/"+lib.PebbleDB.Proverb.Name, &options)
	if err != nil {
		log.Errorf("Unable to open proverb db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Monday.Name = "monday"
	lib.PebbleDB.Monday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Monday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open monday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Tuesday.Name = "tuesday"
	lib.PebbleDB.Tuesday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Tuesday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open tuesday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Wednesday.Name = "wednesday"
	lib.PebbleDB.Wednesday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Wednesday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open wednesday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Thursday.Name = "thursday"
	lib.PebbleDB.Thursday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Thursday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open thursday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Friday.Name = "friday"
	lib.PebbleDB.Friday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Friday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open friday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Saturday.Name = "saturday"
	lib.PebbleDB.Saturday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Saturday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open saturday db: %s\n", err)
		os.Exit(1)
	}

	lib.PebbleDB.Sunday.Name = "sunday"
	lib.PebbleDB.Sunday.DB, err = pebble.Open(lib.Config.DataDir+"/friday_db/"+lib.PebbleDB.Sunday.Name, &options)
	if err != nil {
		log.Errorf("Unable to open sunday db: %s\n", err)
		os.Exit(1)
	}
}

// Основная функция программы, не добавить и не убавить
func main() {

	// Откроем лог и скормим его логгеру
	if lib.Config.Log != "" {
		logfile, err := os.OpenFile(lib.Config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Unable to open log file %s: %s", lib.Config.Log, err)
		}

		log.SetOutput(logfile)
	}

	// Иницализируем клиента
	lib.RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", lib.Config.Server, lib.Config.Port),
	})

	log.Debugf("Lazy connect() to redis at %s:%d", lib.Config.Server, lib.Config.Port)
	lib.Subscriber = lib.RedisClient.Subscribe(lib.Ctx, lib.Config.Channel)

	// Самое время поставить траппер сигналов
	signal.Notify(lib.SigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go lib.SigHandler()

	// Обработчик событий от редиски
	for {
		if lib.Shutdown {
			time.Sleep(1 * time.Second)
			continue
		}

		msg, err := lib.Subscriber.ReceiveMessage(lib.Ctx)

		if err != nil {
			if !lib.Shutdown {
				log.Warnf("Unable to connect to redis at %s:%d: %s", lib.Config.Server, lib.Config.Port, err)
			}

			time.Sleep(1 * time.Second)
			continue
		}

		go lib.MsgParser(lib.Ctx, msg.Payload)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
