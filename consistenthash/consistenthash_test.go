package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	m := New(5, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	m.Set("1", "4", "8")
	testMap := map[string]string{
		"1":  "1", // 1,11,21,31,41
		"11": "1",
		"24": "4", // 4,14,24,34,44
		"38": "8", //8,18,28,38,48
	}
	for k, v := range testMap {
		if m.Get(k) != v {
			t.Fatal("Hash func()  error")
		}
	}
}
