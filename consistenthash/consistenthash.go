package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Map struct {
	hash Hash
	// 虚拟节点数
	virtualNodeCount int
	keys             []int
	hashMap          map[int]string
}

// 一致性算法
type Hash func(data []byte) uint32

func New(count int, hash Hash) *Map {
	m := &Map{
		hash:             hash,
		virtualNodeCount: count,
		hashMap:          make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 根据环的顺时针来选择命中的节点。
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]

}

func (m *Map) Set(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.virtualNodeCount; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}
