package env

import (
	"os"
	"shorturl/redis"
	"shorturl/services"
	"strconv"
)

type Env struct {
	S services.Storage
}

func GetEnv() *Env {
	addr := os.Getenv("APP_REDIS_ARRD")
	if addr == "" {
		addr = "192.168.63.10:6379"
	}

	passwd := os.Getenv("APP_REDIS_PASSWD")
	if passwd == "" {
		passwd = "secret_redis"
	}
	dbs := os.Getenv("APP_REDIS_DB")
	if dbs == "" {
		dbs = "0"
	}

	db, _ := strconv.Atoi(dbs)

	c := redis.NewRedisCli(addr, passwd, db)
	return &Env{S: c}
}
