package zkcache

import (
	"testing"
)

func TestNewSynCache(t *testing.T) {
	cache := NewCache(0, nil)
	cache.set("k", "v")
	if value, ok := cache.get("k"); !ok || value != "v" {
		t.Fatal("error set or get")
	}
}
