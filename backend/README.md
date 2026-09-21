# genbi-go-backend

用 Go 重写的 GenBI 后端，与同仓前端 [`frontend`](../frontend) 完全对接（前端零改动）。
原 Java 后端 [`Genbi-backend`](../../Genbi-backend) 作为只读业务参考。

技术栈：**Gin + GORM + MySQL + Redis(会话/限流) + RabbitMQ(异步) + DeepSeek(AI)**。

详细设计见 [`docs/spec/`](docs/spec)，变更记录见 [`CHANGELOG.md`](CHANGELOG.md)。

## 前置依赖

- Go 1.24+
- MySQL 8
- Redis：`brew install redis && brew services start redis`
- RabbitMQ：`brew install rabbitmq && brew services start rabbitmq`

## 快速开始

```bash
# 1. 建库并建表
mysql -uroot -p < sql/create_table.sql

# 2. 配置
cp .env.example .env    # 按需修改 DB_DSN / DEEPSEEK_API_KEY 等

# 3. 启动 HTTP 服务（:8080）
make run

# 4. 启动异步消费者（另开终端）
make consumer
```

## 与前端联调

```bash
cd ../frontend
npm install
npm run dev     # dev 代理把 /api 转发到 http://localhost:8080
```

打开 http://localhost:8000 ，完成 登录 → 上传数据生成图表 → 我的图表 等流程。

## 常用命令

| 命令 | 说明 |
|---|---|
| `make run` | 启动 HTTP 服务 |
| `make consumer` | 启动 RabbitMQ 消费者 |
| `make build` | 编译到 `bin/` |
| `make vet` | 静态检查 |
| `make test` | 运行单测 |

## 目录结构

```
cmd/server      HTTP 服务入口
cmd/consumer    RabbitMQ 消费者入口
internal/
  config        环境变量配置
  bootstrap     DB/Redis/MQ/Session/Router 装配
  common        统一响应、错误码、分页
  model         GORM 实体：User / Chart
  dto           请求/响应结构体
  middleware    recover / logger / auth / admin
  handler       gin handler：user / chart
  service       业务：user / chart / deepseek
  mq            RabbitMQ 连接/生产/消费
  ratelimit     redis_rate 限流封装
  util          excel→csv / echarts 清洗 / md5 密码
docs/spec       设计文档
sql             建表脚本
```
