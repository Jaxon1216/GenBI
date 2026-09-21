# Changelog

本项目 changelog 遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 风格，
按实施阶段（Phase）增量记录。

## [Unreleased]

### Phase 0 — Spec 与脚手架
#### Added
- 初始化 Go module `genbi-go-backend`（go 1.24）。
- 目录骨架：`cmd/server`、`cmd/consumer`、`internal/{config,bootstrap,common,model,dto,middleware,handler,service,mq,ratelimit,util}`。
- 配置层 `internal/config`：从环境变量 + `.env` 读取配置。
- Spec 文档 `docs/spec/00~04`：总览、API 契约、数据模型、架构、Java→Go 迁移地图。
- 建表脚本 `sql/create_table.sql`（前端口径字段：`name` / `execMessage`）。
- 工程文件：`.env.example`、`.gitignore`、`Makefile`、`README.md`、本 `CHANGELOG.md`。

### Phase 1 — 基础设施
#### Added
- `internal/common`：统一响应 `BaseResponse`、错误码 `ErrorCode`、业务异常 `BusinessError`、泛型分页 `Page[T]`。
- `internal/model`：GORM 实体 `User`、`Chart`（列名精确映射，逻辑删除 `isDelete` 用 soft_delete flag）。
- `internal/bootstrap`：MySQL(GORM)、Redis、RabbitMQ 客户端与拓扑声明、Redis 会话存储、gin 路由装配。
- `internal/mq`：RabbitMQ 连接与拓扑（`BI_exchange`/`BI_Queue`/`BI_routingkey`，与 Java 一致）。
- `internal/middleware`：`Recovery`、`Logger`。
- `internal/util/password.go`：MD5+固定盐密码加密（盐 `tanter`，与 Java 一致）。
- `internal/handler/health.go` + 路由 `GET /api/health`。
- `cmd/server`：装配基础设施并启动 gin（:8080）。
#### Verified
- MySQL(genbi 库) + Redis 连接成功；`GET /api/health` 返回 `{"code":0,"data":"ok","message":"ok"}`。

### Phase 2 — 用户模块
#### Added
- `internal/dto`：用户请求体（register/login/add/update/updateMy/query/delete）与视图 `LoginUserVO`/`UserVO`。
- `internal/service/user_service.go`：注册（校验+MD5 加密+查重）、登录、按 id 查询、增删改、分页；VO 转换与 `IsAdmin`。
- `internal/common/constant.go`：会话/角色常量。
- `internal/middleware/auth.go`：`Auth`（会话→查库→注入登录用户，缺失 40100）、`Admin`（非管理员 40101）、`GetLoginUser`。
- `internal/handler`：`helper.go`（统一成功/失败/错误码/解析 id）、`user_handler.go`（12 个接口）。
- 路由：`/api/user/{register,login,logout,get/login,update/my,get/vo,list/page/vo}`（登录态）与管理员组 `{add,delete,update,get,list/page}`。
#### Verified
- 注册→登录(下发 GENBI_SESSION cookie)→带 cookie 取 get/login→logout→再取 get/login 返回 40100。
- 注册校验：账号重复、两次密码不一致均返回 40000。
- 管理员守卫：普通用户访问 `/user/list/page` 返回 40101。

### Phase 3 — 图表 CRUD
#### Added
- `internal/dto/chart.go`：`ChartAddRequest`/`ChartEditRequest`/`ChartUpdateRequest`/`ChartQueryRequest`/`GenChartByAiRequest`/`BiResponse`（`genChart` 纠偏）。
- `internal/service/chart_service.go`：新增、按 id 查询、删除(本人/管理员)、编辑(本人)、更新(管理员)、分页(size≤20，默认按 createTime 倒序)；暴露 `DB()` 供后续生成流程使用。
- `internal/handler/chart_handler.go`：add/delete/get/list/page/my/list/page/edit/update 处理器。
- 路由：`/api/chart/{add,delete,get,list/page,my/list/page,edit}`（登录态）与 `/update`（管理员）。
#### Verified
- add→my/list/page 返回 `records`/`total`，字段为前端口径 `name`/`execMessage`。
- get?id= 正常；list/page 支持 name 模糊过滤。
- 归属校验：非本人删除返回 40101；本人删除成功；逻辑删除后 get 返回 40400。

### Phase 4 — 同步 AI 生成
#### Added
- `internal/util/echarts.go`：`CleanEchartsJSON` 清洗 AI 返回的 ECharts 配置为合法 JSON（移植 Java cleanEchartsJsToJson）。
- `internal/util/excel.go`：`ExcelToCSV` 支持 `.xlsx`（excelize 读首个 sheet）与 `.csv`；不支持老式 `.xls`。
- `internal/service/deepseek_service.go`：调用 DeepSeek Chat Completions（deepseek-chat/temperature=0.1/max_tokens=1024），解析 content 为 `{genChart, genResult}`。
- `internal/ratelimit/limiter.go`：redis_rate 令牌桶封装（每秒 3 次）。
- `internal/service/chart_service.go`：`BuildPrompt` + `GenSync`（AI→清洗→入库→返回 BiResponse，status=succeed）。
- `internal/handler/chart_handler.go`：`Gen` + `parseGenRequest`（multipart 解析、文件与参数校验、限流）。
- 路由：`POST /api/chart/gen`（登录+限流）。
#### Verified
- 文件校验：非 xlsx/csv 返回 40000；缺失 goal 返回 40000。
- 无 API key 时链路可达 AI 步骤，返回 50000（证明 上传→解析→限流→AI 全链路打通）。
- 限流：连发 6 次，第 4/5 次返回 42900，令牌恢复后第 6 次放行。

### Phase 5 — 异步生成（RabbitMQ）
#### Added
- `internal/mq/producer.go`：`Producer.Publish` 向 `BI_exchange` 持久化发布 chartId。
- `internal/mq/consumer.go`：`Consumer.Start`（Qos 预取 1、手动 ack）+ 应用层重试 2 次/退避 2s + 兜底置 failed（对标 `@Retryable/@Recover`）。
- `internal/service/chart_service.go`：`SaveWait`（入库 wait）、`ProcessChart`（running→AI→清洗→succeed）、`MarkFailed`。
- `internal/handler/chart_handler.go`：`GenAsync`（入库 wait→发消息→返回 chartId）。
- `cmd/consumer/main.go`：装配并启动消费者。
- 路由：`POST /api/chart/gen/async`（登录+限流）；server 侧 RabbitMQ 连接可选（失败不阻塞启动）。
#### Verified
- 启动 RabbitMQ 容器；server+consumer 均连接成功并声明拓扑。
- 提交 gen/async 返回 chartId；状态由 `wait`→`running`；无 AI key 时经 2 次重试后置 `failed` 并写入 execMessage；消息正确 ack 未重投。

### Phase 6 — 收尾与文档
#### Added
- 单元测试：`internal/util`（密码 MD5+盐、ECharts 清洗、CSV 转义/空行判断）、`internal/service`（BuildPrompt、buildOrder 防注入、normalizePage）。
- SHOULD 级用户接口已在 Phase 2 一并实现（add/delete/update/update-my/get/get-vo/list-page/list-page-vo）。
#### Verified
- `go build ./...`、`go vet ./...` 通过；`gofmt -l .` 无输出；`go test ./...` 全绿。
- 全量 MUST 接口端到端联通：health / register / login(下发 cookie) / get-login / chart-add / list-page / my-list-page / logout / 未登录 40100；同步生成(限流+AI 链路)与异步生成(wait→running→failed) 均已在 Phase 4/5 验证。
