package zkcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBaseUrl          = "/_zkCache/"
	defaultVirtualNodeCount = 100
)

type NodePool struct {
	url     string
	coreUrl string
	mu      sync.Mutex
	// coreMap *consistenthash.Map
}

// // 选择真实节点
// func (n *NodePool) PickRealNode(key string) (string, bool) {
// 	n.mu.Lock()
// 	defer n.mu.Unlock()
// 	if peer := n.coreMap.Get(key); peer != "" && peer != n.url {
// 		n.Log("Pick node %s", peer)
// 		return peer, true
// 	}
// 	return "", false
// }

func (h *NodePool) Get(baseUrl string, group string, key string) ([]byte, error) {

	fmt.Println("baseUrl: ", baseUrl, ".............")
	fmt.Println("h.coreUrl: ", h.coreUrl, ".............")
	u := fmt.Sprintf(
		"%v%v%v/%v",
		baseUrl,
		h.coreUrl,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	fmt.Println("u: ", u, ".............")
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

func NewNodePool(url string) *NodePool {
	return &NodePool{
		url:     url,
		coreUrl: defaultBaseUrl,
	}
}

func (n *NodePool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", n.url, fmt.Sprintf(format, v...))
}

func (n *NodePool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// revicer msg by system callBack      mark: /topicCall/
	if strings.HasPrefix(r.URL.Path, "/topicCall/") {
		temp := strings.Replace(n.url, "http://", "", -1)
		temp = strings.Replace(temp, "https://", "", -1)
		if temp == r.Host {
			parts := strings.SplitN(r.URL.Path[len("/topicCall/"):], "/", -1)
			if len(parts) != 2 {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			msg := parts[1]
			// call back print || sub
			fmt.Println(msg)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte(msg + "\n"))
			return
		}
	}

	// search cache
	if !strings.HasPrefix(r.URL.Path, n.coreUrl) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	n.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(n.coreUrl):], "/", -1)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	controllerName := parts[0]
	key := parts[1]

	controller, ok := GetController(controllerName)
	if !ok {
		http.Error(w, "can not found controller", http.StatusNotFound)
		return
	}
	if controller == nil {
		http.Error(w, "no such group: "+controllerName, http.StatusNotFound)
		return
	}

	value, err := controller.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value)
	w.Write([]byte("\n"))
}

// func (n *NodePool) Set(addrs ...string) {
// 	n.mu.Lock()
// 	defer n.mu.Unlock()
// 	n.coreMap = consistenthash.New(defaultVirtualNodeCount, nil)
// 	n.coreMap.Set(addrs...)
// }
