package geecache

import (
	"fmt"
	"go_code/geecache/consistenthash"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	//自己的地址
	self string
	//通信前缀
	basePath string
	mu       sync.Mutex
	//一致性哈希map
	peersMaster *consistenthash.ConsistentHashingMaster
	//每一个远程节点对应一个 httpGetter，
	//因为 httpGetter 与远程节点的地址 baseURL 有关。
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 打印调用消息
func (hp *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", hp.self, fmt.Sprintf(format, v...))
}

func (hp *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//前缀不符合要求
	if !strings.HasPrefix(r.URL.Path, hp.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	hp.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> 为固定格式
	//使用strings.SplitN切割出groupName和key
	parts := strings.SplitN(r.URL.Path[len(hp.basePath):], "/", 2)
	//url格式错误
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		//500服务器内部错误(Internal server error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header()["Content-Type"] = []string{"application/octet-stream"}
	w.Write(view.ByteSlice())
}

func (hp *HTTPPool) Set(realPeersName ...string) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	hp.peersMaster = consistenthash.New(defaultReplicas, nil)
	hp.peersMaster.Add(realPeersName...)
	hp.httpGetters = make(map[string]*httpGetter, len(realPeersName))
	for _, realPeerName := range realPeersName {
		hp.httpGetters[realPeerName] = &httpGetter{baseURL: realPeerName + hp.basePath}
	}
}

// PickPeer picks a peer according to key
func (hp *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	if peer := hp.peersMaster.GetRealNodeName(key); peer != "" && peer != hp.self {
		hp.Log("Pick peer %s", peer)
		return hp.httpGetters[peer], true
	}
	return nil, false
}
