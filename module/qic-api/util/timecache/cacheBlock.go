package timecache

//CacheBlock the data block to store the user data and set the timestamp
type CacheBlock struct {
	data       interface{}
	updateTime int64
	createTime int64
}
