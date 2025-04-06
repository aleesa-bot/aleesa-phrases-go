package lib

import (
	"context"
	"math/rand"
	"os"

	"github.com/cockroachdb/pebble"
	"github.com/go-redis/redis/v8"
)

// To break circular message forwarding we must set some sane default, it can be overridden via config
var ForwardMax int64 = 5

// Config - это у нас глобальная штука :)
var Config MyConfig

// Просто счётчик, использован в seed-программах
var Counter = 0

var Dow = []string{
	"monday",
	"tuesday",
	"wednesday",
	"thursday",
	"friday",
	"saturday",
	"sunday",
}

// Struct с дескрипторами БД
var PebbleDB struct {
	// Счётчик записей для каждой БД
	Count struct {
		DB   *pebble.DB
		Name string
	}
	// База модуля пословиц
	Proverb struct {
		DB   *pebble.DB
		Name string
	}
	// База модуля fortune
	Fortune struct {
		DB   *pebble.DB
		Name string
	}
	// 7 баз по дням недели модуля friday
	Monday struct {
		DB   *pebble.DB
		Name string
	}
	Tuesday struct {
		DB   *pebble.DB
		Name string
	}
	Wednesday struct {
		DB   *pebble.DB
		Name string
	}
	Thursday struct {
		DB   *pebble.DB
		Name string
	}
	Friday struct {
		DB   *pebble.DB
		Name string
	}
	Saturday struct {
		DB   *pebble.DB
		Name string
	}
	Sunday struct {
		DB   *pebble.DB
		Name string
	}
}

// Мапка с открытыми дескрипторами баз с кармой
var KarmaDB = make(map[string]*pebble.DB)

// Дефолтный конфиг клиента-редиски
var RedisClient *redis.Client
var Subscriber *redis.PubSub

// Создадим источник рандома
var Random *rand.Rand

// Канал, в который приходят уведомления для хэндлера сигналов от траппера сигналов
var SigChan = make(chan os.Signal, 1)

// Ставится в true, если мы получили сигнал на выключение
var Shutdown = false

// Main context
var Ctx = context.Background()

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
