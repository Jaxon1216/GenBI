# 导学：GenBI

> 已确认输入：项目简称为 **GenBI**；目标岗位为 **全栈开发**；项目材料以当前仓库代码、设计文档和提交记录为准。本文默认你参与了仓库中呈现的实现，但面试时必须按真实职责把“我负责”调整为“我参与”或“我主导”，不要扩大个人边界。

## 1. 前置知识（面试高频标注）

| 知识点 | 为何需要 | 在本项目中的位置 | 高频度 |
|---|---|---|---|
| React 状态与副作用 | 理解表单提交、异步请求、轮询清理和页面状态切换 | `frontend/src/pages/AddChart/index.tsx`、`frontend/src/hooks/usePolling.ts` | 高 |
| TypeScript 与接口契约 | 保证前端请求参数、响应结构和页面消费方式一致 | `frontend/src/services/yubi/chartController.ts`、`frontend/src/services/yubi/typings.d.ts` | 高 |
| 不可信 JSON 的解析与校验 | 大模型输出不是稳定 API，必须校验、过滤并降级 | `frontend/src/utils/chartSchema.ts`、`frontend/src/components/SafeChart/index.tsx` | 高 |
| Gin 中间件与分层 | 理解请求如何经过恢复、日志、会话鉴权、handler 和 service | `backend/internal/bootstrap/router.go`、`backend/internal/middleware/`、`backend/internal/handler/` | 高 |
| Cookie Session 与 RBAC | 能解释登录态、服务端注销、管理员权限和数据权限 | `backend/internal/bootstrap/session.go`、`backend/internal/middleware/auth.go` | 高 |
| GORM、分页与索引 | 理解查询拼装、软删除、分页上限和索引选择 | `backend/internal/service/chart_service.go`、`backend/internal/model/`、`backend/sql/create_table.sql` | 高 |
| RabbitMQ 消息语义 | 理解持久化消息、预取、手动确认、重试和失败状态 | `backend/internal/mq/`、`backend/cmd/consumer/main.go` | 高 |
| 异步任务状态机 | 将长耗时 AI 调用从 HTTP 请求中解耦并反馈进度 | `backend/internal/service/chart_service.go`、`frontend/src/pages/myChart/index.tsx` | 高 |
| Redis 令牌桶 | 防止用户连续提交高成本模型请求 | `backend/internal/ratelimit/limiter.go`、`backend/internal/handler/chart_handler.go` | 中 |
| 文件与 CSV 处理 | 理解 multipart 上传、Excel 读取、CSV 转义和内存边界 | `backend/internal/util/excel.go`、`backend/internal/handler/chart_handler.go` | 中 |
| 可观测性与故障恢复 | 回答如何判断异步积压、模型失败、接口变慢和前端解析异常 | `backend/internal/middleware/logger.go`、`backend/internal/mq/consumer.go` | 高 |
| 测试与 CI | 区分已有单测、缺失的集成测试以及当前工具链问题 | `backend/internal/**/*_test.go`、`.github/workflows/`、`frontend/package.json` | 高 |

## 2. 重点亮点与学习顺序（先看这个）

| 顺序 | 亮点标题 | 为什么重要 | 通用技术关键词 | 先看哪些文件 |
|---|---|---|---|---|
| 1 | 端到端接口契约与跨语言迁移 | 项目不是单纯换语言，而是在不改前端的约束下用 Go 替换原后端，需要准确保留字段、状态和错误语义 | Contract First、Backward Compatibility、DTO、Migration | `backend/docs/spec/01-api-contract.md` → `backend/internal/dto/chart.go` → `frontend/src/services/yubi/typings.d.ts` |
| 2 | 长耗时任务的异步状态机 | AI 调用可能持续数十秒，异步化决定了接口吞吐、用户等待体验和失败恢复能力 | Message Queue、State Machine、Manual Ack、Retry、Eventual Consistency | `backend/internal/handler/chart_handler.go` → `backend/internal/service/chart_service.go` → `backend/internal/mq/consumer.go` |
| 3 | 不可信 AI 输出的多层防御 | 大模型即使收到严格提示，也可能返回非法 JSON 或危险配置，前后端都必须把它当作不可信输入 | Schema Validation、Sanitization、Error Boundary、Graceful Degradation | `backend/internal/service/deepseek_service.go` → `backend/internal/util/echarts.go` → `frontend/src/utils/chartSchema.ts` → `frontend/src/components/SafeChart/index.tsx` |
| 4 | 后端状态与前端轮询闭环 | 异步后端只有与前端的等待、执行、成功、失败视图以及停止条件配合，才是完整功能 | Polling、Exponential Backoff、Page Visibility、Race Condition、Cleanup | `frontend/src/pages/myChart/index.tsx` → `frontend/src/hooks/usePolling.ts` → `backend/internal/model/chart.go` |
| 5 | 会话鉴权、资源授权与成本保护 | 登录态、角色校验、资源所有权和高成本接口限流共同组成服务边界 | Session、RBAC、Object-Level Authorization、Rate Limiting | `backend/internal/bootstrap/session.go` → `backend/internal/middleware/auth.go` → `backend/internal/service/chart_service.go` → `backend/internal/ratelimit/limiter.go` |
| 6 | 可视化工作台与大数据展示 | 图表组合、响应式拖拽、布局持久化和虚拟列表体现了业务页面而非单接口 Demo | Responsive Layout、Local Persistence、Virtualization、Memoization | `frontend/src/pages/Dashboard/index.tsx` → `frontend/src/components/VirtualDataTable/index.tsx` |

## 3. 必备知识点

- [ ] 能从浏览器上传文件开始，完整画出“前端表单 → Gin handler → 文件转 CSV → 模型调用 → MySQL → ECharts 渲染”的同步链路。
- [ ] 能解释同步模式为什么会占用 HTTP 连接，以及异步模式如何通过消息队列和任务状态机解除等待。
- [ ] 能说明持久化消息、手动确认和重试分别解决什么问题，也能指出它们不等于端到端 exactly-once。
- [ ] 能指出“数据库已写入等待任务，但消息发布失败”这一双写一致性窗口，并提出本地消息表或事务 Outbox。
- [ ] 能解释为什么模型输出必须经历 JSON 解析、Schema 校验、危险字段过滤和渲染错误边界。
- [ ] 能解释轮询中的指数退避、页面不可见暂停、组件卸载清理和旧请求覆盖新状态的竞态。
- [ ] 能区分登录认证、角色权限与资源级授权，并检查图表查询是否限制到当前用户。
- [ ] 能说明 Session 与 JWT 的取舍，以及 Cookie 的 `HttpOnly`、`SameSite`、`Secure`、CSRF 风险。
- [ ] 能说明令牌桶的限流维度、突发能力、Redis 故障策略和高成本 AI 接口的配额设计。
- [ ] 能分析当前表索引是否覆盖“用户 + 状态 + 创建时间”的常用查询，并会用 `EXPLAIN` 验证。
- [ ] 能说明当前文件处理和 CSV 展示的内存、文件大小、转义解析边界。
- [ ] 能诚实说明现有测试覆盖与缺口，不把“命令能运行”表述成“生产质量已经完备”。

## 4. 推荐阅读（结合仓库）

| 主题 | 通用技术点 | 建议阅读位置 | 预计时间 | 读完能回答什么 |
|---|---|---|---|---|
| 项目目标与边界 | Scope、技术选型、迁移约束 | `README.md`、`backend/docs/spec/00-overview.md` | 15 分钟 | 项目解决什么问题，为什么是全栈项目 |
| API 契约兼容 | DTO、状态枚举、统一响应 | `backend/docs/spec/01-api-contract.md`、`backend/internal/dto/chart.go`、`frontend/src/services/yubi/chartController.ts` | 25 分钟 | Go 重写时如何做到前端零改动 |
| HTTP 请求主链路 | Middleware、Handler、Service | `backend/internal/bootstrap/router.go`、`backend/internal/handler/chart_handler.go`、`backend/internal/service/chart_service.go` | 40 分钟 | 一次图表生成请求经过哪些层 |
| 异步任务处理 | Queue、Ack、Retry、状态机 | `backend/internal/mq/client.go`、`backend/internal/mq/producer.go`、`backend/internal/mq/consumer.go`、`backend/cmd/consumer/main.go` | 45 分钟 | 异步任务怎样投递、消费和失败兜底 |
| 模型调用边界 | Timeout、Prompt、JSON Decode | `backend/internal/service/deepseek_service.go` | 25 分钟 | 模型超时或返回异常时系统如何处理 |
| AI 输出安全渲染 | Zod、Sanitize、Error Boundary | `backend/internal/util/echarts.go`、`frontend/src/utils/chartSchema.ts`、`frontend/src/components/SafeChart/index.tsx`、`frontend/src/components/ChartErrorBoundary/index.tsx` | 40 分钟 | 为什么不能直接把模型 JSON 交给图表库 |
| 异步前端反馈 | Polling、Backoff、Visibility API | `frontend/src/hooks/usePolling.ts`、`frontend/src/pages/myChart/index.tsx` | 35 分钟 | 如何避免固定频率轮询造成空耗 |
| 会话与权限 | Cookie Session、RBAC、资源授权 | `backend/internal/bootstrap/session.go`、`backend/internal/middleware/auth.go`、`backend/internal/service/user_service.go`、`backend/internal/service/chart_service.go` | 40 分钟 | 登录、管理员校验和本人操作分别在哪一层 |
| 数据建模与查询 | GORM、Soft Delete、Index、Pagination | `backend/internal/model/chart.go`、`backend/internal/model/user.go`、`backend/sql/create_table.sql`、`backend/internal/service/chart_service.go` | 35 分钟 | 表结构、查询条件和索引如何配合 |
| 文件数据管道 | Multipart、Excel、CSV Escaping | `backend/internal/handler/chart_handler.go`、`backend/internal/util/excel.go` | 30 分钟 | Excel 如何变成模型可消费的数据 |
| 可视化看板 | Responsive Grid、Local Storage | `frontend/src/pages/Dashboard/index.tsx` | 30 分钟 | 拖拽布局如何保存，当前方案有什么边界 |
| 大数据表格展示 | Windowing、Memoization、CSV Parsing | `frontend/src/components/VirtualDataTable/index.tsx` | 25 分钟 | 虚拟列表优化了什么，未优化什么 |
| 测试与工程化 | Unit Test、Static Check、CI | `backend/internal/service/chart_service_test.go`、`backend/internal/util/echarts_test.go`、`backend/internal/util/excel_test.go`、`.github/workflows/` | 30 分钟 | 现有质量门禁覆盖了哪些风险 |
| 跨语言重写演进 | Migration、ADR、兼容策略 | `backend/docs/spec/04-migration-map.md`、`backend/CHANGELOG.md` | 25 分钟 | Java 到 Go 的映射和取舍是什么 |

## 5. 自学提醒

若某文件或原理看不懂，请继续追问 AI；本 skill 负责给学习路径与题目，不提供逐行讲解。建议每读完一个主题，都脱离源码口述一次“问题、机制、落点、风险、验证”，再回到源码补证据。

## 6. 项目技术定位

- **定位：交叉（全栈，后端与 AI 工程链路偏重）。**
- **依据：**前端覆盖 React 表单、状态反馈、安全图表渲染、轮询、虚拟列表和拖拽看板；后端覆盖 Gin API、MySQL、Redis Session/限流、RabbitMQ 异步任务及模型服务集成，形成了从交互到存储和异步计算的完整闭环。

## 7. 核心原理解析

### 7.1 接口契约驱动的跨语言替换

- **问题：**后端由 Java 改写为 Go 时，前端已经依赖固定字段、状态值、Cookie 会话和统一响应格式，任一细节漂移都会导致页面静默失败。
- **机制：**先把路由、入参、出参、错误码和状态机固化为契约，再让 DTO、数据库映射和前端类型围绕同一语义实现。
- **项目落点：**`backend/docs/spec/01-api-contract.md` 明确契约，`backend/internal/dto/` 与 `frontend/src/services/yubi/typings.d.ts` 提供两端映射，实际纠正了生成字段和异步状态值不一致的问题。

### 7.2 同步调用向异步任务状态机演进

- **问题：**模型请求最长可等待 60 秒，若所有生成都绑定 HTTP 生命周期，会占用连接并放大超时、重复提交和突发流量风险。
- **机制：**请求阶段只持久化任务并投递任务标识，消费者独立执行模型调用，以等待、执行中、成功、失败四态表达生命周期，前端只查询状态。
- **项目落点：**`backend/internal/service/chart_service.go` 负责状态迁移，`backend/internal/mq/consumer.go` 负责消费、重试和确认，`frontend/src/pages/myChart/index.tsx` 根据状态展示结果。

### 7.3 消息可靠性与一致性边界

- **问题：**消息持久化和手动确认只能提高 RabbitMQ 内部可靠性，不能自动解决数据库与消息队列之间的双写一致性，也不能保证业务只执行一次。
- **机制：**当前实现用持久化队列、单条预取、有限重试和最终失败状态形成最小闭环；更高等级方案需要消费幂等、发布确认、事务 Outbox、死信队列和补偿扫描。
- **项目落点：**`backend/internal/mq/producer.go` 发布持久化消息，`backend/internal/mq/consumer.go` 手动确认；`backend/internal/handler/chart_handler.go` 中“先入库再发消息”暴露了需要演进的双写窗口。

### 7.4 不可信模型输出的纵深防御

- **问题：**提示词只能约束概率输出，不能保证模型始终返回合法、完整、安全的 ECharts 配置。
- **机制：**服务端先做格式清洗和 JSON 合法性检查，前端再做结构校验、危险字段过滤和渲染异常隔离；失败时显示可理解的降级结果，而不是白屏。
- **项目落点：**`backend/internal/util/echarts.go`、`frontend/src/utils/chartSchema.ts`、`frontend/src/components/SafeChart/index.tsx` 和 `frontend/src/components/ChartErrorBoundary/index.tsx` 构成多层边界。

### 7.5 自适应轮询与资源控制

- **问题：**异步任务需要及时反馈，但固定高频轮询会在任务无变化、页面切后台或用户开多个标签页时持续制造无效请求。
- **机制：**仅在存在未完成任务时轮询；数据无变化时指数退避，变化时恢复基础间隔；页面不可见时暂停，恢复时立即刷新；组件卸载时清理定时器。
- **项目落点：**`frontend/src/hooks/usePolling.ts` 封装调度，`frontend/src/pages/myChart/index.tsx` 通过任务状态快照反馈是否有变化。

### 7.6 认证、授权与资源隔离

- **问题：**“已登录”不代表可以操作任意资源，前端隐藏按钮也不能代替服务端授权。
- **机制：**Redis Session 解决身份认证，角色中间件解决管理员权限，service 层按资源所有者检查编辑与删除；查询接口还需要逐项审查对象级授权。
- **项目落点：**`backend/internal/middleware/auth.go`、`backend/internal/bootstrap/router.go` 与 `backend/internal/service/chart_service.go`。当前按 ID 获取图表和通用图表列表只要求登录，若数据应私有，需要补充本人或管理员限制。

## 8. 关键设计决策

| 决策点 | 备选 | 取舍 | 风险 | 验证 |
|---|---|---|---|---|
| Go 重写后端且保持前端不变 | 同时重写前后端；新增 BFF 适配 | 保留前端可缩小迁移范围，但后端必须严格兼容旧契约 | 旧契约中的不合理语义会被继续继承 | 契约测试、关键页面 E2E、旧新接口响应快照对比 |
| Redis Session 而非 JWT | JWT Access/Refresh Token | 与现有 Cookie 登录和服务端注销兼容，开发成本较低 | Redis 成为登录关键依赖，跨域 Cookie 和 CSRF 需治理 | 登录/注销集成测试、会话过期测试、Redis 故障演练 |
| RabbitMQ 承载异步生成 | 进程内 goroutine；Redis 队列；任务平台 | 独立消费者可削峰并复用原架构语义 | 双写不一致、重复消费、连接恢复和死信治理尚不完整 | 故障注入、积压监控、重复投递测试、Outbox 对照实验 |
| 轮询查询任务状态 | WebSocket；SSE；长轮询 | 实现简单且适合低频状态更新，通过退避降低请求量 | 多标签页和大规模用户下仍有空耗 | 统计每任务平均轮询次数、接口 QPS、完成感知延迟 |
| 前后端双层校验模型输出 | 只靠 Prompt；只在后端解析 | 纵深防御能同时保护数据层和渲染层 | 清洗可能掩盖模型质量问题，宽松 Schema 仍可能放入异常配置 | 非法样本集、解析成功率、降级率、渲染错误率 |
| 看板布局保存在 localStorage | 服务端持久化；URL 快照 | 无需新增表和接口，适合单机演示 | 无法跨设备同步，图表数据会陈旧，存储可被用户清除 | 刷新恢复测试、账号切换测试、服务端方案原型 |
| GORM service 直接访问数据库 | Repository 层；SQL 生成器 | 当前业务规模下层级更少、路径清晰 | 业务增长后查询复用和事务边界可能分散 | 统计重复查询与事务需求，到达阈值再抽象 |
| 固定盐 MD5 保持旧数据兼容 | bcrypt；Argon2id | 迁移期可以读取旧账号 | 不满足生产密码安全要求 | 双哈希渐进迁移测试，登录时升级旧哈希 |

## 9. 量化与验证（含待测，建议）

### 当前可确认结果

- `cd backend && go test ./...`：通过；现有测试集中在 Prompt 拼装、排序字段过滤、分页默认值、ECharts JSON 清洗、CSV 转义和密码哈希。
- `cd backend && go vet ./...`：通过。
- `cd frontend && npm run lint:js`：通过，仅提示 Browserslist 数据陈旧。
- `cd frontend && npm run tsc`：当前未通过。错误来自依赖包声明使用 TypeScript 5 的 `const` 类型参数语法，而仓库声明 TypeScript 4.9.5；应先统一包管理器、锁文件和 TypeScript/Ant Design Pro 版本，再把类型检查设为可靠 CI 门禁。

### 建议补测

| 目标 | 环境与工具 | 指标 | 基线与对比注意 |
|---|---|---|---|
| 同步与异步生成体验 | 固定 CSV、固定 Prompt，在本地或测试环境分别压测；使用 k6/JMeter + 浏览器 Performance | API P50/P95、超时率、任务完成耗时、HTTP 并发占用 | 模型响应波动大，应固定样本并分离“排队时间”和“模型时间”（待测） |
| 消息可靠性 | 测试 RabbitMQ/MySQL，注入发布失败、消费者崩溃、重复消息 | 丢单数、重复执行数、失败任务可恢复率、队列积压 | 当前无幂等与 Outbox，先记录现状再验证改造（待测） |
| 轮询收益 | 浏览器埋点或 DevTools，对比固定 3 秒轮询与指数退避 | 单任务请求数、完成感知延迟、后台标签页请求数 | 保证任务时长分布一致，分别测试前台、后台和多标签页（待测） |
| 图表防御能力 | 构造合法、缺字段、代码围栏、函数串、超大数组等样本 | 解析成功率、降级率、渲染异常率、解析耗时 | 不能只测 `JSON.parse`，要覆盖 Schema、过滤和 Error Boundary（待测） |
| 文件处理容量 | 生成 1 MB、10 MB、50 MB 的 CSV/XLSX；记录 Go pprof | 峰值内存、转换耗时、请求失败率 | 当前实现整体读入内存，需设置合理文件上限后再比较流式方案（待测） |
| 查询与索引 | MySQL 构造 10 万至 100 万图表记录，使用 `EXPLAIN ANALYZE` | 查询 P95、扫描行数、索引命中、分页深度影响 | 对比现有单列用户索引与“用户 + 状态 + 创建时间”复合索引（待测） |
| 前端长列表 | Chrome Performance/React Profiler，对比普通列表和窗口化列表 | Commit 时间、主线程阻塞、DOM 节点数、滚动帧率 | 当前只虚拟化行，CSV 解析和横向列仍需单独评估（待测） |

### 面试量化口径

没有生产流量与监控数据时，应说“我已完成可运行闭环和静态检查，性能与稳定性指标正在按上述方案补测”，不要编造 QPS、延迟下降比例或线上事故数据。可优先补出三组最有价值的数据：异步接口返回耗时、每个任务的平均轮询请求数、异常图表配置的前端降级成功率。
