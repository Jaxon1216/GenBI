# 02 · 数据模型（Data Model）

建库脚本见 [`sql/create_table.sql`](../../sql/create_table.sql)。字段命名采用**前端口径**，与前端 JSON key 完全一致，避免中间层再做映射。

## user 表

| 列名 | 类型 | 说明 |
|---|---|---|
| id | bigint PK auto_increment | 主键 |
| userAccount | varchar(256) | 账号，唯一（索引） |
| userPassword | varchar(512) | 密码，`md5Hex("tanter"+明文)` |
| userName | varchar(256) | 昵称 |
| userAvatar | varchar(1024) | 头像 URL |
| userRole | varchar(256) default 'user' | 角色：user/admin/ban |
| createTime | datetime | 创建时间 |
| updateTime | datetime | 更新时间 |
| isDelete | tinyint default 0 | 逻辑删除 |

## chart 表

| 列名 | 类型 | 说明 |
|---|---|---|
| id | bigint PK auto_increment | 主键 |
| userId | bigint | 创建者 id（索引） |
| name | varchar(256) | 图表名称（**前端口径**，Java 为 chartName） |
| goal | text | 分析目标 |
| chartData | text | 原始数据（CSV） |
| chartType | varchar(256) | 图表类型 |
| genChart | text | 生成的 ECharts JSON |
| genResult | text | 生成的分析结论 |
| status | varchar(256) default 'wait' | wait/running/succeed/failed |
| execMessage | text | 执行信息（**前端口径**，Java 为 Message） |
| createTime | datetime | 创建时间 |
| updateTime | datetime | 更新时间 |
| isDelete | tinyint default 0 | 逻辑删除 |

## GORM struct ↔ 列 ↔ JSON key 映射

GORM 实体用 `gorm:"column:xxx"` 精确指定列名（避免 GORM 默认 snake_case 转换），JSON tag 对齐前端 typings。

```go
type User struct {
    ID           int64     `gorm:"column:id;primaryKey" json:"id"`
    UserAccount  string    `gorm:"column:userAccount" json:"userAccount"`
    UserPassword string    `gorm:"column:userPassword" json:"-"` // 不下发
    UserName     string    `gorm:"column:userName" json:"userName"`
    UserAvatar   string    `gorm:"column:userAvatar" json:"userAvatar"`
    UserRole     string    `gorm:"column:userRole" json:"userRole"`
    CreateTime   time.Time `gorm:"column:createTime" json:"createTime"`
    UpdateTime   time.Time `gorm:"column:updateTime" json:"updateTime"`
    IsDelete     soft_delete.DeletedAt `gorm:"column:isDelete;softDelete:flag" json:"isDelete"`
}
```

`TableName()` 分别返回 `"user"` / `"chart"`。

## 逻辑删除

用 `gorm.io/plugin/soft_delete` 的 flag 模式映射 `isDelete`（0=未删，1=已删），查询默认自动追加 `isDelete = 0`，`Delete` 时置 1（与 MyBatis-Plus `@TableLogic` 行为等价）。
