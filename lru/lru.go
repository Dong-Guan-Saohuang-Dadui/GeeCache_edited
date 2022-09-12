package lru

import "container/list"

// CacheLru LRU,并发不安全
type CacheLru struct {
	maxBytes    int64      //最大内存
	nBytes      int64      //已使用内存
	ll          *list.List //Go自带双向链表
	cacheLruMap map[string]*list.Element
	//删除缓存时的回调函数
	OnEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, OnEvicted func(string, Value)) *CacheLru {
	return &CacheLru{
		maxBytes:    maxBytes,
		ll:          list.New(),
		cacheLruMap: make(map[string]*list.Element),
		OnEvicted:   OnEvicted,
	}
}

// Add 添加与更新
func (cache *CacheLru) Add(key string, value Value) {
	//防止value大于内存大小
	if int64(value.Len()) > cache.maxBytes {
		return
	}
	if ele, ok := cache.cacheLruMap[key]; ok {
		//命中情况，更新缓存+更新已用容量
		cache.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		cache.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		//不命中情况，添加缓存+更新已用容量
		cache.cacheLruMap[key] = cache.ll.PushFront(&entry{key, value})
		cache.nBytes += int64(value.Len()) + int64(len(key))
	}
	//担心恶意攻击
	for cache.maxBytes != 0 && cache.maxBytes < cache.nBytes {
		cache.RemoveOldest()
	}
}

// RemoveOldest 移除旧元素
func (cache *CacheLru) RemoveOldest() {
	ele := cache.ll.Back()
	if ele != nil {
		cache.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(cache.cacheLruMap, kv.key)
		cache.nBytes -= int64(kv.value.Len()) + int64(len(kv.key))
		if cache.OnEvicted != nil {
			cache.OnEvicted(kv.key, kv.value)
		}
	}
}

// Get 查找元素
func (cache *CacheLru) Get(key string) (value Value, ok bool) {
	if ele, ok := cache.cacheLruMap[key]; ok {
		cache.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (cache *CacheLru) ListLen() int64 {
	return int64(cache.ll.Len())
}
