package redis

import (
	"context"
	"time"

	"os"

	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

var redisClient *redis.Client

type RedisConfig struct {
	Data struct {
		Redis struct {
			Addr         string        `yaml:"addr"`
			Password     string        `yaml:"password"`
			DB           int           `yaml:"db"`
			PoolSize     int           `yaml:"pool_size"`
			ReadTimeout  time.Duration `yaml:"read_timeout"`
			WriteTimeout time.Duration `yaml:"write_timeout"`
		} `yaml:"redis"`
	} `yaml:"data"`
}

func Init() {
	redisConfig := &RedisConfig{}
	file, err := os.Open("config/config.yaml")
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(redisConfig); err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:         redisConfig.Data.Redis.Addr,
		Password:     redisConfig.Data.Redis.Password,
		DB:           redisConfig.Data.Redis.DB,
		PoolSize:     redisConfig.Data.Redis.PoolSize,
		ReadTimeout:  redisConfig.Data.Redis.ReadTimeout,
		WriteTimeout: redisConfig.Data.Redis.WriteTimeout,
	})
	if err != nil {
		panic(err)
	}
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	log.Infof("Redis client connected to %s", redisConfig.Data.Redis.Addr)
}
func Close() {
	redisClient.Close()
}
