package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashFunc 自定义的hash函数
type HashFunc func(data []byte) uint32

// ConsistentHashingMaster 是一致性哈希算法的主数据结构
type ConsistentHashingMaster struct {
	//hash函数
	hashFunc HashFunc
	//虚拟节点倍率
	replicas int
	//hash环
	keysHashLoopArr []int
	//虚拟节点和真实节点的映射关系
	VirNodeToRealNodeMap map[int]string
}

func New(replicas int, hashFunc HashFunc) *ConsistentHashingMaster {
	m := &ConsistentHashingMaster{
		hashFunc:             hashFunc,
		replicas:             replicas,
		VirNodeToRealNodeMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		//默认hash为crc32.ChecksumIEEE
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// Add 函数传入 0 或 多个真实节点的名称添加到环(Arr)和map
func (m *ConsistentHashingMaster) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//计算hash
			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
			//添加到环
			m.keysHashLoopArr = append(m.keysHashLoopArr, hash)
			//在 VirNodeToRealNodeMap 中增加虚拟节点和真实节点的映射关系
			m.VirNodeToRealNodeMap[hash] = key
		}
	}
	//整理顺序
	sort.Ints(m.keysHashLoopArr)
}

// GetRealNodeName gets the closest item in the hash to the provided key.
func (m *ConsistentHashingMaster) GetRealNodeName(key string) string {
	if len(m.keysHashLoopArr) == 0 {
		return ""
	}
	//求key hash值
	hash := int(m.hashFunc([]byte(key)))
	// Search函数采用二分法搜索找到[0, n)区间内最小的满足f(i)==true的值i
	//即keysLoop大于key hash的最小值
	idx := sort.Search(len(m.keysHashLoopArr), func(i int) bool {
		return m.keysHashLoopArr[i] >= hash
	})
	//如果 idx == len(m.keys)，说明应选择 m.keys[0]
	//因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	return m.VirNodeToRealNodeMap[m.keysHashLoopArr[idx%len(m.keysHashLoopArr)]]
}
