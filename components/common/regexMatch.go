package common

import (
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxCacheSize  = 1000          // 最大缓存容量
	cacheTimeout  = 1 * time.Hour // 缓存超时时间（1小时未访问则清理）
	cleanupPeriod = 1 * time.Hour // 定期清理间隔
)

// cacheItem 缓存项结构，包含正则对象和最后访问时间
type cacheItem struct {
	re         *regexp.Regexp
	lastAccess int64 // 纳秒级时间戳，原子操作更新
}

var (
	regexCache    sync.Map
	cleanupTicker *time.Ticker
)

func init() {
	cleanupTicker = time.NewTicker(cleanupPeriod)
	go func() {
		for range cleanupTicker.C {
			cleanupCache()
		}
	}()
}

// RegexMatch 判断字节数组是否匹配正则表达式，错误情况（如无效正则）返回false
func RegexMatch(bytesToMatch []byte, regexStr string) bool {
	// 空输入快速返回
	if len(bytesToMatch) == 0 || regexStr == "" {
		return false
	}

	// 尝试从缓存加载
	if val, ok := regexCache.Load(regexStr); ok {
		item := val.(*cacheItem)
		// 原子更新最后访问时间
		atomic.StoreInt64(&item.lastAccess, time.Now().UnixNano())
		return item.re.Match(bytesToMatch)
	}

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return false // 编译失败视为不匹配
	}

	// 准备缓存项（初始访问时间为当前时间）
	now := time.Now().UnixNano()
	newItem := &cacheItem{
		re:         re,
		lastAccess: now,
	}

	actual, loaded := regexCache.LoadOrStore(regexStr, newItem)
	if loaded {
		// 其他goroutine已存储，使用已有项并更新访问时间
		existingItem := actual.(*cacheItem)
		atomic.StoreInt64(&existingItem.lastAccess, now)
		return existingItem.re.Match(bytesToMatch)
	}

	// 新项存储后检查容量，超则清理最久未用项
	cleanupIfNeeded()

	return newItem.re.Match(bytesToMatch)
}

// cleanupCache 清理超时项和超容量项
func cleanupCache() {
	now := time.Now().UnixNano()
	timeoutNano := int64(cacheTimeout)

	// 清理超时项（1小时未访问）
	regexCache.Range(func(key, value interface{}) bool {
		item := value.(*cacheItem)
		if now-atomic.LoadInt64(&item.lastAccess) > timeoutNano {
			regexCache.Delete(key)
		}
		return true
	})

	// 检查容量，超则清理最久未用项
	cleanupIfNeeded()
}

// cleanupIfNeeded 若缓存超容量，清理最久未用项
func cleanupIfNeeded() {
	// 统计当前缓存数量
	count := 0
	regexCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	if count <= maxCacheSize {
		return
	}

	// 计算需要清理的数量（保留maxCacheSize项）
	evictCount := count - maxCacheSize
	evictOldest(evictCount)
}

// evictOldest 清理指定数量的最久未用项
func evictOldest(evictCount int) {
	if evictCount <= 0 {
		return
	}

	// 收集所有项的键和访问时间
	type keyAccess struct {
		key        string
		lastAccess int64
	}
	var items []keyAccess

	regexCache.Range(func(key, value interface{}) bool {
		k := key.(string)
		item := value.(*cacheItem)
		items = append(items, keyAccess{
			key:        k,
			lastAccess: atomic.LoadInt64(&item.lastAccess),
		})
		return true
	})

	// 按访问时间升序排序（最久未用在前）
	sort.Slice(items, func(i, j int) bool {
		return items[i].lastAccess < items[j].lastAccess
	})

	// 清理前evictCount项
	cleanupNum := evictCount
	if cleanupNum > len(items) {
		cleanupNum = len(items)
	}
	for i := 0; i < cleanupNum; i++ {
		regexCache.Delete(items[i].key)
	}
}

// RegexCleanStop 停止定时清理任务
func RegexCleanStop() {
	if cleanupTicker != nil {
		cleanupTicker.Stop()
	}
}
