// Package model 定义 GORM 实体。
// 列名用 gorm:"column:xxx" 精确指定（禁用默认 snake_case 转换），与建表脚本一致；
// JSON tag 对齐前端 typings.d.ts。
package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// User 用户实体（表 user）。
type User struct {
	ID           int64                 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserAccount  string                `gorm:"column:userAccount" json:"userAccount"`
	UserPassword string                `gorm:"column:userPassword" json:"-"` // 密码不下发
	UserName     string                `gorm:"column:userName" json:"userName"`
	UserAvatar   string                `gorm:"column:userAvatar" json:"userAvatar"`
	UserRole     string                `gorm:"column:userRole" json:"userRole"`
	CreateTime   time.Time             `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time             `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
	IsDelete     soft_delete.DeletedAt `gorm:"column:isDelete;softDelete:flag" json:"isDelete"`
}

// TableName 指定表名。
func (User) TableName() string { return "user" }
