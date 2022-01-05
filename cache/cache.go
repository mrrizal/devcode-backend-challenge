package cache

import (
	"time"

	"github.com/allegro/bigcache/v3"
)

var Cache *bigcache.BigCache

func InitCache() {
	cache, _ := bigcache.NewBigCache(bigcache.DefaultConfig(60 * time.Second))
	Cache = cache
}
