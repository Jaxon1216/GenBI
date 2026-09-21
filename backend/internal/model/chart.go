package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// 图表任务状态常量（对齐前端可识别值：wait/running/succeed/failed）。
const (
	ChartStatusWait    = "wait"
	ChartStatusRunning = "running"
	ChartStatusSucceed = "succeed"
	ChartStatusFailed  = "failed"
)

// Chart 图表实体（表 chart）。字段名采用前端口径：name / execMessage。
type Chart struct {
	ID          int64                 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      int64                 `gorm:"column:userId" json:"userId"`
	Name        string                `gorm:"column:name" json:"name"`
	Goal        string                `gorm:"column:goal" json:"goal"`
	ChartData   string                `gorm:"column:chartData" json:"chartData"`
	ChartType   string                `gorm:"column:chartType" json:"chartType"`
	GenChart    string                `gorm:"column:genChart" json:"genChart"`
	GenResult   string                `gorm:"column:genResult" json:"genResult"`
	Status      string                `gorm:"column:status" json:"status"`
	ExecMessage string                `gorm:"column:execMessage" json:"execMessage"`
	CreateTime  time.Time             `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time             `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
	IsDelete    soft_delete.DeletedAt `gorm:"column:isDelete;softDelete:flag" json:"isDelete"`
}

// TableName 指定表名。
func (Chart) TableName() string { return "chart" }
