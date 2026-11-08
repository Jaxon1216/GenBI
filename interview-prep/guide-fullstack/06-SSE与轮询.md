# C06 从生成结果到页面：SSE 与轮询

> 课程性质：核心  
> 返回：[主导学](../GenBI/导学.md)；前置：[C05 消费与取消](05-消费幂等与取消.md)、[C02 会话与契约](02-会话与接口契约.md)；下一课：[C07 图表渲染](07-图表契约与容错.md)。  
> 对应：B4；[Q4 SSE 与轮询](../GenBI/面经.md#q4-sse-与轮询)；目标掌握：看懂链路、理解机制、设计与表达。

## 本课重点

用户打开“我的图表”，看到图表 101 正在生成。数据库保存事实状态，Redis Pub/Sub 转发通知，SSE 把通知送到浏览器，页面收到事件后重新查列表。事件不携带完整图表，也不代替查询结果。

当前先启用轮询，SSE 打开后关闭轮询；断连或建连超时再开启轮询。这个策略有即时性优势，但“连接已打开”不能证明所有事件都送达，订阅前完成的任务可能漏掉通知。

## 主链与职责

```text
consumer 写 succeed / 重试耗尽路径写 failed
 → RedisPublisher 发布 {chartId, status}
 → server SSEHandler 按图表 ID 过滤 → data: JSON\n\n
 → 浏览器 EventSource.onmessage
 → loadData(true) 查询当前页 → React 更新卡片 → 终态关闭连接

页面 pending && sseFallback → usePolling → 查询同一个列表接口
```

SSE 是 server 到浏览器的单向传输。Redis Pub/Sub 连接两个独立 Go 进程；RabbitMQ 负责生成任务投递，两个消息通道承担不同职责。

## 源码阅读

1. [consumer](../../backend/internal/mq/consumer.go)的 `handle`、`onExhausted`：成功写库后发 succeed；耗尽路径调用 MarkFailed 后发 failed。没有发 running 事件；失败写库方法也没有把更新错误返回给此处。
2. [消费者入口](../../backend/cmd/consumer/main.go)：Redis 不可用时注入 NoopPublisher，生成仍可继续。
3. [RedisPublisher](../../backend/internal/event/redis_publisher.go)：Publish 错误只记日志，通知为尽力而为。
4. [SSEHandler.Stream](../../backend/internal/handler/sse_handler.go)：鉴权、读取归属、设置流式响应头，再订阅频道；每 15 秒写注释心跳，收到终态结束流。
5. [useChartEvents](../../frontend/src/hooks/useChartEvents.ts)：Map 持有每个 pending ID 的 EventSource 与定时器；回调用 ref 保持最新引用；默认 10 秒建连超时。
6. [我的图表](../../frontend/src/pages/myChart/index.tsx)：pending ID 来自当前页，默认 `sseFallback=true`，事件回调仅触发 `loadData(true)`；[usePolling](../../frontend/src/hooks/usePolling.ts)管理下一次查询与可见性事件。

决定两条前端路径的真实条件：

```tsx
usePolling(pollCallback, {
  enabled: hasPendingCharts && sseFallback,
  baseInterval: 3000,
  maxInterval: 30000,
});
// useChartEvents 中：
es.onopen = () => onFallbackRef.current?.(false);
// onerror 关闭本条连接，再通知 onFallback(true)
```

轮询快照由 `id:status` 拼接。状态有变化时恢复基础间隔；快照相同时乘以默认 1.5，最高 30 秒。图表名称或结果字段变化不会单独改变这个快照。

## 机制与边界

正常路径是先查询到 wait/running，再建立 SSE；任务之后完成，通知触发静默刷新，刷新读取完整结果，卡片显示 succeed。终态关闭连接，pending 清空后停止刷新。

**丢失窗口**发生在“列表返回 running”与“Redis 订阅生效”之间。如果此时消费者写 succeed 并发布事件，后来的订阅没有回放。SSE 建连仍成功，onopen 把轮询关闭，页面可能一直保留 running。服务端没有终态初始快照、事件 ID、Last-Event-ID 或历史事件存储；心跳只证明连接有字节流，不证明任务仍在运行。

多图表连接共用一个 fallback 布尔值。一条失败设 true，另一条成功设 false，缺少按连接统计的健康状态。hook 的 effect cleanup 在 ID 集合改变时关闭全部连接，所以即使 Map 用于去重，集合变化仍会重建剩余连接。onerror 主动 close 后，本 hook 没有持续重建同一 ID 的恢复定时器；通常靠轮询与后续依赖变化推进。

资源清理有明确位置：后端请求 Context 取消后退出，defer 关闭 Redis 订阅和 ticker；前端 effect 清理 EventSource 与建连定时器。轮询按“请求完成后安排下次”避免普通周期重叠，但 visibilitychange 的立即查询可与已发请求重叠，且没有 AbortController。初始页面若已经 hidden，也没有在每次 schedule 时判断可见性；清理后旧 callback 返回仍可能再次 schedule，因此“后台一定没有请求”超出了源码保证。

[交互专题：比较正常通知与订阅前完成](topics/C06-sse.html)展示为什么打开 SSE 与拿到最新状态需要分别确认。

## 需要懂什么

通知是“现在值得查一次”的信号，数据库查询是实际状态来源。尽力而为通知不应成为唯一完成通道；事件丢失后，周期对账才能让页面收敛。SSE 与轮询可以协作，但将轮询完全关闭需要额外的初始状态补齐和连接健康契约。

若重新设计，可先订阅确认、再读当前状态并发送终态快照，客户端保留低频对账；要防止快照与事件乱序，还应带版本并拒绝旧状态覆盖。这个方案尚未实现，不能把专题里的当前丢失窗口描述成已被修好。

ref 用于承接最新回调，effect 依赖控制资源生命周期，两者不等价。分页列表的请求竞态另见[C08](08-列表状态与数据预览.md)，需要按查询身份接受响应，才能避免旧请求覆盖新搜索。

## 需要记住什么

- 主链：消费者更新数据库 → Redis 通知 → SSE → 浏览器重新查列表。
- 当前只推终态；心跳 15 秒，建连超时 10 秒；轮询基础 3 秒、退避倍数 1.5、上限 30 秒。
- 断连能触发兜底；健康连接漏事件仍是当前限制，多个连接健康状态也未聚合。
- 终态、依赖变化和卸载均要清理资源；暂停定时器不代表取消已发请求。回到[Q4](../GenBI/面经.md#q4-sse-与轮询)复述取舍。
