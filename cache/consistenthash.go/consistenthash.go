package consistenthashgo

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 函数将字节数组映射为一个无符号 32 位整数。
type Hash func(data []byte) uint32

// Map 结构体包含了所有散列过的键。
type Map struct {
	hash     Hash           // 散列函数
	replicas int            // 虚拟节点的数量
	keys     []int          // 按顺序排序的虚拟节点的哈希值
	hashMap  map[int]string // 虚拟节点的哈希值到真实节点的映射
}

// New 创建一个 Map 实例。
// replicas 表示虚拟节点的数量，fn 是散列函数，如果未指定，则默认使用 CRC32 校验和。
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 如果没有指定散列函数，使用默认的 CRC32 校验和函数。
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	// 遍历传入的节点（键）列表。
	for _, key := range keys {
		// 为每个节点（键）创建多个虚拟节点（副本），并为每个虚拟节点计算哈希值。
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点的哈希值，将虚拟节点的索引和节点键组合后进行哈希计算。
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))

			// 将虚拟节点的哈希值添加到 keys 列表中，以便后续查找。
			m.keys = append(m.keys, hash)

			// 在 hashMap 中建立虚拟节点的哈希值到实际节点的映射。
			m.hashMap[hash] = key
		}
	}
	// 对 keys 列表中的虚拟节点哈希值进行排序，以便进行二分查找。
	sort.Ints(m.keys)
}

// Get 方法用于根据给定的键（key）查找对应的节点。
func (m *Map) Get(key string) string {
	// 如果没有任何节点可用，直接返回空字符串。
	if len(m.keys) == 0 {
		return ""
	}

	// 计算给定键的哈希值，将其转换为整数类型。
	hash := int(m.hash([]byte(key)))

	// 使用二分查找算法查找最接近的虚拟节点哈希值。
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 通过模运算找到最终映射的节点。
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
