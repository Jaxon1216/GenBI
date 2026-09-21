// Package bootstrap 负责装配全局基础设施：MySQL(GORM)、Redis、RabbitMQ、会话、路由。
// 无 IoC 容器，一切显式构造并注入（对照 Java 的 @Resource 自动注入）。
package bootstrap

import (
	"log"

	"genbi-go-backend/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB 连接 MySQL 并返回 *gorm.DB。连接失败直接 fatal（服务无法工作）。
func NewDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("[bootstrap] 连接 MySQL 失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[bootstrap] 获取底层 sql.DB 失败: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("[bootstrap] MySQL Ping 失败: %v", err)
	}
	log.Println("[bootstrap] MySQL 连接成功")
	return db
}
