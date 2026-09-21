# 04 · 迁移地图（Java → Go）

面向「学 Go 后端」的对照表：左边是原 Java 组件，右边是本仓库的 Go 等价物。

## 框架 / 分层

| Java (Spring) | Go | 说明 |
|---|---|---|
| `@RestController` + `@RequestMapping` | gin `RouterGroup` + handler 函数 | 路由 |
| `@Resource` 注入 | 显式构造函数传参 | 无 IoC 容器，手动装配（bootstrap） |
| `@Service` | `service` 包普通 struct | 业务层 |
| MyBatis-Plus `ServiceImpl` / `QueryWrapper` | GORM `*gorm.DB` + 链式 `Where/Order` | 数据层 |
| `application.yml` | 环境变量 + `internal/config` | 12-factor 配置 |
| `GlobalExceptionHandler` | gin `Recovery` 中间件 + `common.Error` | 统一错误 |

## 鉴权 / 会话

| Java | Go |
|---|---|
| Spring Session（`HttpSession` + Redis） | `gin-contrib/sessions` + redis store |
| `request.getSession().setAttribute(USER_LOGIN_STATE, user)` | `session.Set("userId", id)` |
| `AuthInterceptor` / `@AuthCheck(mustRole=...)` | `middleware.Auth` / `middleware.Admin` |
| `DigestUtils.md5DigestAsHex((SALT+pwd).getBytes())` | `util.EncryptPassword`（crypto/md5，盐 `tanter`） |

## 异步 / 限流

| Java | Go |
|---|---|
| RabbitMQ `InitMain`（声明 exchange/queue/bind） | `mq` 包启动时 `ExchangeDeclare/QueueDeclare/QueueBind` |
| `MessageProducer.send()` | `mq.Producer.Publish(chartId)` |
| `@RabbitListener(ackMode=MANUAL)` + `channel.basicAck` | `channel.Consume(autoAck=false)` + `d.Ack(false)` |
| `@Retryable(maxAttempts=2, backoff=2000)` + `@Recover` | 消费者内 for 循环重试 2 次 + `time.Sleep(2s)` + 兜底置 failed |
| Redisson `RRateLimiter`（3/s） | `redis_rate.Limiter.Allow(key, PerSecond(3))` |

## 工具 / AI

| Java | Go |
|---|---|
| EasyExcel `excelToCsv` | excelize/v2 读 sheet → CSV（`util/excel.go`） |
| `cleanEchartsJsToJson` | `util.CleanEchartsJSON`（`util/echarts.go`） |
| `DeepSeekServiceImpl`（RestTemplate） | `service/deepseek.go`（net/http） |
| fastjson `JSONObject` | `encoding/json` + `map[string]any` |

## 备选（记录，不实现）

- 会话：JWT（Bearer Token）—— 前后端彻底解耦时的练习项。
- 异步：asynq（仅依赖 Redis）—— 若不想跑 RabbitMQ 时的轻量替代。
- 密码：bcrypt / argon2id —— 生产环境应采用的更安全方案。
