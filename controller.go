package zkcache

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Zhangzk6666/zkCache/lru"
	"github.com/Zhangzk6666/zkCache/singleflight"
)

type Controller struct {
	name     string
	get      Get
	cache    *synCache
	nodePool *NodePool
	loader   *singleflight.Group
}

var (
	mu         sync.Mutex
	controller = make(map[string]*Controller)
)

type Get func(key string) (string, error)

func NewController(name string, maxSize int, get Get, onEvicted lru.OnEvictedFunc) *Controller {
	mu.Lock()
	defer mu.Unlock()
	if _, exist := controller[name]; exist {
		panic("controller name exist")
	}
	c := &Controller{
		name:   name,
		get:    get,
		cache:  NewCache(maxSize, onEvicted),
		loader: &singleflight.Group{},
	}
	controller[name] = c
	return c
}

func GetController(name string) (*Controller, bool) {
	mu.Lock()
	defer mu.Unlock()
	if c, ok := controller[name]; ok {
		return c, ok
	}
	return nil, false
}

// topic
func (c *Controller) SetTopic(key string) {
	if _, ok := c.cache.get(key); !ok {
		c.cache.set(key, "")
	}
}

// topic
func (c *Controller) Subscribe(key, value string) {
	c.cache.set(key, value)
}
func (c *Controller) CancelSubscribe(key string) {
	fmt.Println("@@@@@@@@@!!!133311")

	c.cache.remove(key)
}

// get all sub
func (c *Controller) PublicTopicMsg(topicName, msg string) {
	cache := c.cache.getAll()
	fmt.Println("cache len, ", len(cache))
	for k, v := range cache {

		fmt.Println(k, "@", v)
		if v == topicName {
			go func(k string) {
				_, err := http.Get(k + "/topicCall/" + topicName + "/" + msg)
				if err != nil {
					log.Fatal(err.Error())
					return
				}
			}(k)
		}
	}

}
func (c *Controller) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("key not exist")
	}
	if v, ok := c.cache.get(key); ok {
		log.Printf("key:%s,hit ...", key)
		return []byte(v), nil
	}
	log.Printf("key:%s,not hit, load ...", key)
	return c.load(key)
}

func (c *Controller) load(key string) ([]byte, error) {
	viewi, err := c.loader.Do(key, func() ([]byte, error) {
		if c.nodePool != nil {
			if realNode, ok := c.nodePool.PickRealNode(key); ok {
				if value, err := c.getFromPeer(realNode, key); err == nil {
					return value, nil
				}
				log.Println("[zkCache] Failed to get from realNode: " + realNode)
			}
		}
		return c.getLocalhost(key)
	})

	if err == nil {
		return viewi, nil
	}
	return nil, errors.New("can not find the value by key: " + key)
}

func (c *Controller) getFromPeer(baseUrl string, key string) ([]byte, error) {
	if value, err := c.nodePool.Get(baseUrl, c.name, key); err != nil {
		return nil, err
	} else {
		return value, nil
	}
}
func (c *Controller) getLocalhost(key string) ([]byte, error) {
	value, err := c.get(key)
	if err != nil {
		log.Println("search [DB] fail ,not hit ...")
		return nil, err
	}
	log.Println("search [DB] success ,hit ...")
	c.cache.set(key, value)
	return []byte(value), nil

}

func (c *Controller) RegisterPeers(nodePool *NodePool) {
	c.nodePool = nodePool
}
