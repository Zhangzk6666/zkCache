package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	zkcache "zkCache"
	"zkCache/pkg/response"
	"zkCache/registry"
	"zkCache/service"
	"zkCache/zklog"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

var db = map[string]string{
	"demo":            "demoValue",
	"zkCache":         "zkCacheValue",
	"game":            "gameValue",
	"test11111111111": "test11111111111Value",
}

func createGroup() *zkcache.Controller {
	return zkcache.NewController("scores", 2<<10, func(key string) (string, error) {
		zklog.Logger.WithField("msg", fmt.Sprintf("[Data Source] search key: %v", key)).Debug()
		if v, ok := db[key]; ok {
			zklog.Logger.WithField("msg", fmt.Sprintf("[Data Source]  search success, key: %v", key)).Debug()
			return v, nil
		}
		zklog.Logger.WithField("msg", fmt.Sprintf("[Data Source]  search failed, key: %v", key)).Debug()
		return "", fmt.Errorf("%s not exist", key)

	}, nil)
}

func getKeyService(router *gin.Engine, controller *zkcache.Controller) {
	router.GET("/api", func(ctx *gin.Context) {
		key, _ := ctx.GetQuery("key")
		code, _ := ctx.GetQuery("code")
		reqCode := com.StrTo(code).MustInt64()
		view, err := controller.Get(key, reqCode)
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
	flag.IntVar(&port, "port", 8881, "zkCache server port")
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
		zklog.Logger.WithField("err", err).Error()
		panic(err)
	}

	<-ctx.Done()
	zklog.Logger.WithField("msg", "shutdown ....").Warn()
}
