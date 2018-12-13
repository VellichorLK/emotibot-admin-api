package timecache

import (
	"testing"
	"time"
)

func TestTimeCache(t *testing.T) {
	config := &TCacheConfig{}
	config.SetCollectionDuration(2 * time.Second)
	config.SetCollectionMethod(OnUpdate)

	TCache.Activate(config)

	TCache.SetCache("hello", "world")

	for i := 0; i < 5; i++ {

		v, ok := TCache.GetCache("hello")

		if !ok {
			t.Fatalf("[Unit test] getting key, hello, failed\n")
		}
		if v.(string) != "world" {
			t.Fatalf("[Unit test] getting key value %s, but expecting world\n", v.(string))
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(3 * time.Second)
	_, ok := TCache.GetCache("hello")
	if ok {
		t.Fatal("[Unit test] expecting no data for key hello, but get data\n")
	}

}
