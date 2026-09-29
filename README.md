# GenBI

基于 React 与 Spring Boot 的智能数据分析应用：上传表格数据，由 AI 生成可视化图表与分析结论。

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

- Node.js 16+
- Go 1.24+
- MySQL 8，创建数据库并执行 `backend/sql/create_table.sql`
- Redis：`brew install redis && brew services start redis`
- RabbitMQ：`brew install rabbitmq && brew services start rabbitmq`

### 端口（开发）

| 端口 | 说明 |
|------|------|
| 8080 | 后端 API（前端 dev 代理写死到该端口） |
| 8000 | 前端开发服务器 |
| 3306 | MySQL |
| 6379 | Redis |
| 5672 | RabbitMQ |

### 数据库

```sql
-- 在 MySQL 中执行
source backend/sql/create_table.sql;
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
npm install
npm run dev
```

默认访问：http://localhost:8000 。dev 模式下 `/api` 已代理到 `http://localhost:8080`。

### 生成前端 API 代码

后端保持运行时：

```bash
cd frontend
npm run openapi
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
