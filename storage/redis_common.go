package storage

import (
	"context"

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

func (cli *RedisClient) Insert(ctx context.Context, key string, value []byte) error {
	state := cli.client.Set(ctx, key, value, 0)
	if state.Err() != nil {
		return state.Err()
	}
	return nil
}

func (cli *RedisClient) GetByPattern(ctx context.Context, pattern string) ([]string, error) {
	return cli.client.Keys(ctx, pattern).Result()
}

func (cli *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	state := cli.client.Del(ctx, pattern)
	if state.Err() != nil {
		return state.Err()
	}
	return nil
}
