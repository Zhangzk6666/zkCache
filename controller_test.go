package zkcache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"1":   "2",
	"11":  "22",
	"111": "222",
}

func TestXxx(t *testing.T) {
	controller := NewController("ss", 0, func(key string) (string, error) {
		log.Println("[SlowDB] search key", key)
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

	/**
		type Controller struct {
	    name     string
	    get      Get
	    cache    *synCache
	    nodePool *NodePool
	    loader   *singleflight.Group
			callBack func(add string, msg string)
	}




	serveHttp
	/topicCall/topic/msg

		**/
	controller := NewController("topic", 0,
		func(key string) (string, error) {
			// call other
			return "", fmt.Errorf("%s not exist", key)
		}, nil)

	// controller.Get() // get  topic and string(checkout exeist things)
	// public just add
	controller.cache.set("1", "2")
	// controller.DoByCallbac(

	// 	func(url string) {

	// 					res, err := http.Get(url)
	// 					if err != nil {
	// 						return nil, err
	// 					}
	// 					defer res.Body.Close()

	// 					if res.StatusCode != http.StatusOK {
	// 						return nil, fmt.Errorf("server returned: %v", res.Status)
	// 					}

	// 					bytes, err := ioutil.ReadAll(res.Body)
	// 					if err != nil {
	// 						return nil, fmt.Errorf("reading response body: %v", err)
	// 					}
	// 					fmt.Println("send success........,res:",res)

	// }, nil)

	// )
	// to call on all sub

	// sub
	// controllerSub := NewController("sub", 0,
	// func(key string) (string, error) {
	// 	// call other
	// 	return "", fmt.Errorf("%s not exist", key)
	// }, nil)

	// controllerSub
	//  addr : topicName
	// to print addr msg
}
