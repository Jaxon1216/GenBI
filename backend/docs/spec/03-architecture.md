# 03 · 架构（Architecture）

## 分层

```
HTTP 请求
  → middleware（recover / logger / auth / admin）
  → handler（解析入参、调用 service、封装响应）
  → service（业务逻辑：user / chart / deepseek）
  → model（GORM，直接操作 MySQL）
```

学习期保持 moderate 分层，不额外抽 repository 层。`service` 直接持有 `*gorm.DB`。

## 请求生命周期（gin）

1. `bootstrap` 装配全局单例：`*gorm.DB`、`*redis.Client`、RabbitMQ 连接、session store。
2. 路由组 `/api` 挂载 user、chart、health。
3. 需要登录的路由套 `auth` 中间件；需要管理员的套 `admin`。
4. handler 用 `common.Success/Error` 统一封装返回。

## 会话（Cookie Session + Redis）

- 库：`gin-contrib/sessions` + `gin-contrib/sessions/redis`。
- Cookie：名 `GENBI_SESSION`，`HttpOnly=true`，`SameSite=Lax`，`MaxAge=SESSION_MAX_AGE`（默认 2592000 秒 / 30 天）。
- 登录成功：`session.Set("userId", user.ID)` 并 `session.Save()`。
- `auth` 中间件：从 session 取 `userId` → 查库 → 注入 `*model.User` 到 `gin.Context`（key `loginUser`）；缺失返回 `40100`。
- `admin` 中间件：在 auth 之后校验 `userRole=="admin"`，否则 `40101`。
- 注销：`session.Clear()` + `session.Save()`。

## 限流（redis_rate）

- 库：`github.com/go-redis/redis_rate/v10` + `github.com/redis/go-redis/v9`。
- key：`gen_chart:{userId}`，速率 `redis_rate.PerSecond(3)`（对齐 Redisson 每秒 3 次）。
- 超限：返回 `42900`。

## 异步（RabbitMQ）

- 库：`github.com/rabbitmq/amqp091-go`。
- 拓扑（与 Java `Constant` 一致）：
  - exchange：`BI_exchange`（direct，durable）
  - queue：`BI_Queue`（durable）
  - routing key：`BI_routingkey`
- 生产者（server 进程内）：`/gen/async` 先入库 `status=wait`，再 `Publish` 消息体为 `chartId`（`DeliveryMode=Persistent`）。
- 消费者（`cmd/consumer` 独立进程）：
  - `channel.Qos(prefetch=1)`，`Consume(autoAck=false)` → **手动 ack**。
  - 处理：置 `running` → 调 DeepSeek → 清洗 ECharts → 置 `succeed`（写 genChart/genResult）或 `failed`（写 execMessage）。
  - 重试：应用层最多 2 次、间隔 2s（对标 `@Retryable(maxAttempts=2, backoff=2000)`）；耗尽后置 `failed`（对标 `@Recover`）。
  - 无论成败，最终 `Ack` 移出队列（对齐 Java 手动 ack 逻辑）。

## DeepSeek

- endpoint：`{DEEPSEEK_BASE_URL}/v1/chat/completions`，model `deepseek-chat`，`temperature=0.1`，`max_tokens=1024`。
- system prompt 沿用 Java（要求输出 `{"genChart": "...", "genResult": [...]}`）。
- 解析 `choices[0].message.content` 为 JSON，取 `genChart`、`genResult`（数组则 `json.Marshal` 成字符串）。
- API key 从环境变量读取，**不硬编码**；缺失或调用失败时返回带 error 的结果，调用方据此置 `failed` / 返回系统错误。

## ECharts 清洗（util/echarts.go）

移植 Java `cleanEchartsJsToJson`：去 `option =`、去尾分号/逗号、单引号转双引号、去换行，再用 `json.Valid` 校验，非法则兜底 `"{}"`。

## 配置项

见 [`.env.example`](../../.env.example)：`PORT / DB_DSN / REDIS_ADDR / REDIS_PASSWORD / RABBITMQ_URL / DEEPSEEK_API_KEY / DEEPSEEK_BASE_URL / SESSION_SECRET / SESSION_MAX_AGE`。
