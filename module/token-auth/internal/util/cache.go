package util

type cacheMap map[string]string
type Cache map[string]*cacheMap

var internalCache *Cache

func init() {
	internalCache = &Cache{}
}

func GetCacheValue(module, key string) string {
	if internalCache == nil {
		return ""
	}
	if m, ok := (*internalCache)[module]; ok {
		if v, ok := (*m)[key]; ok {
			return v
		}
	}
	return ""
}

func SetCache(module, key, value string) {
	if internalCache == nil {
		internalCache = &Cache{}
	}

	if _, ok := (*internalCache)[module]; !ok {
		(*internalCache)[module] = &cacheMap{}
	}
	m := (*internalCache)[module]
	(*m)[key] = value
}
