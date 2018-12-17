package timecache

import "time"

//CollectionMethod is the method type, used to determine which collect method should be used
type CollectionMethod int

//the method use to reclaim the cache
const (
	OnUpdate CollectionMethod = iota
)

//TCacheConfig the config used in TimeCache structure
type TCacheConfig struct {
	period time.Duration
	method CollectionMethod
}

//SetCollectionDuration set the duration to check the cache
func (c *TCacheConfig) SetCollectionDuration(t time.Duration) {
	c.period = t
}

//SetCollectionMethod set the method used to judge whether method is expired or not
func (c *TCacheConfig) SetCollectionMethod(m CollectionMethod) {
	c.method = m
}

//GetDefaultTConfig return the default config
func GetDefaultTConfig() *TCacheConfig {
	return &TCacheConfig{
		period: 30 * time.Minute,
		method: OnUpdate,
	}
}
