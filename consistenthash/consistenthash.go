package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// 包含所有的hash keys

type Map struct {
	hash     Hash           // 自定义的哈希函数
	replicas int            // 虚拟结点倍数
	keys     []int          // 哈希环,存的是hash值
	hashMap  map[int]string // 虚拟节点与真实节点的映射表，key是虚拟结点Hash值，value是真实节点
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	// hash函数采用依赖注入的方式，允许替换成自定义的Hash函数
	// 默认为crc32.ChecksumIEEE算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点/机器的Add()方法
// 传入0或者多个真实节点的名称

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 一个真实节点key，对应创建m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	// 环上的哈希值排序
	sort.Ints(m.keys)
}

// 选择节点的Get()方法

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找第一个匹配hash的节点hash
	// 匹配规则为第一个大于等于hash的值
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 拿到hash值对应的真实节点key
	// 环形结构
	return m.hashMap[m.keys[idx%len(m.keys)]]

}
