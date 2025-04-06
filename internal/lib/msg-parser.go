package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Горутинка, которая парсит json-чики прилетевшие из REDIS-ки
func MsgParser(ctx context.Context, msg string) {
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
		j.Misc.CSign = Config.CSign
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
	if j.Misc.FwdCnt > Config.ForwardsMax {
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
		} else if match, _ := regexp.MatchString("^(rum|ром)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me притаскивает на подносе стопку рома для %s, края стопки искрятся кристаллами соли."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(vodka|водка)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me подаёт шот водки с небольшим маринованным огурчиком на блюдце для %s. Из огурчика торчит "
			format += "небольшая вилочка."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(beer|пиво)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me бахает об стол перед %s кружкой холодного пива, часть пенной шапки сползает по запотевшей "
			format += "стенке кружки."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(tequila|текила)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me ставит рядом с %s шот текилы, аккуратно на ребро стопки насаживает дольку лайма и ставит "
			format += "кофейное блюдце с горочкой соли."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(whisky|виски)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me демонстративно достаёт из морозилки пару кубических камушков, бросает их в толстодонный "
			format += "стакан и аккуратно наливает Jack Daniels. Запускает стакан вдоль барной стойки, он "
			format += "останавливается около %s."
			j.Message = fmt.Sprintf(format, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^(absinth|абсент)[[:space:]]?([[:space:]].+)*$", cmd); match {
			format := "/me наливает абсент в стопку. Смочив кубик сахара в абсенте кладёт его на дырявую ложечку и "
			format += "поджигает. Как только пламя потухнет, %s размешивает оплавившийся кубик в абсенте и подносит "
			format += "стопку %s."
			j.Message = fmt.Sprintf(format, j.Misc.BotNick, j.Misc.Username)
		} else if match, _ := regexp.MatchString("^fuck[[:space:]]*$", cmd); match {
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
		if err := RedisClient.Publish(ctx, sendTo, data).Err(); err != nil {
			log.Warnf("Unable to send data to redis channel %s: %s", sendTo, err)
		} else {
			log.Debugf("Send msg to redis channel %s: %s", sendTo, string(data))
		}
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
