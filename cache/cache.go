package cache

import (
	"runtime/debug"

	"github.com/coocood/freecache"
)

var Cache *freecache.Cache

func InitCache() {
	cacheSize := 500 * 1024 * 1024
	Cache = freecache.NewCache(cacheSize)
	debug.SetGCPercent(20)
}
