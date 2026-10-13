// Command consumer 启动 RabbitMQ 消费者，异步处理图表生成任务。
package main

import (
	"context"
	"log"
	"time"

	"genbi-go-backend/internal/bootstrap"
	"genbi-go-backend/internal/config"
	"genbi-go-backend/internal/event"
	"genbi-go-backend/internal/mq"
	"genbi-go-backend/internal/service"

	"github.com/redis/go-redis/v9"
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

	// 事件发布器（wt-sse）：装配 Redis 并注入 Redis Pub/Sub 实现。
	// 降级策略：consumer 本不强依赖 Redis，故不复用会 Fatal 的 bootstrap.NewRedis，
	// 而是自建客户端并 Ping；连接失败只告警并回退 NoopPublisher，不阻断消费主流程。
	var publisher event.EventPublisher = event.NoopPublisher{}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Printf("[consumer] Redis 连接失败，SSE 事件发布降级为 no-op: %v", err)
		_ = rdb.Close()
	} else {
		log.Println("[consumer] Redis 连接成功，启用 SSE 事件发布")
		publisher = event.NewRedisPublisher(rdb, cfg.ChartEventsChannel)
	}
	pingCancel()

	consumer := mq.NewConsumer(mqClient, chartService, cfg.ChartGenTimeout, publisher)
	log.Println("[consumer] 启动，等待图表生成任务...")
	if err := consumer.Start(); err != nil {
		log.Fatalf("[consumer] 消费失败: %v", err)
	}
}
