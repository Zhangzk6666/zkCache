package lru

type OnEvictedFunc func(key string, value string)
