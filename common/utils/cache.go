package utils

import (
	"encoding/json"

	"github.com/coocood/freecache"
	"golang.org/x/sync/singleflight"
)

func CacheGet[T any](cache *freecache.Cache, sf *singleflight.Group, key string, fn func() (T, error), ttl int) (T, error) {
	var data T
	cacheKey := []byte(key)

	if val, err := cache.Get(cacheKey); err == nil {
		if err = json.Unmarshal(val, &data); err == nil {
			return data, nil
		}
	}

	result, err, _ := sf.Do(key, func() (interface{}, error) { return fn() })
	if err != nil {
		return data, err
	}
	data = result.(T)

	if bjson, err := json.Marshal(data); err == nil {
		cache.Set(cacheKey, bjson, ttl)
	}

	return data, nil
}
