package geecache

import (
	"sync"
	"testProject/cache/lru"
)

// cache 结构体用于管理缓存，包含了互斥锁、LRU 缓存、以及缓存大小限制。
type cache struct {
	mu         sync.Mutex // 互斥锁，用于在并发操作中保护缓存数据
	lru        *lru.Cache // LRU 缓存实例，用于实现缓存淘汰策略
	cacheBytes int64      // 缓存的最大内存限制
}

// add 方法用于向缓存中添加键值对。
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()         // 加锁以确保并发安全
	defer c.mu.Unlock() // 函数返回前解锁

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil) // 如果 LRU 缓存为空，创建一个新的
	}

	c.lru.Add(key, value) // 调用 LRU 缓存的 Add 方法，将键值对添加到缓存中
}

// get 方法用于从缓存中获取指定键的值。
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()         // 加锁以确保并发安全
	defer c.mu.Unlock() // 函数返回前解锁

	if c.lru == nil {
		return // 如果 LRU 缓存为空，直接返回
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok // 调用 LRU 缓存的 Get 方法，返回对应键的值和是否命中
	}

	return // 如果未命中，直接返回
}
