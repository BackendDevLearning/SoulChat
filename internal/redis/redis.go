package redis

import (
	"context"
	"time"
)

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.Get(ctx, key).Result()
}

func SSet(ctx context.Context, key string, members []interface{}, expiration time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.SAdd(ctx, key, members...).Err()
}

func SGet(ctx context.Context, key string) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.SMembers(ctx, key).Result()
}

func HSet(ctx context.Context, key string, field string, value interface{}, expiration time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.HSet(ctx, key, field, value).Err()
}

func HGet(ctx context.Context, key string, field string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return redisClient.HGet(ctx, key, field).Result()
}
