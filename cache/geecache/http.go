package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// const defaultBasePath = "/_geecache/"

// // HTTPPool 结构体实现了 PeerPicker 接口，用于管理一组 HTTP 同伴。
// type HTTPPool struct {
// 	// 本节点的基本 URL，例如 "https://example.net:8000"
// 	self     string //用来记录自己的地址，包括主机名/IP 和端口
// 	basePath string //作为节点间通讯地址的前缀，默认是 /_geecache/
// }

// NewHTTPPool 创建并初始化一个 HTTPPool 实例。
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 用于记录带有服务器名称的日志信息。
// 它接受一个格式字符串和可选的参数，并使用服务器名称格式化日志消息。
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有的 HTTP 请求。
// 它接受一个 HTTP 响应写入器（w）和 HTTP 请求（r）作为参数。
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 检查请求路径是否以指定的基本路径（basePath）开头。
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	// 记录日志，包括 HTTP 方法和请求路径。
	p.Log("%s %s", r.Method, r.URL.Path)

	// 从请求路径中提取组名（groupName）和键（key）。
	// 请求路径格式为 /<basepath>/<groupname>/<key>。
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		// 如果路径不符合预期格式，返回 "bad request" 错误。
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 根据组名获取对应的缓存组（group）。
	group := GetGroup(groupName)
	if group == nil {
		// 如果找不到对应的组，返回 "no such group" 错误。
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 使用组的 Get 方法获取指定键（key）的数据视图（view）。
	view, err := group.Get(key)
	if err != nil {
		// 如果获取失败，返回内部服务器错误并包含错误信息。
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头的内容类型为 "application/octet-stream"。
	w.Header().Set("Content-Type", "application/octet-stream")
	// 将数据视图（view）的字节切片写入响应。
	w.Write(view.ByteSlice())
}
