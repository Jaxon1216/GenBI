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
- 拓扑（主队列已换新名以支持死信，见下）：
  - exchange：`BI_exchange`（direct，durable）
  - queue：`BI_Queue.v2`（durable，带 `x-dead-letter-exchange` 参数）
  - routing key：`BI_routingkey`
  - 死信：exchange `BI_exchange.DLX` + queue `BI_Queue.DLX` + routing key `BI_routingkey.DLX`
- 生产者（server 进程内）：`/gen/async` 不再直接发消息，改走 **Outbox**（见下）；由 relay 补发消息体为 `chartId`（`DeliveryMode=Persistent`）。
- 消费者（`cmd/consumer` 独立进程）：
  - `channel.Qos(prefetch=1)`，`Consume(autoAck=false)` → **手动 ack**。
  - 处理：**CAS 抢占** `wait→running` → 调 DeepSeek（**带 ctx 超时**）→ 清洗 ECharts → 置 `succeed`（写 genChart/genResult）或 `failed`（写 execMessage）。
  - **幂等消费**：置 running 用 `UPDATE chart SET status='running' WHERE id=? AND status='wait'`，`RowsAffected==0` 返回哨兵 `ErrAlreadyProcessed`，消费者直接 ack 跳过，重复消息不重复生成。
  - 重试：应用层最多 2 次、间隔 2s（对标 `@Retryable(maxAttempts=2, backoff=2000)`）。
  - **死信兜底**：重试耗尽先 `MarkFailed`（置 `failed` + execMessage）并发 failed 事件，再 `Nack(requeue=false)` 让毒消息经 `x-dead-letter-exchange` 进入死信队列保留排查，而非直接丢弃。

> 换队列名说明：RabbitMQ 不允许修改已存在队列的声明参数，给主队列补死信参数会 `PRECONDITION_FAILED`，故换新名 `BI_Queue.v2` 重新声明（exchange/routingKey 不变，端到端连通）。**已部署环境**：旧 `BI_Queue` 若有积压需人工处理。

## 超时取消（context 传播）

- `ProcessChart(ctx, chartID)` / `GenSync(ctx, ...)` 一路把 ctx 透传到 `DeepSeekService.AnalyzeCsv(ctx, prompt)`，后者用 `http.NewRequestWithContext` 发请求。
- 消费者 `handle` 用 `context.WithTimeout(ctx, CHART_GEN_TIMEOUT)`（默认 30s）包装后传入，使 AI 调用可被**主动打断**，而非仅靠 `http.Client{Timeout:60s}` 兜底（60s 兜底保留作双保险）。
- 超时/取消归一化为 `errors.Is(err, context.DeadlineExceeded/Canceled)`；超时时 `ProcessChart` 返回 execMessage 为「生成超时」的业务错误。

## 事务消息表（Outbox）

- 目的：修复 `/gen/async`「先入库再发消息」的双写不一致窗口（入库成功但发消息失败 → 任务永久 wait）。
- 表 `outbox`（id/aggregateId/payload/status[pending/sent]/retryCount/createTime/updateTime，见 `sql/02_outbox.sql`）。
- `SaveWait` 在**一个 DB 事务**内写入 `chart(wait)` + `outbox(pending)`，二者原子化；`GenAsync` 不再直接 `Publish`。
- Relay（server 进程内后台 goroutine，开关 `OUTBOX_RELAY_ENABLED` 默认开、间隔 `OUTBOX_RELAY_INTERVAL` 默认 5s）：轮询 `pending` → `producer.Publish(payload)` → 成功置 `sent`，失败累加 `retryCount` 下轮重试。producer 为 nil（RabbitMQ 未连通）时不启动。
- **语义边界**：relay 是 **at-least-once**（可能重复投递），依赖上面「幂等消费 CAS」兜底，整体达到 **effectively-once**（终态不重复生成）。

## SSE 实时推送

- 目的：任务状态变更实时推前端，去掉无效轮询；前端优雅降级回轮询。
- 跨进程：`consumer` 进程处理完成/失败后，经 **Redis Pub/Sub** 向频道 `CHART_EVENTS_CHANNEL`（默认 `chart:events`）发布 `{chartId,status}`（实现 `event.EventPublisher`，连接失败降级为 `NoopPublisher`，不阻塞消费）。
- `server` 进程 SSE handler：`GET /api/chart/sse?id=`（走 auth + 校验该 chart 属于当前用户），订阅频道、按 chartId 过滤后用 `text/event-stream` 推送（`Cache-Control:no-cache`、`Connection:keep-alive`、15s 心跳、`c.Request.Context().Done()` 清理）。
- 前端 `useChartEvents`（`EventSource`，同源带 cookie）：收 succeed/failed 刷新该图表；不可用/断连/建连超时则回退现有 `usePolling`（保留未删）。

## DeepSeek

- endpoint：`{DEEPSEEK_BASE_URL}/v1/chat/completions`，model `deepseek-chat`，`temperature=0.1`，`max_tokens=1024`。
- system prompt 沿用 Java（要求输出 `{"genChart": "...", "genResult": [...]}`）。
- 解析 `choices[0].message.content` 为 JSON，取 `genChart`、`genResult`（数组则 `json.Marshal` 成字符串）。
- API key 从环境变量读取，**不硬编码**；缺失或调用失败时返回带 error 的结果，调用方据此置 `failed` / 返回系统错误。

## ECharts 清洗（util/echarts.go）

移植 Java `cleanEchartsJsToJson`：去 `option =`、去尾分号/逗号、单引号转双引号、去换行，再用 `json.Valid` 校验，非法则兜底 `"{}"`。

## 配置项

见 [`.env.example`](../../.env.example)：`PORT / DB_DSN / REDIS_ADDR / REDIS_PASSWORD / RABBITMQ_URL / DEEPSEEK_API_KEY / DEEPSEEK_BASE_URL / SESSION_SECRET / SESSION_MAX_AGE`。

异步可靠性相关（均有默认值，连接失败不阻塞主流程）：
- `CHART_GEN_TIMEOUT`：单次 AI 生成超时（默认 `30s`，支持 `30s`/`2m` 或纯数字秒）。
- `OUTBOX_RELAY_ENABLED`：是否启用 outbox relay（默认 `true`）。
- `OUTBOX_RELAY_INTERVAL`：relay 轮询间隔（默认 `5s`）。
- `CHART_EVENTS_CHANNEL`：SSE 事件的 Redis Pub/Sub 频道名（默认 `chart:events`）。
