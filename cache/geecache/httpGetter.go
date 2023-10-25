package geecache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/golang/groupcache/consistenthash"
)

// httpGetter 结构体表示一个 HTTP 请求获取器，用于向远程 HTTP 服务器发起 GET 请求。
type httpGetter struct {
	baseURL string // baseURL 存储远程服务器的基本 URL 地址
}

// Get 方法用于从远程服务器获取指定 group 和 key 对应的数据。
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// 构建完整的请求 URL，将 group 和 key 编码为 URL 安全格式。
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	// 发起 HTTP GET 请求。
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 检查响应状态码，如果不是 200 OK，则返回错误。
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	// 读取响应体的内容。
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// httpGetter 类型实现了 PeerGetter 接口，这意味着它可以作为 PeerGetter 接口的实现。
// 这是通过将 (*httpGetter)(nil) 赋值给 _ PeerGetter 来实现的，表示 httpGetter 满足 PeerGetter 接口的要求。
var _ PeerGetter = (*httpGetter)(nil)

// defaultBasePath 定义了 HTTP 池的默认基本路径。
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool 结构体实现了 PeerPicker 接口，用于管理一组 HTTP 对等节点的池。
type HTTPPool struct {
	// self 表示当前节点的基本 URL 地址，例如 "https://example.net:8000"。
	self        string
	basePath    string
	mu          sync.Mutex             // 互斥锁，用于保护 peers 和 httpGetters。
	peers       *consistenthash.Map    // 一致性哈希算法的映射，用于管理对等节点。
	httpGetters map[string]*httpGetter // 存储 HTTP 请求获取器的映射，按键值 "http://10.0.0.2:8008" 存储。
}

// Set 方法用于更新池的对等节点列表。
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建一个新的一致性哈希映射，设置副本数为默认值，并将传入的节点添加到映射中。
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)

	// 初始化 HTTP 请求获取器映射，为每个节点创建一个对应的 HTTP 客户端。
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 方法根据给定的键选择一个对等节点。
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 使用一致性哈希算法根据键获取对等节点。
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		// 如果找到了合适的对等节点，则返回对应的 HTTP 客户端。
		return p.httpGetters[peer], true
	}

	// 如果没有找到合适的对等节点，返回 nil 和 false。
	return nil, false
}

// HTTPPool 类型实现了 PeerPicker 接口，这表示它可以用作 PeerPicker 接口的实现。
var _ PeerPicker = (*HTTPPool)(nil)
