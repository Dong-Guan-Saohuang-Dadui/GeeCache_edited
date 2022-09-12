package geecache

import (
	"fmt"
	"log"
	"sync"
)

//定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。
//这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。

// Getter 定义接口 Getter 和 回调函数 Get(key string)([]byte, error)，在缓存不命中时调用
//参数是 key，返回值是 []byte
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 定义函数类型 GetterFunc，并实现 Getter 接口的 Get 方法。
type GetterFunc func(key string) ([]byte, error)

// Get 实现 Getter 接口的 Get 方法。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type structName struct {
	Get func(key string) ([]byte, error)
}

// Group theStruct:=structName{GetRealNodeName:.....}
// A Group is a cacheLruMutex namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cacheLruMutex
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	//必须要有回调函数
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cacheLruMutex{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 用来获取特定名称的 Group
// 无则返回nil
func GetGroup(name string) *Group {
	//这里使用了只读锁 RLock()，因为不涉及任何冲突变量的写操作。
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从 cacheLruMutex 获取缓存
func (g *Group) Get(key string) (ByteView, error) {
	//key为空
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	//缓存命中
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit", key)
		return v, nil
	}

	return g.load(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	//其实就是调用用户自定义的getter方法
	bytes, err := g.getter.Get(key)
	if err != nil {
		//在源也没有
		return ByteView{}, err
	}
	//切片需要复制
	value := ByteView{b: cloneBytes(bytes)}
	//加入缓存
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}

	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
