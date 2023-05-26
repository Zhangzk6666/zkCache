package consistenthash

import (
	"hash/crc32"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Map struct {
	hash Hash
	// 虚拟节点数
	virtualNodeCount int
	keys             Int64Slice
	hashMap          map[int64]string
	mu               sync.RWMutex
	nodes            []string // 用于推送  || 确定顺序
}

type Int64Slice []int64

func (x Int64Slice) Len() int           { return len(x) }
func (x Int64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// 一致性算法
type Hash func(data []byte) uint32

func New(count int, hash Hash) *Map {
	m := &Map{
		hash:             hash,
		virtualNodeCount: count,
		hashMap:          make(map[int64]string),
		mu:               sync.RWMutex{},
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 根据环的顺时针来选择命中的节点。
func (m *Map) Get(key string) string {
	m.mu.RLock()
	defer m.mu.Unlock()
	if len(m.keys) == 0 {
		return ""
	}
	hash := int64(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *Map) Set(urls ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, url := range urls {
		urlLocation := int64(m.hash([]byte(url)))
		for m.existLocation(urlLocation) {
			rand.Seed(time.Now().UnixNano())
			urlLocation += rand.Int63n(math.MaxInt32)
			urlLocation %= math.MaxInt32
		}
		for i := 0; i < m.virtualNodeCount; i++ {
			hash := urlLocation + int64(math.MaxUint32/uint32(m.virtualNodeCount)*uint32(i))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = url
		}
	}
	sort.Sort(m.keys)
}

func (m *Map) existLocation(i int64) bool {
	for _, key := range m.keys {
		if key == i {
			return true
		}
	}
	return false
}

func (m *Map) RemoveNodeByUrl(targetUrl string) {
	m.mu.Lock()
	m.mu.Unlock()
	var nums []int64
	for i, url := range m.hashMap {
		if url == targetUrl {
			nums = append(nums, i)
		}
	}
	for _, val := range nums {
		delete(m.hashMap, val)
		for i := range m.keys {
			if m.keys[i] == val {
				m.keys = append(m.keys[:i], m.keys[i+1:]...)
				break
			}
		}
	}
}

func (m *Map) GetUrlsSortByKey() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	urls := make([]string, 0)
	set := make(map[string]struct{}, 0)
	for _, key := range m.keys {
		val := m.hashMap[key]
		if _, ok := set[val]; !ok {
			urls = append(urls, val)
			set[val] = struct{}{}
		}
	}
	return urls
}
