package geecache

import (
	"fmt"
	"log"
	"sync"
	"testProject/cache/singleflight"
)

// 回调函数 缓存未命中时从数据库中读取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 结构表示一个缓存组，包括组名、Getter 接口实现和主缓存。
// type Group struct {
// 	name      string // 组的名称
// 	getter    Getter // 数据获取接口 :缓存未命中时 获取源数据的回调
// 	mainCache cache  // 主缓存：并发缓存
// }

var (
	mu     sync.RWMutex              // 用于保护 groups 映射的读写锁
	groups = make(map[string]*Group) // 存储已创建的组的映射
)

// NewGroup 创建一个新的 Group 实例。
// 它接受组名、缓存大小限制（cacheBytes），以及实现 Getter 接口的数据获取器（getter）。
// 如果 getter 为 nil，将会引发 panic。

// GetGroup 返回之前使用 NewGroup 创建的具有指定名称的组，如果没有找到则返回 nil。
func GetGroup(name string) *Group {
	mu.RLock()        // 以只读模式加锁以防止数据竞争
	g := groups[name] // 查找指定名称的组
	mu.RUnlock()      // 解锁
	return g          // 返回找到的组，或者 nil 如果未找到
}

// Get 方法用于从缓存中获取指定键的值。
// 它接受一个键名作为参数，返回一个 ByteView 和可能的错误。
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required") // 如果键为空，返回错误
	}

	// 尝试从主缓存中获取值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit") // 命中缓存，记录日志
		return v, nil
	}

	// 如果没有命中，调用 load 方法来加载数据
	return g.load(key)
}

// load 方法用于加载指定键的数据。
// 它接受一个键名作为参数，调用 getLocally 方法从数据源获取数据，并将数据加载到缓存中。
// func (g *Group) load(key string) (value ByteView, err error) {
// 	return g.getLocally(key) // 调用 getLocally 方法获取数据
// }

// getLocally 方法用于从数据源获取指定键的数据。
// 它接受一个键名作为参数，调用 Getter 接口的 Get 方法从数据源获取数据。
// 如果获取成功，将数据封装为 ByteView，并调用 populateCache 方法将数据存入缓存。
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 从数据源获取数据
	if err != nil {
		return ByteView{}, err // 如果获取失败，返回错误
	}
	value := ByteView{b: cloneBytes(bytes)} // 将数据封装为 ByteView
	g.populateCache(key, value)             // 存入缓存
	return value, nil                       // 返回数据视图
}

// populateCache 方法用于将指定键值对存入缓存。
// 它接受一个键名和 ByteView 作为参数，将数据存入主缓存。
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value) // 将数据存入主缓存
}

// RegisterPeers 方法用于注册一个 PeerPicker，用于选择远程对等节点。
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// getFromPeer 方法用于从远程对等节点获取数据。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// Group 结构体表示一个缓存命名空间，以及相关的数据分布在多个节点上。
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	// 使用 singleflight.Group 以确保每个键只获取一次
	loader *singleflight.Group
}

// NewGroup 创建一个新的 Group 实例。
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	// ...
	g := &Group{
		// ...
		loader: &singleflight.Group{},
	}
	return g
}

// load 方法用于从缓存或远程节点加载数据。
func (g *Group) load(key string) (value ByteView, err error) {
	// 确保每个键只被获取一次（无论有多少并发调用）
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}
