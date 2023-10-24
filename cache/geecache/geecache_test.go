package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// TestGet 是一个测试函数，用于测试从缓存中获取值的逻辑。
// 在这个测试中，首先初始化一个 GeeCache 组（gee）并用于从缓存中获取值。
// 该组的 GetterFunc 用于模拟数据获取，实际上从一个模拟的数据库（db）中获取值。
func TestGet(t *testing.T) {
	// 创建一个用于记录加载次数的映射，初始化为空，长度与模拟数据库的键值对数量相同。
	loadCounts := make(map[string]int, len(db))

	// 创建一个 GeeCache 组（gee），指定组名为 "scores"，缓存大小为 2^10 字节。
	// 使用 GetterFunc 包装的函数用于模拟数据获取逻辑。
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			// 模拟数据获取的过程，记录日志并查找模拟数据库（db）中的数据。
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				// 如果在模拟数据库中找到了值，将其封装为字节数组返回。
				// 同时，记录加载次数以模拟缓存命中。
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			// 如果未找到值，返回错误信息。
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 遍历模拟数据库中的键值对，尝试从 GeeCache 组（gee）中获取值。
	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function

		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	// 测试获取一个未知键时，预期会返回空值并报告错误。
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
