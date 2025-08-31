package data

import (
	"context"
	"errors"
	"fmt"
	"kratos-realworld/internal/model"
	"reflect"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/mapstructure"
)

// UserRedisKey 根据不同参数生成redisKey <prefix1>:<prefix2>:<value>
func UserRedisKey(cachePrefix, subPrefix, value interface{}) string {
	return fmt.Sprintf("%s:%v:%v", cachePrefix, subPrefix, value)
}

// HSetStruct 将结构体struct转换成map并逐个写入 Redis Hash
func HSetStruct(ctx context.Context, data *model.Data, log *log.Helper, key string, obj interface{}) error {
	values := StructToMap(obj)

	// HSet批量写入
	_, err := data.Cache().HMSet(ctx, key, values)
	if err != nil {
		log.Warnf("failed to cache data: %v", err)
		return err
	}

	data.Cache().Expire(ctx, key, UserCacheTTL)
	log.Debugf("data cached successfully, set TTL to %s for key %s", UserCacheTTL, key)

	return nil
}

// HGetStruct 从 Redis Hash 获取值
func HGetStruct(ctx context.Context, data *model.Data, log *log.Helper, key string, obj interface{}) error {
	// HLen判断该数据是否放在redis缓存中
	length, err := data.Cache().HLen(ctx, key)
	if err != nil {
		log.Warnf("failed to get hash length, fallback to DB: %v", err)
		return err
	}
	if length == 0 {
		log.Debugf("hash key %s is empty", key)
		return errors.New("cache miss")
	}

	res, err := data.Cache().HGetAll(ctx, key)
	if err != nil {
		log.Warnf("failed to get from cache, fallback to DB: %v", err)
		return err
	}

	err = MapToStruct(res, obj)
	if err != nil {
		log.Warnf("failed to map redis result to struct: %v", err)
		return err
	}

	data.Cache().Expire(ctx, key, UserCacheTTL)
	log.Debugf("get data from cache successfully, refreshed TTL to %s for key %s", UserCacheTTL, key)
	return nil
}

// StructToMap 把struct转成map
func StructToMap(obj interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	// 如果是指针，取 Elem
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		// 忽略嵌套 struct
		if reflect.TypeOf(value).Kind() == reflect.Struct {
			continue
		}

		// 用 json tag 做 key（没有就用字段名）
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		res[tag] = value
	}
	return res
}

// MapToStruct 将 map[string]string 自动填充到 struct 指针 obj 中
func MapToStruct(data map[string]string, obj interface{}) error {
	// 转换成 map[string]interface{}
	m := make(map[string]interface{}, len(data))
	for k, v := range data {
		m[k] = v
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           obj,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}

	_ = decoder.Decode(m)

	return nil
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
