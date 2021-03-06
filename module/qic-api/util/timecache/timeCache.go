package timecache

import (
	"errors"
	"sync"
	"time"
)

//TimeCache the structure of time cache
type TimeCache struct {
	mux         sync.Mutex
	cacheMap    map[string]*CacheBlock
	expireCheck expireMethod
	period      int64
	stop        chan struct{}
	isInit      bool
}

//TCache the default time cache
var TCache TimeCache

//Activate activate the time cache
func (c *TimeCache) Activate(config *TCacheConfig) error {

	if config == nil {
		return errors.New("No config is assigned")
	}
	if config.period == 0 {
		return errors.New("No period is assigned")
	}

	c.mux.Lock()

	if !c.isInit {
		c.period = int64(config.period / time.Second)
		switch config.method {
		case OnUpdate:
			c.expireCheck = &onUpdateMethod{}
		}
		c.cacheMap = make(map[string]*CacheBlock)
		c.stop = make(chan struct{})
		go waker(c, config.period)
		c.isInit = true
	}

	c.mux.Unlock()
	return nil
}

func (c *TimeCache) recycleExpired() {
	c.mux.Lock()
	for k, v := range c.cacheMap {
		if c.expireCheck.IsExpire(v, c.period) {
			delete(c.cacheMap, k)
		}
	}
	c.mux.Unlock()
}

//SetCache cache the data
func (c *TimeCache) SetCache(key string, value interface{}) {
	c.mux.Lock()
	if c.isInit {

		now := time.Now().Unix()

		val, ok := c.cacheMap[key]
		if !ok {
			b := &CacheBlock{data: value, updateTime: now, createTime: now}
			c.cacheMap[key] = b
		} else {
			val.updateTime = now
			val.data = value
		}
	}
	c.mux.Unlock()
}

//GetCache get cache data
func (c *TimeCache) GetCache(key string) (interface{}, bool) {

	var v interface{}
	var ok bool
	var b *CacheBlock
	c.mux.Lock()
	if c.isInit {
		b, ok = c.cacheMap[key]
		if ok {
			v = b.data
			b.updateTime = time.Now().Unix()
		}
	}
	c.mux.Unlock()
	return v, ok
}

//DeleteCache delete the data by key
func (c *TimeCache) DeleteCache(key string) {
	c.mux.Lock()
	if _, ok := c.cacheMap[key]; ok {
		delete(c.cacheMap, key)
	}
	c.mux.Unlock()
}

//Destroy clean the cache
func (c *TimeCache) Destroy() {
	c.mux.Lock()
	c.stop <- struct{}{}
	c.expireCheck = &alwaysMethod{}
	c.isInit = false
	c.mux.Unlock()
	c.recycleExpired()
}

func waker(c *TimeCache, period time.Duration) {
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			c.recycleExpired()
		case <-c.stop:
			ticker.Stop()
			break
		}
	}
}
