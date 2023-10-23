package lru

import "container/list"

type Cache struct {
	maxBytes int64 //允许使用的最大内存
	nbytes   int64 //当前已经使用的内存大小
	ll       *list.List
	cache    map[string]*list.Element

	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾。
//如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//缓存淘汰。即移除最近最少访问的节点（队首）
// RemoveOldest 从缓存中淘汰最不常访问的元素，即位于队首的元素。
func (c *Cache) RemoveOldest() {
	// 获取队尾元素（最不常访问的元素）
	ele := c.ll.Back()
	if ele != nil {
		// 从双向链表中移除队尾元素
		c.ll.Remove(ele)
		// 通过队尾元素获取其对应的键值对（entry）
		kv := ele.Value.(*entry)
		// 从缓存映射表中删除对应的键
		delete(c.cache, kv.key)
		// 减去被移除元素的大小以更新当前已使用的内存大小
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果定义了回调函数 OnEvicted，执行它，并传递被淘汰元素的键和值作为参数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 将一个键值对添加或更新到缓存中。
func (c *Cache) Add(key string, value Value) {
	// 检查键是否已存在于缓存中
	if ele, ok := c.cache[key]; ok {
		// 如果存在，将对应的节点移动到队首，表示最近访问过
		c.ll.MoveToFront(ele)
		// 获取节点对应的键值对
		kv := ele.Value.(*entry)
		// 更新缓存占用的内存大小，减去旧值大小并加上新值大小
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		// 更新节点的值为新的值
		kv.value = value
	} else {
		// 如果键不存在，创建一个新的节点并添加到队首
		ele := c.ll.PushFront(&entry{key, value})
		// 在缓存映射表中添加新的键值对映射
		c.cache[key] = ele
		// 更新缓存占用的内存大小，加上新键和新值的大小
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 如果设置了最大内存限制且当前内存占用超过了限制
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		// 执行淘汰操作，移除最不常访问的元素
		c.RemoveOldest()
	}
}

//获取添加了多少条数据
func (c *Cache) Len() int {
	return c.ll.Len()
}
