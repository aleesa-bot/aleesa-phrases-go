package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// Входящее сообщение из pubsub-канала redis-ки
type rMsg struct {
	From     string `json:"from,omitempty"`
	ChatID   string `json:"chatid,omitempty"`
	UserID   string `json:"userid,omitempty"`
	ThreadID string `json:"threadid"`
	Message  string `json:"message,omitempty"`
	Plugin   string `json:"plugin,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Misc     struct {
		Answer      int64  `json:"answer,omitempty"`
		BotNick     string `json:"bot_nick,omitempty"`
		CSign       string `json:"csign,omitempty"`
		FwdCnt      int64  `json:"fwd_cnt,omitempty"`
		GoodMorning int64  `json:"good_morning,omitempty"`
		MsgFormat   int64  `json:"msg_format,omitempty"`
		Username    string `json:"username,omitempty"`
	} `json:"Misc"`
}

// Исходящее сообщение в pubsub-канал redis-ки
type sMsg struct {
	From     string `json:"from"`
	ChatID   string `json:"chatid"`
	Userid   string `json:"userid"`
	ThreadID string `json:"threadid"`
	Message  string `json:"message"`
	Plugin   string `json:"plugin"`
	Mode     string `json:"mode"`
	Misc     struct {
		Answer      int64  `json:"answer"`
		BotNick     string `json:"bot_nick"`
		CSign       string `json:"csign"`
		FwdCnt      int64  `json:"fwd_cnt"`
		GoodMorning int64  `json:"good_morning"`
		MsgFormat   int64  `json:"msg_format"`
		Username    string `json:"username"`
	} `json:"misc"`
}

// Дефолтный конфиг клиента-редиски
var redisClient *redis.Client
var subscriber *redis.PubSub

// Мапка с открытыми дескрипторами баз с кармой
var karmaDB = make(map[string]*pebble.DB)

// Канал, в который приходят уведомления для хэндлера сигналов от траппера сигналов
var sigChan = make(chan os.Signal, 1)

// Main context
var ctx = context.Background()

// Ставится в true, если мы получили сигнал на выключение
var shutdown = false

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

	if _, ok := karmaDB[database]; !ok {
		var options pebble.Options
		// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными
		// настройками порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук,
		// что нас не устраивает, поэтому немного снизим эту цифру до более приемлемых значений
		options.L0CompactionFileThreshold = 8

		karmaDB[database], err = pebble.Open(config.DataDir+"/"+database, &options)

		if err != nil {
			log.Errorf("Unable to open karma db, %s: %s\n", database, err)

			if phrase == " " {
				return "Карма пустоты равна 0"
			} else {
				return fmt.Sprintf("Карма %s равна 0", phrase)
			}
		}
	}

	karmaString, err := FetchV(karmaDB[database], phrase)

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

	if err := StoreKV(karmaDB[database], "", phrase, karmaString, false); err != nil {
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

	if _, ok := karmaDB[database]; !ok {
		var options pebble.Options
		// По дефолту ограничение ставится на мегабайты данных, а не на количество файлов, поэтому с дефолтными
		// настройками порождается огромное количество файлов. Умолчальное ограничение на количество файлов - 500 штук,
		// что нас не устраивает, поэтому немного снизим эту цифру до более приемлемых значений
		options.L0CompactionFileThreshold = 8
		karmaDB[database], err = pebble.Open(config.DataDir+"/"+database, &options)

		if err != nil {
			log.Errorf("Unable to open karma db, %s: %s\n", database, err)

			if phrase == " " {
				return "Карма пустоты равна 0"
			} else {
				return fmt.Sprintf("Карма %s равна 0", phrase)
			}
		}
	}

	answer, err := FetchV(karmaDB[database], phrase)

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

// А не пятница ли сейчас?
func friday() string {
	var answer = fmt.Sprintf("Today is %s", time.Now().Weekday().String())
	var phrase string
	var err error

	// Но нам-то интересно какое-то более доругое высказывание относительно текущего дня недели...
	dayNum := int(time.Now().Weekday())

	switch dayNum {
	case 0:
		phrase, err = FetchRandomV(pebbleDB.sunday.db, pebbleDB.sunday.name)
	case 1:
		phrase, err = FetchRandomV(pebbleDB.monday.db, pebbleDB.monday.name)
	case 2:
		phrase, err = FetchRandomV(pebbleDB.tuesday.db, pebbleDB.tuesday.name)
	case 3:
		phrase, err = FetchRandomV(pebbleDB.wednesday.db, pebbleDB.wednesday.name)
	case 4:
		phrase, err = FetchRandomV(pebbleDB.thursday.db, pebbleDB.thursday.name)
	case 5:
		phrase, err = FetchRandomV(pebbleDB.friday.db, pebbleDB.friday.name)
	case 6:
		phrase, err = FetchRandomV(pebbleDB.saturday.db, pebbleDB.saturday.name)
	}

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

// Пословица
func proverb() string {
	var answer = "There is no knowledge that is not power."
	phrase, err := FetchRandomV(pebbleDB.proverb.db, pebbleDB.proverb.name)

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

// Фортунка
func fortune() string {
	var answer = "There is no knowledge that is not power."
	phrase, err := FetchRandomV(pebbleDB.fortune.db, pebbleDB.fortune.name)

	if err != nil {
		log.Errorln(err)
		return answer
	}

	return phrase
}

// Горутинка, которая парсит json-чики прилетевшие из REDIS-ки
func msgParser(ctx context.Context, msg string) {
	var sendTo string
	var j rMsg
	var message sMsg

	log.Debugf("Incomming raw json: %s", msg)

	if err := json.Unmarshal([]byte(msg), &j); err != nil {
		log.Warnf("Unable to to parse message from redis channel: %s", err)
		return
	}

	// Validate our j
	if exist := j.From; exist == "" {
		log.Warnf("Incorrect msg from redis, no from field: %s", msg)
		return
	}

	if exist := j.ChatID; exist == "" {
		log.Warnf("Incorrect msg from redis, no chatid field: %s", msg)
		return
	}

	if exist := j.UserID; exist == "" {
		log.Warnf("Incorrect msg from redis, no userid field: %s", msg)
		return
	}

	// j.ThreadID может быть пустым, это нормально

	if exist := j.Message; exist == "" {
		log.Warnf("Incorrect msg from redis, no message field: %s", msg)
		return
	}

	if exist := j.Plugin; exist == "" {
		log.Warnf("Incorrect msg from redis, no plugin field: %s", msg)
		return
	} else {
		sendTo = j.Plugin
	}

	if exist := j.Mode; exist == "" {
		log.Warnf("Incorrect msg from redis, no mode field: %s", msg)
		return
	}

	// j.Misc.Answer может и не быть, тогда ответа на такое сообщение не будет
	if j.Misc.Answer == 0 {
		log.Debug("Field Misc->Answer = 0, skipping message")
		return
	}

	// j.Misc.BotNick тоже можно не передавать, тогда будет записана пустая строка
	// j.Misc.CSign если нам его не передали, возьмём значение из конфига
	if exist := j.Misc.CSign; exist == "" {
		j.Misc.CSign = config.CSign
	}

	// j.Misc.FwdCnt если нам его не передали, то будет 0
	if exist := j.Misc.FwdCnt; exist == 0 {
		j.Misc.FwdCnt = 1
	}

	// j.Misc.GoodMorning может быть быть 1 или 0, по-умолчанию 0
	// j.Misc.MsgFormat может быть быть 1 или 0, по-умолчанию 0
	// j.Misc.Username можно не передавать, тогда будет пустая строка

	// Отвалидировались, теперь вернёмся к нашим баранам.

	// Если у нас циклическая пересылка сообщения, попробуем её тут разорвать, отбросив сообщение
	if j.Misc.FwdCnt > config.ForwardsMax {
		log.Warnf("Discarding msg with fwd_cnt exceeding max value: %s", msg)
		return
	} else {
		j.Misc.FwdCnt++
	}

	// Классифицирем входящие сообщения. Первым делом, попробуем определить команды
	msgLen := len(j.Message)

	if j.Message[0:len(j.Misc.CSign)] == j.Misc.CSign { //nolint:gocritic
		cmd := j.Message[len(j.Misc.CSign):]

		if cmd == "friday" || cmd == "пятница" {
			j.Message = friday()
		} else if (cmd == "proverb") || (cmd == "пословица") || (cmd == "пословиться") {
			j.Message = proverb()
		} else if (cmd == "f") || (cmd == "ф") {
			j.Message = fortune()
		} else if (cmd == "fortune") || (cmd == "фортунка") {
			j.Message = fortune()
		} else if (cmd == "karma") || (cmd == "карма") {
			j.Message = getKarma(j.ChatID, "")
		} else if match, _ := regexp.MatchString("^(rum|ром)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me притаскивает на подносе стопку рома для %s, края стопки искрятся кристаллами соли."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(vodka|водка)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me подаёт шот водки с небольшим маринованным огурчиком на блюдце для %s. Из огурчика торчит "
			format += "небольшая вилочка."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(beer|пиво)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me бахает об стол перед %s кружкой холодного пива, часть пенной шапки сползает по запотевшей "
			format += "стенке кружки."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(tequila|текила)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me ставит рядом с %s шот текилы, аккуратно на ребро стопки насаживает дольку лайма и ставит "
			format += "кофейное блюдце с горочкой соли."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(whisky|виски)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me демонстративно достаёт из морозилки пару кубических камушков, бросает их в толстодонный "
			format += "стакан и аккуратно наливает Jack Daniels. Запускает стакан вдоль барной стойки, он "
			format += "останавливается около %s."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(absinth|абсент)[::space::]?([::space::].+)*$", cmd); match {
			format := "/me наливает абсент в стопку. Смочив кубик сахара в абсенте кладёт его на дырявую ложечку и "
			format += "поджигает. Как только пламя потухнет, %s размешивает оплавившийся кубик в абсенте и подносит "
			format += "стопку %s."
			j.Message = fmt.Sprintf(format, j.Misc.BotNick, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^fuck[::space::]*$", cmd); match {
			j.Message = "Не ругайся."
		} else {
			cmdLen := len(cmd)
			cmds := []string{"karma ", "карма "}

			for _, command := range cmds {
				if cmdLen > len(command) && cmd[0:len(command)] == command {
					j.Message = getKarma(j.ChatID, cmd[len(command):])
					break
				}
			}
		}

		// Это последняя проверка, поэтому во всех else мы исходящее сообщение можем смело ставить в пустую строку
	} else if msgLen > len("++") {
		if j.Message[msgLen-len("--"):msgLen] == "--" || j.Message[msgLen-len("++"):msgLen] == "++" {

			// Предполагается, что менять карму мы будем для одной фразы, то есть для 1 строки
			if len(strings.Split(j.Message, "\n")) == 1 {
				if j.Message[msgLen-len("--"):msgLen] == "--" {
					j.Message = changeKarma(j.ChatID, j.Message[0:msgLen-len("--")], "--")
				} else if j.Message[msgLen-len("++"):msgLen] == "++" {
					j.Message = changeKarma(j.ChatID, j.Message[0:msgLen-len("++")], "++")
				} else {
					j.Message = ""
				}
			} else {
				j.Message = ""
			}
		} else {
			j.Message = ""
		}
	} else {
		j.Message = ""
	}

	if j.Message != "" {
		// Настало время формировать json и засылать его в дальше
		message.From = j.From
		message.Userid = j.UserID
		message.ChatID = j.ChatID
		message.ThreadID = j.ThreadID
		message.Message = j.Message
		message.Plugin = j.Plugin
		message.Mode = j.Mode
		message.Misc.FwdCnt = j.Misc.FwdCnt
		message.Misc.CSign = j.Misc.CSign
		message.Misc.Username = j.Misc.Username
		message.Misc.Answer = j.Misc.Answer
		message.Misc.BotNick = j.Misc.BotNick
		message.Misc.MsgFormat = j.Misc.MsgFormat

		data, err := json.Marshal(message)

		if err != nil {
			log.Warnf("Unable to to serialize message for redis: %s", err)
			return
		}

		// Заталкиваем наш json в редиску
		if err := redisClient.Publish(ctx, sendTo, data).Err(); err != nil {
			log.Warnf("Unable to send data to redis channel %s: %s", sendTo, err)
		} else {
			log.Debugf("Send msg to redis channel %s: %s", sendTo, string(data))
		}
	}
}

// Хэндлер сигналов закрывает все бд и сваливает из приложения
func sigHandler() {
	var err error

	for {
		var s = <-sigChan
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
		shutdown = true

		// Отпишемся от всех каналов и закроем коннект к редиске
		if err = subscriber.Unsubscribe(ctx); err != nil {
			log.Errorf("Unable to unsubscribe from redis channels cleanly: %s", err)
		}

		if err = subscriber.Close(); err != nil {
			log.Errorf("Unable to close redis connection cleanly: %s", err)
		}

		// Закроем статичные базы с фразами, тут их многовато :(
		if err = pebbleDB.fortune.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.fortune.name, err)
		}

		if err = pebbleDB.proverb.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.proverb.name, err)
		}

		if err = pebbleDB.count.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.count.name, err)
		}

		if err = pebbleDB.monday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.monday.name, err)
		}

		if err = pebbleDB.tuesday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.tuesday.name, err)
		}

		if err = pebbleDB.wednesday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.wednesday.name, err)
		}

		if err = pebbleDB.thursday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.thursday.name, err)
		}

		if err = pebbleDB.friday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.friday.name, err)
		}

		if err = pebbleDB.saturday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.saturday.name, err)
		}

		if err = pebbleDB.sunday.db.Close(); err != nil {
			log.Errorf("Unable to close %s db: %s", pebbleDB.sunday.name, err)
		}

		// Закроем базы с кармой
		for dbName, karma := range karmaDB {
			if err := karma.Close(); err != nil {
				log.Errorf("Unable to close %s karma db: %s", dbName, err)
			}
		}

		os.Exit(0)
	}
}

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
	random = rand.New(rngSeed)

	config = ReadConfig()

	// no panic, no trace
	switch config.LogLevel {
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

	pebbleDB.count.name = "count_db"
	pebbleDB.count.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.count.name, &options)
	if err != nil {
		log.Errorf("Unable to open count_db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.fortune.name = "fortune_db"
	pebbleDB.fortune.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.fortune.name, &options)
	if err != nil {
		log.Errorf("Unable to open fortune db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.proverb.name = "proverb_db"
	pebbleDB.proverb.db, err = pebble.Open(config.DataDir+"/"+pebbleDB.proverb.name, &options)
	if err != nil {
		log.Errorf("Unable to open proverb db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.monday.name = "monday"
	pebbleDB.monday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.monday.name, &options)
	if err != nil {
		log.Errorf("Unable to open monday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.tuesday.name = "tuesday"
	pebbleDB.tuesday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.tuesday.name, &options)
	if err != nil {
		log.Errorf("Unable to open tuesday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.wednesday.name = "wednesday"
	pebbleDB.wednesday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.wednesday.name, &options)
	if err != nil {
		log.Errorf("Unable to open wednesday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.thursday.name = "thursday"
	pebbleDB.thursday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.thursday.name, &options)
	if err != nil {
		log.Errorf("Unable to open thursday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.friday.name = "friday"
	pebbleDB.friday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.friday.name, &options)
	if err != nil {
		log.Errorf("Unable to open friday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.saturday.name = "saturday"
	pebbleDB.saturday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.saturday.name, &options)
	if err != nil {
		log.Errorf("Unable to open saturday db: %s\n", err)
		os.Exit(1)
	}

	pebbleDB.sunday.name = "sunday"
	pebbleDB.sunday.db, err = pebble.Open(config.DataDir+"/friday_db/"+pebbleDB.sunday.name, &options)
	if err != nil {
		log.Errorf("Unable to open sunday db: %s\n", err)
		os.Exit(1)
	}
}

// Основная функция программы, не добавить и не убавить
func main() {

	// Откроем лог и скормим его логгеру
	if config.Log != "" {
		logfile, err := os.OpenFile(config.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Unable to open log file %s: %s", config.Log, err)
		}

		log.SetOutput(logfile)
	}

	// Иницализируем клиента
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Server, config.Port),
	})

	log.Debugf("Lazy connect() to redis at %s:%d", config.Server, config.Port)
	subscriber = redisClient.Subscribe(ctx, config.Channel)

	// Самое время поставить траппер сигналов
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go sigHandler()

	// Обработчик событий от редиски
	for {
		if shutdown {
			time.Sleep(1 * time.Second)
			continue
		}

		msg, err := subscriber.ReceiveMessage(ctx)

		if err != nil {
			if !shutdown {
				log.Warnf("Unable to connect to redis at %s:%d: %s", config.Server, config.Port, err)
			}

			time.Sleep(1 * time.Second)
			continue
		}

		go msgParser(ctx, msg.Payload)
	}
}
