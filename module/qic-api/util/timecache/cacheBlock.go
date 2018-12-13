package timecache

type CacheBlock struct {
	data       interface{}
	updateTime int64
	createTime int64
}
