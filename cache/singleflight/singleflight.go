package singleflight

import "sync"

//call 代表正在进行中，或已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// singleflight 的主数据结构，管理不同 key 的请求(call)
type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// Do 方法接受一个键值（key）和一个函数（fn）作为参数，用于处理缓存请求。
// 如果缓存中已经有该键值的调用，它会等待调用结果并返回结果。
// 如果缓存中没有该键值的调用，它会执行提供的函数 fn，并将结果存储在缓存中。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 加锁以确保在并发访问中的安全性

	// 如果缓存 map 为空，初始化它
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 检查缓存中是否已经存在该键值的调用
	if c, ok := g.m[key]; ok {
		g.mu.Unlock() // 解锁
		c.wg.Wait()   // 等待调用结果
		return c.val, c.err
	}

	// 如果缓存中没有该键值的调用，创建一个新的调用并存储在缓存中
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock() // 解锁

	// 执行提供的函数 fn，获取结果
	c.val, c.err = fn()
	c.wg.Done() // 通知调用已经完成

	g.mu.Lock()      // 再次加锁以进行最后的处理
	delete(g.m, key) // 从缓存中删除调用结果
	g.mu.Unlock()    // 解锁

	return c.val, c.err // 返回调用结果
}
