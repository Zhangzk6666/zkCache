package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	zkcache "zkCache"
	"zkCache/pkg/response"
	"zkCache/registry"

	"github.com/gin-gonic/gin"
)

// func startCacheServer(addr string, addrs []string, gee *zkcache.Controller) {
// 	nodePool := zkcache.NewNodePool(addr)
// 	nodePool.Set(addrs...)
// 	gee.RegisterPeers(nodePool)
// 	log.Println("zkCache is running at", addr)
// 	log.Fatal(http.ListenAndServe(addr[7:], nodePool))
// }

// 启动服务并注册
func Start(ctx context.Context, host string, port int,
	reg registry.RegistrationVO,
	routerFunc func(router *gin.Engine, controller *zkcache.Controller),
	createGroup func() *zkcache.Controller) (context.Context, error) {
	ctx = startService(ctx, reg.ServiceName, host, port, routerFunc, createGroup)
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, serviceName registry.ServiceName,
	host string, port int, routerFunc func(router *gin.Engine, controller *zkcache.Controller),
	createGroup func() *zkcache.Controller) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	router := gin.New()
	controller := createGroup()
	baseService(router, controller)
	routerFunc(router, controller)
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		Handler:        router,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		// addr := fmt.Sprintf("%s:%d", host, port)
		// nodePool := zkcache.NewNodePool(addr)
		// nodePool.Set(addrs...)
		// gee.RegisterPeers(nodePool)
		log.Println(srv.ListenAndServe())
		err := registry.ShutdownService(serviceName, fmt.Sprintf("http://%s:%d", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()
	go func() {
		log.Printf("%v started. Press use 'Ctrl + c' to stop. \n", serviceName)
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		<-stop
		srv.Shutdown(ctx)
	}()
	return ctx
}

func baseService(router *gin.Engine, controller *zkcache.Controller) {
	router.GET("/healthy", func(ctx *gin.Context) {
		response.ResponseMsg.SuccessResponse(ctx, nil)
	})
	router.GET("/updateNodePool", func(ctx *gin.Context) {
		urls := NodePoolMsg{}
		if err := ctx.ShouldBindJSON(&urls); err != nil {
			log.Println(err)
		}
		log.Println("urls.....", urls.Urls)
		controller.UpdateNodePool(urls.Urls)
		response.ResponseMsg.SuccessResponse(ctx, nil)
	})
}

type NodePoolMsg struct {
	Urls []string `json:"urls"`
}

// 获取服务
func GetService(sericeName registry.ServiceName) (string, error) {
	reqUrl := fmt.Sprintf("%s?serviceName=%s", registry.ServiceURL, url.QueryEscape(string(sericeName)))
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, reqUrl, nil)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err

	}
	msg := MSG{}
	json.Unmarshal(body, &msg)
	if msg.Code == response.SUCCESS {
		return msg.Data.URL, nil
	}
	return "", response.NewErr(response.ERROR)
}

// {"code":200,"data":{"url":"http://localhost:8111"},"msg":"success"}
type MSG struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data URL    `json:"data"`
}
type URL struct {
	URL string `json:"url"`
}
