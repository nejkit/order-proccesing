package storage

import (
	redisLib "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	logger *logrus.Logger
	client redisLib.Client
}

func GetNewRedisCli(logger *logrus.Logger, connectionString string) RedisClient {
	client := redisLib.NewClient(&redisLib.Options{
		Addr:     connectionString,
		Password: "",
		DB:       0,
	})
	return RedisClient{logger: logger, client: *client}

}
