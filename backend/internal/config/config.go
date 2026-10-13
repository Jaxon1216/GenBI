// Package config 负责从环境变量读取配置。
// 对应 Java 的 application.yml，但遵循 12-factor，把配置全部外置到环境变量。
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 聚合所有运行期配置项。
type Config struct {
	Port string // HTTP 端口，前端 dev 代理写死 8080

	DBDSN string // MySQL DSN（GORM）

	RedisAddr     string
	RedisPassword string

	RabbitMQURL string

	DeepSeekAPIKey  string
	DeepSeekBaseURL string

	SessionSecret string
	SessionMaxAge int // 秒

	// ChartGenTimeout 单次图表 AI 生成的超时上限（wt-timeout 使用）。
	// consumer.handle 用它包一层 context.WithTimeout 后传入 ProcessChart，
	// 使 AI 调用可被主动取消，而非仅依赖 HTTP 客户端的 60s 兜底。默认 30s。
	ChartGenTimeout time.Duration

	// OutboxRelayEnabled 是否启用事务消息表（outbox）的后台投递 relay（wt-outbox 使用）。
	// 默认开启；关闭后 relay goroutine 不启动，GenAsync 的消息将不会被补发。
	OutboxRelayEnabled bool
	// OutboxRelayInterval relay 轮询 outbox 中 pending 记录的间隔（wt-outbox 使用）。默认 5s。
	OutboxRelayInterval time.Duration

	// ChartEventsChannel 图表状态变更事件的 Redis Pub/Sub 频道名（wt-sse 使用）。
	// consumer 进程向此频道 Publish，server 进程 SSE handler 订阅此频道。默认 chart:events。
	ChartEventsChannel string
}

// Load 读取 .env（若存在）后从环境变量装配配置，并给出合理默认值。
func Load() *Config {
	// .env 可选；不存在时忽略错误，直接读进程环境变量。
	if err := godotenv.Load(); err != nil {
		log.Printf("[config] 未找到 .env 文件，使用系统环境变量: %v", err)
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBDSN:           getEnv("DB_DSN", "root:password@tcp(localhost:3306)/genbi?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		SessionSecret:   getEnv("SESSION_SECRET", "please-change-this-secret"),
		SessionMaxAge:   getEnvInt("SESSION_MAX_AGE", 2592000),

		ChartGenTimeout:     getEnvDuration("CHART_GEN_TIMEOUT", 30*time.Second),
		OutboxRelayEnabled:  getEnvBool("OUTBOX_RELAY_ENABLED", true),
		OutboxRelayInterval: getEnvDuration("OUTBOX_RELAY_INTERVAL", 5*time.Second),
		ChartEventsChannel:  getEnv("CHART_EVENTS_CHANNEL", "chart:events"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// getEnvBool 读取布尔配置，支持 1/t/true/0/f/false（strconv.ParseBool 口径）。
func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// getEnvDuration 读取时长配置：优先按 time.ParseDuration 解析（如 "30s"、"2m"），
// 其次按纯数字解析为秒；均失败则用默认值。
func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	return def
}
