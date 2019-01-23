package cache

type Cache interface {
	Get(namespace, key string) interface{}
	GetString(namespace, key, defaultValue string) string

	Set(namespace, key string, value interface{}, expire int)

	IsKeyValid(namespace, key string) bool
}
