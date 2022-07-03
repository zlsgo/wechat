package wechat

import (
	"sync"
	"time"
)

var (
	CacheFile     = "wechat.json"
	CacheTime     = time.Second * 60 * 10
	cacheFileOnce sync.Once
)
