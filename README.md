# GenBI

基于 React 与 Go 的智能数据分析应用：上传表格数据，由 AI 生成可视化图表与分析结论。

## 演示视频

仓库根目录包含录屏文件 **[`Demo.mov`](./Demo.mov)**（约 62MB）。在 GitHub 上可点击链接下载，或在本地克隆后直接播放。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | React 18、Ant Design Pro、UmiJS Max、TypeScript、ECharts |
| 后端 | Go 1.24、Gin、GORM、MySQL、Redis（会话/限流）、RabbitMQ（异步）、DeepSeek |
| 校验 | 前端对 AI 返回的图表 JSON 做 Zod 结构校验与安全过滤 |

## 仓库结构

```
GenBI/
├── frontend/     # Web 前端
├── backend/      # REST API
├── backend/sql/  # 建表等 SQL
└── Demo.mov      # 功能演示录屏
```

## 本地运行

### 环境

- Docker（含 Docker Compose v2）
- Node.js 20（与 CI 保持一致）
- pnpm 9.15.9：`npm install -g pnpm@9.15.9`
- Go 1.25+（实际版本以 `backend/go.mod` 为准）

### 依赖服务（Docker Compose）

MySQL、Redis、RabbitMQ 三个中间件由仓库根目录的 `docker-compose.yml` 统一管理，无需在本机逐个安装：

```bash
docker compose up -d          # 启动全部（后台）
docker compose ps             # 查看状态，STATUS 显示 healthy 才表示可用
docker compose logs -f mysql  # 查看日志
docker compose down           # 停止并删除容器（数据保留）
docker compose down -v        # 停止并清空数据（下次启动会重新执行建表脚本）
```

MySQL 首次启动会自动执行 `backend/sql/` 下的脚本，默认账号 `root`、密码 `123`、库名 `genbi`，与 `backend/.env.example` 的默认 `DB_DSN` 一致。

> 端口冲突：若本机已有 brew 安装的 Redis / RabbitMQ 在运行，请先执行 `brew services stop redis rabbitmq`，否则 6379 / 5672 被占用会导致容器启动失败。

### 端口（开发）

| 端口 | 说明 |
|------|------|
| 8080 | 后端 API（前端 dev 代理写死到该端口） |
| 8000 | 前端开发服务器 |
| 3306 | MySQL（Docker Compose） |
| 6379 | Redis（Docker Compose） |
| 5672 | RabbitMQ（Docker Compose） |
| 15672 | RabbitMQ 管理台（默认 guest/guest） |

### 数据库

无需手动执行：MySQL 容器首次启动时会自动运行 `backend/sql/*.sql`（建库、建表、outbox 表）。

如需重新初始化（例如修改了 SQL 脚本）：

```bash
docker compose down -v && docker compose up -d
```

### 后端

```bash
cd backend
cp .env.example .env    # 按需修改 DB_DSN / DEEPSEEK_API_KEY 等
make run                # 启动 HTTP 服务（:8080）
make consumer           # 另开终端，启动 RabbitMQ 异步消费者
```

详见 [`backend/README.md`](./backend/README.md)。

### 前端

```bash
cd frontend
pnpm install --frozen-lockfile
pnpm run dev
```

默认访问：http://localhost:8000 。dev 模式下 `/api` 已代理到 `http://localhost:8080`。

### 生成前端 API 代码

后端保持运行时：

```bash
cd frontend
pnpm run openapi
```

### 账号

通过接口注册后登录，具体以当前后端路由为准。

## Git 提交说明（约定式）

| 前缀 | 含义 |
|------|------|
| feat | 新功能 |
| fix | 修复 |
| chore | 依赖、构建、后端同步等 |
| docs | 文档 |
| refactor | 重构 |
