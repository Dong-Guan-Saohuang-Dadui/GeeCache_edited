package geecache

import (
	"go_code/geecache/lru"
	"sync"
)

//lru 加锁解决冲突
type cacheLruMutex struct {
	//加锁
	mu  sync.Mutex
	lru *lru.CacheLru
	//最大容量
	cacheBytes int64
}

//add 添加ByteView判断了 c.lru 是否为 nil，如果等于 nil 再创建实例
func (c *cacheLruMutex) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//延迟初始化(Lazy Initialization)，
	//一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。
	//主要用于提高性能，并减少程序内存要求。
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	//使用lru.Add添加一个ByteView
	c.lru.Add(key, value)
}

func (c *cacheLruMutex) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//未创建lru直接返回
	if c.lru == nil {
		return
	}
	//命中
	if v, ok := c.lru.Get(key); ok {
		//缓存命中，返回
		return v.(ByteView), ok
	}
	//不命中
	return
}
