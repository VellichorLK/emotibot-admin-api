package timecache

import (
	"time"
)

type expireMethod interface {
	IsExpire(b *CacheBlock, period int64) bool
}

type onUpdateMethod struct {
}

type alwaysMethod struct {
}

//IsExpire is used to check whether the cache block is expired
func (m *alwaysMethod) IsExpire(b *CacheBlock, period int64) bool {
	return true
}

//IsExpire is used to check whether the cache block is expired
func (m *onUpdateMethod) IsExpire(b *CacheBlock, period int64) bool {
	if b != nil {
		now := time.Now().Unix()
		if b.updateTime+period >= now {
			return false
		}
	}
	return true
}
