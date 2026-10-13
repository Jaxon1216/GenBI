// Command consumer 启动 RabbitMQ 消费者，异步处理图表生成任务。
package main

import (
	"log"

	"genbi-go-backend/internal/bootstrap"
	"genbi-go-backend/internal/config"
	"genbi-go-backend/internal/event"
	"genbi-go-backend/internal/mq"
	"genbi-go-backend/internal/service"
)

func main() {
	cfg := config.Load()

	// 装配基础设施。
	db := bootstrap.NewDB(cfg)

	deepSeek := service.NewDeepSeekService(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL)
	chartService := service.NewChartService(db, deepSeek)

	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("[consumer] RabbitMQ 连接失败: %v", err)
	}
	defer mqClient.Close()

	// 事件发布器：默认 no-op；wt-sse 会在此装配 Redis 并注入 Redis Pub/Sub 实现。
	var publisher event.EventPublisher = event.NoopPublisher{}

	consumer := mq.NewConsumer(mqClient, chartService, cfg.ChartGenTimeout, publisher)
	log.Println("[consumer] 启动，等待图表生成任务...")
	if err := consumer.Start(); err != nil {
		log.Fatalf("[consumer] 消费失败: %v", err)
	}
}
