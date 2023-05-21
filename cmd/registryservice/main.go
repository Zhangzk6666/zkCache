package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	zkcache "zkCache"
	"zkCache/registry"

	"github.com/gin-gonic/gin"
)

func startAPIServer(apiAddr string, gee *zkcache.Controller) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
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
		log.Println(srv.ListenAndServe())
		log.Println("注册中心退出")
		cancel()
	}()
	go func() {
		log.Println("注册中心 started. Press use 'Ctrl + c' to stop.")
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		<-c
		srv.Shutdown(ctx)
		cancel()
	}()
	<-ctx.Done()
	fmt.Println("shutdown ....")
}
