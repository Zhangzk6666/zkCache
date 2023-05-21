package main

import (
	"fmt"
	"sync"
	"time"
	"zkCache/registry"
	"zkCache/service"
)

type res struct {
	M    sync.Mutex
	Data map[string]int
}

func main() {
	start := time.Now().UnixMilli()
	dd := &res{
		M:    sync.Mutex{},
		Data: make(map[string]int),
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(dd *res) {
			defer wg.Done()
			dd.M.Lock()
			defer dd.M.Unlock()
			data, err := service.GetService(registry.ServiceName("Test Service"))
			if err != nil {
				fmt.Println(err)
				return
			}
			dd.Data[data] = dd.Data[data] + 1
		}(dd)
	}

	wg.Wait()

	end := time.Now().UnixMilli()
	fmt.Println("耗时", end-start)
	for key, value := range dd.Data {
		fmt.Println(key, ":", value)
	}
}
