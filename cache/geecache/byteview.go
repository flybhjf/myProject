package geecache

// ByteView 表示一个不可变的字节视图。
type ByteView struct {
	b []byte // 存储字节数据的切片
}

// Len 返回视图的长度
func (v ByteView) Len() int {
	return len(v.b) // 返回字节切片的长度
}

// ByteSlice 返回数据的字节切片副本。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) // 调用 cloneBytes 函数，返回一个字节切片的深拷贝
}

// String 返回数据作为字符串，如果需要则创建一个副本。
func (v ByteView) String() string {
	return string(v.b) // 将字节切片转换为字符串并返回
}

// cloneBytes 创建并返回字节切片的深拷贝。
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b)) // 创建与原字节切片相同长度的新字节切片
	copy(c, b)                // 将原字节切片的数据复制到新切片
	return c                  // 返回新切片，它是原切片的深拷贝
}
