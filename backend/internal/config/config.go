// Package config 负责从环境变量读取配置。
// 对应 Java 的 application.yml，但遵循 12-factor，把配置全部外置到环境变量。
package config

import (
	"log"
	"os"
	"strconv"

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
