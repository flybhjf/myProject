package getter

// Getter 接口定义了获取键值对数据的方法。
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 类型是一个函数类型，它实现了 Getter 接口。
type GetterFunc func(key string) ([]byte, error)

// Get 方法实现了 Getter 接口的 Get 方法。
// 这个方法调用 GetterFunc 类型的函数来获取键对应的值，并返回该值及可能的错误。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
