# 00 · 项目总览（Overview）

## 目标

保留成熟的前端 [`frontend`](../../../frontend)（React18 + UmiJS Max + Ant Design Pro + ECharts），用 **Go** 重写后端，做到**前端零改动**即可联调。

原 Java 后端 [`Genbi-backend`](../../../../Genbi-backend)（`com.intell_BI_backend`）作为**只读业务参考**。本后端（`GenBI/backend`）已由 Go 重写，替换了原先的 Java 模板后端（`com.yupi.springbootinit`，其 Swagger 曾是前端客户端来源）。

## 范围（Scope）

- ✅ 用户：注册 / 登录 / 注销 / 获取当前登录用户（会话基于 Redis）。
- ✅ 图表：增 / 删 / 查 / 分页 / 我的分页 / AI 同步生成 / AI 异步生成。
- ✅ 基础设施：MySQL(GORM)、Redis(会话+限流)、RabbitMQ(异步)、DeepSeek(AI)。
- ⛔ 不做：ES、腾讯云 COS、微信开放平台、帖子(post)等 Java 模板里遗留但前端未用的模块。

## 三仓关系

| 仓库/目录 | 语言 | 角色 |
|---|---|---|
| `GenBI/frontend` | React/TS | **保留**，联调对象（同仓） |
| `GenBI/backend` | **Go** | **本项目**，重写后的后端（已替换原 Java 模板） |
| `Genbi-backend` / `Eleven-intell_BI` | Java(intell_BI) | 用户实现，**只读业务参考**（同级独立仓库） |

## 技术选型与决策（ADR 摘要）

| 维度 | 选型 | 对应 Java |
|---|---|---|
| Web | Gin | Spring Boot Web |
| ORM | GORM + MySQL | MyBatis-Plus |
| 会话 | Cookie Session + Redis | Spring Session + Redis |
| 密码 | MD5 + 固定盐 `tanter` | 同 |
| 异步 | RabbitMQ（amqp091-go） | RabbitMQ |
| 限流 | redis_rate 令牌桶 | Redisson RRateLimiter |
| Excel | excelize/v2 | EasyExcel |
| AI | DeepSeek Chat Completions | 同 |

- **ADR-1 会话用 Redis Session 而非 JWT**：前端仅用 cookie 会话、需服务端登出。JWT 会破坏“前端零改动”。JWT 作为未来前后端彻底解耦时的练习项。
- **ADR-2 异步用 RabbitMQ**：与 Java 架构 1:1 对照迁移（direct exchange + 手动 ack + 应用层重试）。asynq（仅需 Redis）作为备选记录。
- **ADR-3 密码用 MD5 + 固定盐**：与 Java 完全一致，便于对照学习。**已知局限**：快哈希 + 全局固定盐安全性弱，仅学习/演示可接受；生产应升级 bcrypt/argon2id。

## 前置依赖

- Go 1.24+（本机 1.27.1）
- MySQL 8（建库并执行 [`sql/create_table.sql`](../../sql/create_table.sql)）
- Redis（`brew install redis`）
- RabbitMQ（`brew install rabbitmq`）
