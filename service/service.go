package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	zkcache "zkCache"
	"zkCache/pkg/response"
	"zkCache/registry"
	"zkCache/zklog"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
	controller.SetSelfUrl(fmt.Sprintf("http://%s:%d", host, port))
	baseService(router, controller)
	routerFunc(router, controller)
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		Handler:        router,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		zklog.Logger.WithField("msg", srv.ListenAndServe()).Warn()
		err := registry.ShutdownService(serviceName, fmt.Sprintf("http://%s:%d", host, port))
		if err != nil {
			zklog.Logger.WithField("err", err).Error()
		}
		cancel()
	}()
	go func() {
		zklog.Logger.WithField("msg",
			fmt.Sprintf("%v started. Press use 'Ctrl + c' to stop.", serviceName)).Info()
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
			zklog.Logger.WithField("err", err).Error()
		}
		zklog.Logger.WithFields(logrus.Fields{
			"urls": strings.Join(urls.Urls, ","),
		}).Debug()
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
		zklog.Logger.WithField("err", err).Error()
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
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
