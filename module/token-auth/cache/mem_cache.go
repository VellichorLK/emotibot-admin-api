package cache

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/module/token-auth/internal/util"
)

// LocalCache will use cache in local memory
type LocalCache struct {
	Timeout map[string]map[string]int64
	Cache   map[string]map[string]interface{}
}

// NewLocalCache will return an initialized LocalCache
func NewLocalCache() LocalCache {
	return LocalCache{
		Timeout: map[string]map[string]int64{},
		Cache:   map[string]map[string]interface{}{},
	}
}

// Get will get value in cache if key is valid
func (cache LocalCache) Get(namespace, key string) interface{} {
	util.LogTrace.Println("GET")
	if _, ok := cache.Cache[namespace]; ok {
		util.LogTrace.Println("Name space ok")
		timeout := cache.Timeout[namespace][key]
		if !isTimeout(timeout) {
			util.LogTrace.Println("not timeout")
			return cache.Cache[namespace][key]
		} else {
			cache.Set(namespace, key, nil, 0)
		}
	}
	return nil
}

// GetString will try to convert get value to string
// It will return defaultValue if key is invalid or value type is not string
func (cache LocalCache) GetString(namespace, key, defaultValue string) string {
	val := cache.Get(namespace, key)
	if val == nil {
		return defaultValue
	}
	if strVal, ok := val.(string); ok {
		return strVal
	}
	return fmt.Sprintf("%+v", val)
}

// IsKeyValid will check key is expire or not
func (cache LocalCache) IsKeyValid(namespace, key string) bool {
	if _, ok := cache.Cache[namespace]; ok {
		timeout := cache.Timeout[namespace][key]
		return !isTimeout(timeout)
	}
	return false
}

// Set will store value in memory with specific expire time
// If expire is negative, this key will not timeout
func (cache LocalCache) Set(namespace, key string, value interface{}, expire int) {
	if _, ok := cache.Cache[namespace]; !ok {
		cache.Cache[namespace] = map[string]interface{}{}
		cache.Timeout[namespace] = map[string]int64{}
	}

	if value == nil {
		delete(cache.Cache[namespace], key)
		return
	}

	now := time.Now().Unix()
	cache.Cache[namespace][key] = value

	if expire < 0 {
		cache.Timeout[namespace][key] = -1
	} else {
		cache.Timeout[namespace][key] = now + int64(expire)
	}
}

func isTimeout(timeout int64) bool {
	t := time.Now().Unix()
	return timeout >= 0 && t > timeout
}
