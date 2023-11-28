package storage

import (
	"context"
	"errors"
	"fmt"
	"order-processing/statics"

	redisLib "github.com/redis/go-redis/v9"
)

type ZInterOptions struct {
	prefix  string
	keys    []string
	weights []float64
}

type LimitOptions struct {
	minPrice float64 ``
	maxPrice float64
}

type RedisClient struct {
	client redisLib.Client
}

func GetNewRedisCli(address string) RedisClient {
	client := redisLib.NewClient(&redisLib.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return RedisClient{client: *client}
}

func (c *RedisClient) InsertHash(ctx context.Context, hashName string, key string, value []byte) error {
	state := c.client.HSet(ctx, hashName, key, value)
	if state.Err() != nil {
		return errors.New(statics.InternalError)
	}
	return nil
}

func (c *RedisClient) InsertZadd(ctx context.Context, setName string, key string, weight float64) error {
	state := c.client.ZAdd(ctx, setName, redisLib.Z{Member: key, Score: weight})
	if state.Err() != nil {
		return errors.New(statics.InternalError)
	}
	return nil
}

func (c *RedisClient) InsertSet(ctx context.Context, setName string, value string) error {
	state := c.client.SAdd(ctx, setName, value, -1)
	if state.Err() != nil {
		return errors.New(statics.InternalError)
	}
	return nil
}

func (c *RedisClient) GetFromHash(ctx context.Context, hashName string, value string) (*string, error) {
	info, err := c.client.HGet(ctx, hashName, value).Result()
	if err == redisLib.Nil {
		return nil, errors.New(statics.ErrorOrderNotFound)
	}
	if err != nil {
		return nil, errors.New(statics.InternalError)
	}
	return &info, nil
}

func (c *RedisClient) GetFromSet(ctx context.Context, setName string) ([]string, error) {
	return c.client.SMembers(ctx, setName).Result()
}

func (c *RedisClient) DeleteFromHash(ctx context.Context, key string, field string) {
	c.client.HDel(ctx, key, field)
}

func (c *RedisClient) DeleteFromSet(ctx context.Context, setName string, key string) {
	c.client.SRem(ctx, setName, key)
}

func (c *RedisClient) DeleteFromZAdd(ctx context.Context, setName string, key string) {
	c.client.ZRem(ctx, setName, key)
}

func (c *RedisClient) ZInterStorage(ctx context.Context, config ZInterOptions, orderId string) string {
	c.client.ZInterStore(ctx, config.prefix+orderId, &redisLib.ZStore{
		Keys:    config.keys,
		Weights: config.weights,
	})
	return config.prefix + orderId
}

func (c *RedisClient) ZRange(ctx context.Context, setName string, limit int64) ([]string, error) {
	ids, err := c.client.ZRange(ctx, setName, 0, limit).Result()
	if err == redisLib.Nil {
		return nil, errors.New(statics.ErrorOrderNotFound)
	}
	if err != nil {
		return nil, errors.New(statics.InternalError)
	}
	return ids, nil
}

func (c *RedisClient) PrepareIndexWithLimitOption(ctx context.Context, options LimitOptions) (*string, error) {
	indexName := "orders:limit_price:" + fmt.Sprintf("%f", options.minPrice) + fmt.Sprintf("%f", options.maxPrice)
	c.client.ZInterStore(ctx, indexName, &redisLib.ZStore{
		Keys: []string{OrdersPrice},
	})
	c.client.ZRemRangeByScore(ctx, indexName, "-inf", fmt.Sprintf("%f", options.minPrice))
	if options.maxPrice > 0 {
		c.client.ZRemRangeByScore(ctx, indexName, fmt.Sprintf("%f", options.maxPrice), "+inf")
	}
	return &indexName, nil
}
