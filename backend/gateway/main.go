package main

import (
	"backend/gateway/internal/client"
	"backend/gateway/internal/handler"
	"backend/gateway/router"
	"backend/pkg/config"
	"backend/pkg/logger"
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig("../config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	logger.Init(cfg.Logger)

	// 初始化 user 服务 gRPC 客户端
	userClient, err := client.NewUserClient(cfg.Services.User)
	if err != nil {
		log.Fatalf("连接 user 服务失败: %v", err)
	}

	r := gin.New()
	router.SetRouter(r, handler.NewHandler(userClient))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
