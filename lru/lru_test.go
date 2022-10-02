package lru

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Set("key1", "value1")
	lru.Set("key2", "value2")
	var key string

	key = "key1"
	if value, ok := lru.Get(key); !ok || value != "value1" {
		t.Fatal("cache hit ", key, ":", value, " failed")
	}

	key = "key2"
	if value, ok := lru.Get(key); !ok || value != "value2" {
		t.Fatal("cache hit ", key, ":", value, " failed")
	}

	key = "miss key"
	if _, ok := lru.Get(key); ok {
		t.Fatal("why can I get ", key)
	}
}

func TestRemoveBack(t *testing.T) {
	k := []string{"key1", "key2", "k3"}
	v := []string{"value1", "value2", "v3"}
	lru := New(len(k[0]+v[0]+k[1]+v[1]), nil)
	for i := 0; i < len(k); i++ {
		lru.Set(k[i], v[i])
	}

	if _, ok := lru.Get(k[0]); ok || lru.Len() != 2 || lru.list.Front().Value.(*entry).value != v[len(k)-1] {
		t.Fatalf("Remove error || check lru.Len() || check list.Front() when get or set or update")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value string) {
		keys = append(keys, key)
		fmt.Println("c.OnEvicted run......")
		fmt.Println("keys append: ", key)
	}
	k := []string{"key1", "key2"}
	v := []string{"value1", "value2"}
	lru := New(len(k[0]+v[0]), callback)
	for i := 0; i < len(k); i++ {
		lru.Set(k[i], v[i])
	}

	if !reflect.DeepEqual([]string{k[0]}, keys) {
		t.Fatal("check keys", keys)
	}
}
