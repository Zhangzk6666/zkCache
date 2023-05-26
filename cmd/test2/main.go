package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	zkcache "zkCache"
	"zkCache/pkg/response"
	"zkCache/registry"
	"zkCache/service"
	"zkCache/zklog"

	"github.com/gin-gonic/gin"
)

var db = map[string]string{
	"demo":            "demoValue",
	"zkCache":         "zkCacheValue",
	"game":            "gameValue",
	"test11111111111": "test11111111111Value",
}

func createGroup() *zkcache.Controller {
	return zkcache.NewController("scores", 2<<10, func(key string) (string, error) {
		log.Println("[DB] search key", key)
		if v, ok := db[key]; ok {
			log.Println("[DB] search success", key)
			return v, nil
		}
		log.Println("[DB] search failed", key)

		return "", fmt.Errorf("%s not exist", key)

	}, nil)
}

func getKeyService(router *gin.Engine, controller *zkcache.Controller) {
	// gee := createGroup()
	router.GET("/api", func(ctx *gin.Context) {
		key, _ := ctx.GetQuery("key")
		view, err := controller.Get(key)
		if err != nil {
			response.ResponseMsg.FailResponse(ctx, response.NewErrWithMsg(http.StatusInternalServerError, "服务器错误!"), nil)
			return
		}
		response.ResponseMsg.SuccessResponse(ctx, string(view))
	})

}

var serviceName registry.ServiceName

func main() {

	var port int
	var api bool
	var topicName, host string
	flag.IntVar(&port, "port", 8883, "zkCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.StringVar(&topicName, "topicName", "cache", "Start a api server?")
	flag.StringVar(&host, "host", "localhost", "Start a api server?")
	flag.Parse()
	serviceName = registry.ServiceName(topicName)

	reg := registry.RegistrationVO{
		ServiceName: serviceName,
		ServiceURL:  fmt.Sprintf("http://%s:%d", host, port),
	}

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		reg,
		getKeyService,
		createGroup,
	)

	if err != nil {
		zklog.Logger.Error(err)
		panic(err)
	}

	<-ctx.Done()
	fmt.Println("shutdown ....")
}
