package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	zkcache "zkCache"
	"zkCache/registry"
	"zkCache/zklog"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

func startAPIServer(apiAddr string, gee *zkcache.Controller) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			code := r.URL.Query().Get("code")
			reqCode := com.StrTo(code).MustInt64()
			view, err := gee.Get(key, reqCode)
			// view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			// w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(view))

		}))
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	router := gin.New()
	registry.RegisterHandlers(router)
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", registry.ServiceHost, registry.ServicePort),
		Handler:        router,
		MaxHeaderBytes: 1 << 20,
	}
	go registry.Heartbeat(5 * time.Second)
	go func() {
		zklog.Logger.WithField("msg", srv.ListenAndServe()).Warn()
		zklog.Logger.WithField("msg", "注册中心退出").Warn()
		cancel()
	}()
	go func() {
		zklog.Logger.WithField("msg", "注册中心 started. Press use 'Ctrl + c' to stop.").Info()
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		<-c
		srv.Shutdown(ctx)
		cancel()
	}()
	<-ctx.Done()
	zklog.Logger.WithField("msg", "shutdown ....").Warn()
}
