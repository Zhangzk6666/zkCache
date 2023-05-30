package zkcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"zkCache/lru"
	"zkCache/registry"
	"zkCache/singleflight"
	"zkCache/zklog"

	"github.com/sirupsen/logrus"
)

type Key string
type Controller struct {
	name     string
	get      Get
	cache    *synCache
	nodePool *NodePool
	loader   *singleflight.Group

	reqRemoteMap map[Key][]int64
}

var (
	mu          sync.Mutex
	controller  = make(map[string]*Controller)
	serviceName = registry.ServiceName("cache")
)

type Get func(key string) (string, error)

func NewController(name string, maxSize int, get Get, onEvicted lru.OnEvictedFunc) *Controller {
	mu.Lock()
	defer mu.Unlock()
	if _, exist := controller[name]; exist {
		panic("controller name exist")
	}
	c := &Controller{
		name:     name,
		get:      get,
		nodePool: &NodePool{},
		cache:    NewCache(maxSize, onEvicted),
		loader:   &singleflight.Group{},

		reqRemoteMap: make(map[Key][]int64),
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

func (c *Controller) UpdateNodePool(nodes []string) {
	c.nodePool.nodes = nodes
}

func (c *Controller) SetSelfUrl(url string) {
	c.nodePool.url = url
}

func (c *Controller) Get(key string, reqCode int64) ([]byte, error) {

	if key == "" {
		return nil, fmt.Errorf("key not exist")
	}
	if v, ok := c.cache.get(key); ok {
		zklog.Logger.WithFields(logrus.Fields{
			"key": key,
			"msg": "hit...",
		}).Debug()
		return []byte(v), nil
	}
	zklog.Logger.WithFields(logrus.Fields{
		"key": key,
		"msg": "not hit, call load() ...",
	}).Debug()
	val, err := c.load(key, reqCode)
	if err != nil {
		zklog.Logger.WithField("err", err).Warn()
	}
	return val, err
}

func (c *Controller) load(key string, reqCode int64) ([]byte, error) {
	c.nodePool.mu.Lock()
	code := reqCode
	if code == 0 {
		code = time.Now().UnixNano()
	}
	if recordCodeList, exist := c.reqRemoteMap[Key(key)]; exist {
		for _, recordCode := range recordCodeList {
			if recordCode == code {
				c.nodePool.mu.Unlock()
				return nil, errors.New("二次环形访问...")
			}
		}
	}
	c.reqRemoteMap[Key(key)] = append(c.reqRemoteMap[Key(key)], code)
	localNode := 0
	for i := 0; i < len(c.nodePool.nodes); i++ {
		if c.nodePool.url == c.nodePool.nodes[i] {
			localNode = i
			break
		}
	}
	remoteNode := (localNode + 1) % len(c.nodePool.nodes)
	for remoteNode != localNode {
		zklog.Logger.WithFields(logrus.Fields{
			"remoteNode_index":  remoteNode,
			"localNode_index: ": localNode,
		}).Debug()
		remoteUrl := c.nodePool.nodes[remoteNode]
		c.nodePool.mu.Unlock()
		view, err := c.loader.Do(key, code, func() ([]byte, error) {
			zklog.Logger.WithField("remoteUrl", remoteUrl).Debug()
			if value, err := c.getFromPeer(remoteUrl, key, code); err != nil {
				return nil, err
			} else {
				return value, nil
			}
		})

		if err != nil {
			// 忽略错误 继续循环
			zklog.Logger.WithFields(logrus.Fields{
				"msg": "可能是节点突然挂了 || 或者是二次环形访问  ..... ",
				"err": err.Error(),
			}).Warn("Controller request to remote:")
		} else {
			value := ValueResp{}
			json.Unmarshal(view, &value)
			zklog.Logger.WithFields(logrus.Fields{
				"data": value.Data,
			}).Info()
			c.cache.set(key, value.Data)
			return []byte(value.Data), nil
		}
		c.nodePool.mu.Lock()
		remoteNode = (remoteNode + 1) % len(c.nodePool.nodes)
	}
	// 如果所有节点都不存在->访问设定的DB
	data, err := c.getLocalhost(key)

	// 移除http环形访问标志
	for i, recordCode := range c.reqRemoteMap[Key(key)] {
		if code == recordCode {
			c.reqRemoteMap[Key(key)] = append(c.reqRemoteMap[Key(key)][:i], c.reqRemoteMap[Key(key)][i:]...)
			break
		}
	}
	if len(c.reqRemoteMap[Key(key)]) == 0 {
		delete(c.reqRemoteMap, Key(key))
	}

	c.nodePool.mu.Unlock()
	if err != nil {
		zklog.Logger.WithField("err", err).Warn()
		return nil, errors.New("can not find the value by key: " + key)
	}
	return data, nil
}

// {"code":200,"data":"value","msg":"success"}
type ValueResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // value
}

// 向远程发起请求
func (c *Controller) getFromPeer(baseUrl string, key string, code int64) ([]byte, error) {
	if value, err := c.nodePool.Get(baseUrl, c.name, key, code); err != nil {
		return nil, err
	} else {
		return value, nil
	}
}

// 按照设定的规则->search DB
func (c *Controller) getLocalhost(key string) ([]byte, error) {
	zklog.Logger.WithField("msg", "try to search [Data Source]").Debug()
	value, err := c.get(key)
	if err != nil {
		zklog.Logger.WithFields(logrus.Fields{
			"msg": "[Data Source] not hit........",
			"key": key,
			"err": err.Error(),
		}).Warn()
		return nil, err
	}
	zklog.Logger.WithFields(logrus.Fields{
		"msg": "[Data Source] hit........",
		"key": key,
	}).Debug()
	c.cache.set(key, value)
	return []byte(value), nil
}
