package cache

import (
	"context"
	"time"

	. "godopi/internal/pkg/logger"

	"github.com/go-redis/redis/v8"
)

const CacheNil = redis.Nil

type CacheClient interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, interface{}, time.Duration) error
	Del(context.Context, ...string) error
}

type cacheClient struct {
	client *redis.Client
}

func NewCacheClient(address string) CacheClient {
	Logger().Info("Constructing new cache client..")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return cacheClient{client: redisClient}
}

func (cc cacheClient) Get(ctx context.Context, key string) (string, error) {
	return cc.client.Get(ctx, key).Result()
}

func (cc cacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cc.client.Set(ctx, key, value, expiration).Err()
}

func (cc cacheClient) Del(ctx context.Context, keys ...string) error {
	return cc.client.Del(ctx, keys...).Err()
}
