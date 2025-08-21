package data

import (
	"fmt"
	"time"
)

// UserRedisKey 根据不同参数生成redisKey user:<field>:<value>
func UserRedisKey(cachePrefix, value interface{}) string {
	return fmt.Sprintf("%s:%v", cachePrefix, value)
}

const (
	DefaultCacheTTL = 24 * time.Hour     // 默认缓存 24 小时
	UserCacheTTL    = 24 * time.Hour     // 用户信息缓存 1 天
	TokenCacheTTL   = 7 * 24 * time.Hour // token 缓存 7 天
)

const (
	UserCachePrefix  = "user"
	LoginCachePrefix = "login"
	TokenCachePrefix = "token"
)
