# GenBI - 智能 BI 数据分析平台

基于 React + Spring Boot + AI 的智能数据分析平台。用户上传 Excel 数据，AI 自动生成可视化图表和分析结论。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | React 18 + Ant Design Pro + UmiJS Max + TypeScript |
| 后端 | Spring Boot 2.7 + MyBatis-Plus + MySQL |
| AI | 鱼聪明 AI SDK（后续期引入） |
| 异步 | RabbitMQ（后续期引入） |

## 项目结构

```
GenBI/
├── frontend/          # 前端（手写，有完整 commit 记录）
├── backend/           # 后端（每期从教程源码更新）
├── PROBLEM_SOLUTIONS.md   # 学习过程中的问题解决记录
└── README.md
```

## 学习工作流

本项目跟随教程分 7 期迭代开发。**前端是我手写的**，后端使用教程提供的源码。

### 每期开发流程

```
┌─────────────────────────────────────────────────────┐
│  1. 更新后端源码                                       │
│     将第 N 期后端源码覆盖到 backend/ 目录                 │
│     git add backend/                                  │
│     git commit -m "chore: 更新后端至第 N 期"             │
│                                                       │
│  2. 启动后端环境                                       │
│     cd backend && mvn spring-boot:run                 │
│     （确保 MySQL 已启动，yubi 库已建好）                  │
│                                                       │
│  3. 生成前端 API 接口                                   │
│     cd frontend && npm run openapi                    │
│     （后端必须在运行中，才能拉取 Swagger 文档）            │
│                                                       │
│  4. 手写前端代码                                       │
│     跟着教程实现前端功能，正常 commit                     │
│     git add frontend/                                 │
│     git commit -m "feat: 实现 xxx 功能"                │
│                                                       │
│  5. 测试验证                                           │
│     npm run dev 启动前端，连接真实后端测试                │
│                                                       │
│  6. 推送                                               │
│     git push origin main                              │
└─────────────────────────────────────────────────────┘
```

### 各期内容概览

| 期数 | 后端新增 | 前端任务 |
|:---:|---|---|
| 1 | 项目初始化、用户登录注册 | 对接真实登录接口，替换 Mock |
| 2 | 图表 CRUD 接口 | 图表管理页面 |
| 3 | AI 分析接口（同步） | 智能分析页面（上传数据 → 生成图表） |
| 4 | AI 分析接口（异步） | 异步分析页面 + 状态轮询 |
| 5 | RabbitMQ 消息队列 | 优化异步体验 |
| 6 | 限流 + 安全优化 | 前端限流提示 + 错误处理 |
| 7 | 项目上线部署 | 构建打包 + 部署 |

## 本地开发

### 前置条件

- Node.js 16+
- Java 8
- Maven
- MySQL 8.0（创建 `yubi` 数据库）

### 1. 初始化数据库

```sql
-- 连接 MySQL 后执行
source backend/sql/create_table.sql;
```

### 2. 启动后端

```bash
cd backend
mvn spring-boot:run
# 后端运行在 http://localhost:12345/api
# Knife4j 接口文档：http://localhost:12345/api/doc.html
```

### 3. 启动前端

```bash
cd frontend
pnpm install
npm run dev
# 前端运行在 http://localhost:8000
# 请求通过 proxy 转发到后端
```

### 默认账号

在 Knife4j 中调用 `POST /api/user/register` 注册账号，然后登录。

## Git 提交规范

| 前缀 | 用途 | 示例 |
|---|---|---|
| `feat` | 新功能 | `feat: 实现智能分析页面` |
| `fix` | 修复 | `fix: 修复登录态丢失问题` |
| `chore` | 后端更新/配置 | `chore: 更新后端至第 3 期` |
| `docs` | 文档 | `docs: 更新问题记录` |
| `refactor` | 重构 | `refactor: 抽取图表组件` |
| `style` | 样式调整 | `style: 优化分析页面布局` |
