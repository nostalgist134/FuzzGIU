package common

import (
	"regexp"
	"sync"
	"time"
)

var (
	regexCache    sync.Map
	cleanupTicker *time.Ticker
)

func init() {
	// 启动定期清理任务（可选）
	cleanupTicker = time.NewTicker(1 * time.Hour)
	go func() {
		for range cleanupTicker.C {
			cleanupCache()
		}
	}()
}

func RegexMatch(bytesToMatch []byte, regexStr string) bool {
	if len(bytesToMatch) == 0 || regexStr == "" {
		return false
	}
	// 尝试获取缓存
	if cached, ok := regexCache.Load(regexStr); ok {
		return cached.(*regexp.Regexp).Match(bytesToMatch)
	}
	// 编译并存储
	compiled := regexp.MustCompile(regexStr)
	regexCache.Store(regexStr, compiled)

	return compiled.Match(bytesToMatch)
}

func cleanupCache() {
	regexCache.Range(func(key, value interface{}) bool {
		// 可根据最后访问时间等策略清理
		regexCache.Delete(key)
		return true
	})
}
