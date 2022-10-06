package zkcache

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"
)

var db = map[string]string{
	"1":   "2",
	"11":  "22",
	"111": "222",
}

func TestXxx(t *testing.T) {
	controller := NewController("ss", 0, func(key string) (string, error) {
		log.Println("search key", key)
		if v, ok := db[key]; ok {
			return v, nil
		}
		return "", fmt.Errorf("%s not exist", key)
	}, nil)
	GetController("ss")

	for k, v := range db {
		if view, err := controller.Get(k); err != nil || string(view) != v {
			t.Fatalf("hit fail && call load success,key: %s", k)
		}
		if _, err := controller.Get(k); err != nil {
			t.Fatalf("hit success,key: %s", k)
		}
	}

	if view, err := controller.Get("123"); err == nil {
		t.Fatalf("never hit: %s", view)
	}
}

func TestTopic(t *testing.T) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {

		addr := "http://localhost:9999"
		// topic
		topic := NewController("topic", 0, nil, nil)
		// subscribe
		subscribe := NewController("subscribe", 0, nil, nil)

		time.Sleep(4 * time.Second)
		topic.SetTopic("topicName")
		subscribe.Subscribe(addr, "topicName")
		subscribe.PublicTopicMsg("topicName", "msg")
	}()

	go func() {
		addr := "http://localhost:9999"
		zkCache := NewController("test", 0, nil, nil)
		nodePool := NewNodePool(addr)
		nodePool.Set(addr)
		zkCache.RegisterPeers(nodePool)
		log.Println("zkCache is running at", addr)
		log.Fatal(http.ListenAndServe(addr[7:], nodePool))
	}()

	select {
	case <-stop:
		return
	}
}
