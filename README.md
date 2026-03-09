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

### 端口说明

| 端口 | 服务 | 环境 | 配置文件 |
|------|------|------|----------|
| **12345** | 后端 API | dev（本地开发） | `backend/src/main/resources/application.yml` |
| **8101** | 后端 API | test / prod | `backend/src/main/resources/application-test.yml` / `application-prod.yml` |
| **8000** | 前端 | dev | `frontend/package.json` |
| 3306 | MySQL | 所有环境 | 各 `application-*.yml` |

### 1. 检查 MySQL 是否启动

```bash
# 方式一：直接连接
mysql -u root -p

# 方式二：查看 3306 端口是否被占用
lsof -i :3306

# 方式三：macOS 官方安装包用户
# 打开「系统偏好设置 → MySQL」面板查看状态

# 启动 / 停止 MySQL（官方安装包）
sudo /usr/local/mysql/support-files/mysql.server start
sudo /usr/local/mysql/support-files/mysql.server stop

# 启动 / 停止 MySQL（Homebrew 安装）
brew services start mysql
brew services stop mysql
```

### 2. 初始化数据库

```sql
-- 连接 MySQL 后执行
source backend/sql/create_table.sql;
```

### 3. 启动后端

```bash
cd backend
mvn spring-boot:run
# 后端运行在 http://localhost:12345/api
```

后端启动成功后可访问：

| 地址 | 用途 |
|------|------|
| http://localhost:12345/api/doc.html | Knife4j 接口文档（Swagger UI） |
| http://localhost:12345/api/v2/api-docs | Swagger JSON（前端 openapi 插件使用） |

### 4. 启动前端

```bash
cd frontend
npm install        # 首次或依赖变更时执行
npm run dev        # 启动开发服务器，运行在 http://localhost:8000
```

> **注意：** `proxy.ts` 中 dev 环境的代理默认被注释掉了，需要取消注释才能将 `/api/*` 请求转发到后端 `http://localhost:12345`。

### 5. 生成前端 API 代码

```bash
cd frontend
npm run openapi    # 需要后端正在运行，从 Swagger JSON 自动生成 TypeScript 请求代码
```

### 常用命令速查

```bash
# --- 后端 ---
cd backend && mvn spring-boot:run          # 启动后端

# --- 前端 ---
cd frontend && npm install                 # 安装依赖
cd frontend && npm run dev                 # 启动前端开发服务器
cd frontend && npm run openapi             # 从后端 Swagger 生成 API 代码
cd frontend && npm run build               # 构建生产包

# --- 数据库 ---
mysql -u root -p                           # 连接 MySQL
lsof -i :3306                              # 检查 MySQL 是否在运行
lsof -i :12345                             # 检查后端是否在运行
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
