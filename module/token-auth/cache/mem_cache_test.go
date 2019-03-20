package cache

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"emotibot.com/emotigo/module/token-auth/internal/util"
)

func TestMemCache(t *testing.T) {
	util.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr, "TEST")
	cache := NewLocalCache()

	if cache.IsKeyValid("test", "key") {
		t.Errorf("Check key validation fail")
	}
	if cache.Get("test", "key") != nil {
		t.Errorf("Get invalid key doesn't return nil")
	}
	if cache.GetString("test", "key", "default") != "default" {
		t.Errorf("Get invalid key doesn't return default value")
	}

	cache.Set("test", "key", "123", -1)
	if !cache.IsKeyValid("test", "key") {
		t.Errorf("Check key validation fail")
	}
	if cache.Get("test", "key") == nil || cache.Get("test", "key") != "123" {
		t.Errorf("Get valid key value error: %v", cache.Get("test", "key"))
	}
	if cache.GetString("test", "key", "default") != "123" {
		t.Errorf("Get valid key doesn't return correct value")
	}

	cache.Set("test", "key", 123, -1)
	if !cache.IsKeyValid("test", "key") {
		t.Errorf("Check key validation fail")
	}
	if cache.Get("test", "key") == nil || cache.Get("test", "key") != 123 {
		t.Errorf("Get valid key value error: %v", cache.Get("test", "key"))
	}
	if cache.GetString("test", "key", "default") != "123" {
		t.Errorf("Get valid key doesn't return correct value: %s", cache.GetString("test", "key", "default"))
	}
}

func TestCacheTimeout(t *testing.T) {
	util.LogInit(os.Stderr, os.Stderr, os.Stderr, os.Stderr, "TEST")
	cache := NewLocalCache()
	cache.Set("test", "key", "123", 1)

	time.Sleep(time.Second * 5)

	if cache.IsKeyValid("test", "key") {
		t.Errorf("Check key validation fail")
	}
	if cache.Get("test", "key") != nil {
		t.Errorf("Get invalid key doesn't return nil")
	}
	if cache.GetString("test", "key", "default") != "default" {
		t.Errorf("Get invalid key doesn't return default value")
	}
}
