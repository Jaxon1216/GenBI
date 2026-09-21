// Command server 启动 GenBI 的 HTTP 服务（默认 :8080，所有路由挂在 /api 下）。
package main

import (
	"log"

	"genbi-go-backend/internal/bootstrap"
	"genbi-go-backend/internal/config"
	"genbi-go-backend/internal/handler"
	"genbi-go-backend/internal/mq"
	"genbi-go-backend/internal/ratelimit"
	"genbi-go-backend/internal/service"
)

func main() {
	cfg := config.Load()

	// 装配基础设施。
	db := bootstrap.NewDB(cfg)
	rdb := bootstrap.NewRedis(cfg)

	store, err := bootstrap.NewSessionStore(cfg)
	if err != nil {
		log.Fatalf("[server] 创建会话存储失败: %v", err)
	}

	// 装配业务层与处理器。
	userService := service.NewUserService(db)
	userHandler := handler.NewUserHandler(userService)

	deepSeek := service.NewDeepSeekService(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL)
	chartService := service.NewChartService(db, deepSeek)
	limiter := ratelimit.New(rdb)

	// RabbitMQ 生产者（可选）：连接失败不阻塞服务启动，仅 /gen/async 不可用。
	var producer *mq.Producer
	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("[server] RabbitMQ 连接失败，/chart/gen/async 将不可用: %v", err)
	} else {
		defer mqClient.Close()
		producer = mq.NewProducer(mqClient)
	}

	chartHandler := handler.NewChartHandler(chartService, limiter, producer)

	r := bootstrap.NewRouter(bootstrap.RouterDeps{
		Cfg:          cfg,
		Store:        store,
		DB:           db,
		UserHandler:  userHandler,
		ChartHandler: chartHandler,
	})

	addr := ":" + cfg.Port
	log.Printf("[server] 监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[server] 启动失败: %v", err)
	}
}
