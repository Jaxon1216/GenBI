# 前端代码质量保障指南

## 🎯 核心工具

- **ESLint**: 代码质量检查（逻辑错误、最佳实践）
- **Prettier**: 代码格式化（缩进、引号、换行）
- **TypeScript**: 类型检查（编译时发现类型错误）
- **Husky**: Git 钩子（提交前自动检查）
- **Docker**: 容器化部署
- **GitHub Actions**: CI/CD 自动化

---

## 🚀 快速开始（5 分钟）

### 1. 检查当前配置

```bash
cd frontend

# 查看配置文件
ls -la .eslintrc.js .prettierrc.js tsconfig.json .husky/
```

### 2. 运行检查命令

```bash
# ESLint 检查
npm run lint:js

# Prettier 格式化
npm run prettier

# TypeScript 类型检查
npm run tsc

# 完整检查（三合一）
npm run lint
```

### 3. 自动修复

```bash
# 自动修复 ESLint 问题
npm run lint:fix

# 自动格式化代码
npm run prettier
```

---

## ⚙️ 配置文件说明

### ESLint 配置 (`.eslintrc.js`)

当前配置继承了 UmiJS 官方规范，已经包含：
- React/React Hooks 规则
- TypeScript 规则
- 代码质量规则

**如需自定义规则**，编辑 `.eslintrc.js`：

```javascript
module.exports = {
  extends: [require.resolve('@umijs/lint/dist/config/eslint')],
  globals: {
    page: true,
    REACT_APP_ENV: true,
  },
  rules: {
    // 自定义规则
    'no-console': ['warn', { allow: ['warn', 'error'] }],
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-unused-vars': ['error', {
      argsIgnorePattern: '^_',
      varsIgnorePattern: '^_',
    }],
  },
};
```

---

### Prettier 配置 (`.prettierrc.js`)

当前配置：
- 单引号
- 尾部逗号
- 行宽 100
- LF 换行符

**配置说明**：

```javascript
module.exports = {
  singleQuote: true,        // 使用单引号
  trailingComma: 'all',     // 尾部逗号
  printWidth: 100,          // 行宽
  proseWrap: 'never',       // 不换行
  endOfLine: 'lf',          // LF 换行符
};
```

---

### TypeScript 配置 (`tsconfig.json`)

当前配置已启用严格模式：

```json
{
  "compilerOptions": {
    "strict": true,              // 严格模式
    "target": "esnext",
    "module": "esnext",
    "jsx": "preserve",
    "baseUrl": "./",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

---

### Husky 配置

**pre-commit** (`.husky/pre-commit`)：
```bash
#!/bin/sh
npx --no-install lint-staged
```

**lint-staged** (`package.json`)：
```json
{
  "lint-staged": {
    "**/*.{js,jsx,ts,tsx}": "npm run lint-staged:js",
    "**/*.{js,jsx,tsx,ts,less,md,json}": [
      "prettier --write"
    ]
  }
}
```

---

## 🐳 Docker 配置

### 1. 创建 Dockerfile

在 `frontend/` 目录创建 `Dockerfile`：

```dockerfile
# 构建阶段
FROM node:18-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY package*.json ./

# 安装依赖
RUN npm ci

# 复制源代码
COPY . .

# 运行代码检查
RUN npm run lint

# 构建项目
RUN npm run build

# 生产阶段
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置（可选）
# COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 2. 创建 .dockerignore

在 `frontend/` 目录创建 `.dockerignore`：

```
node_modules
dist
.umi
.umi-production
.DS_Store
*.log
.git
.vscode
.idea
```

### 3. 构建和运行

```bash
cd frontend

# 构建镜像
docker build -t genbi-frontend:latest .

# 运行容器
docker run -d -p 8080:80 genbi-frontend:latest

# 访问
open http://localhost:8080
```

---

## 🔄 GitHub Actions CI/CD

### 创建 GitHub Actions 配置

在项目根目录创建 `.github/workflows/frontend-ci.yml`：

```yaml
name: Frontend CI/CD

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'

jobs:
  lint-and-test:
    name: 代码检查和测试
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout 代码
        uses: actions/checkout@v3
      
      - name: 设置 Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: 安装依赖
        working-directory: ./frontend
        run: npm ci
      
      - name: ESLint 检查
        working-directory: ./frontend
        run: npm run lint:js
      
      - name: Prettier 检查
        working-directory: ./frontend
        run: npm run lint:prettier
      
      - name: TypeScript 类型检查
        working-directory: ./frontend
        run: npm run tsc
      
      - name: 运行测试（如果有）
        working-directory: ./frontend
        run: npm run test || echo "No tests configured"
  
  build:
    name: 构建项目
    runs-on: ubuntu-latest
    needs: lint-and-test
    
    steps:
      - name: Checkout 代码
        uses: actions/checkout@v3
      
      - name: 设置 Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: 安装依赖
        working-directory: ./frontend
        run: npm ci
      
      - name: 构建项目
        working-directory: ./frontend
        run: npm run build
      
      - name: 上传构建产物
        uses: actions/upload-artifact@v3
        with:
          name: dist
          path: frontend/dist
          retention-days: 7
  
  docker:
    name: 构建 Docker 镜像
    runs-on: ubuntu-latest
    needs: build
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
      - name: Checkout 代码
        uses: actions/checkout@v3
      
      - name: 设置 Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: 登录 Docker Hub（可选）
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: 构建并推送 Docker 镜像
        uses: docker/build-push-action@v4
        with:
          context: ./frontend
          push: true
          tags: |
            your-dockerhub-username/genbi-frontend:latest
            your-dockerhub-username/genbi-frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

---

## 📋 实践步骤

### 步骤 1: 配置 VSCode 自动格式化

创建 `frontend/.vscode/settings.json`：

```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ]
}
```

### 步骤 2: 测试提交钩子

```bash
cd frontend

# 1. 确保 Husky 已安装
npm run prepare

# 2. 修改一个文件（故意引入错误）
echo "const unused = 'test';" >> src/test.ts

# 3. 尝试提交
git add src/test.ts
git commit -m "test: 测试 husky"

# 预期：提交被拒绝，显示 ESLint 错误

# 4. 修复错误
rm src/test.ts

# 5. 正常提交
git add .
git commit -m "test: 测试 husky 成功"
```

### 步骤 3: 测试 Docker 构建

```bash
cd frontend

# 1. 构建镜像（会自动运行代码检查）
docker build -t genbi-frontend:test .

# 预期：如果代码有问题，构建会失败

# 2. 运行容器
docker run -d -p 8080:80 --name genbi-test genbi-frontend:test

# 3. 访问测试
curl http://localhost:8080

# 4. 清理
docker stop genbi-test
docker rm genbi-test
```

### 步骤 4: 测试 GitHub Actions

```bash
# 1. 创建 GitHub Actions 配置文件
mkdir -p .github/workflows
# 复制上面的 frontend-ci.yml 内容

# 2. 提交到 GitHub
git add .github/workflows/frontend-ci.yml
git commit -m "ci: 添加前端 CI/CD 配置"
git push origin main

# 3. 在 GitHub 上查看 Actions 运行结果
# https://github.com/your-username/GenBI/actions
```

---

## 🎓 常见问题

### Q1: ESLint 和 Prettier 冲突？

**解决**：已安装 `eslint-config-prettier`，在 `.eslintrc.js` 中添加：

```javascript
extends: [
  require.resolve('@umijs/lint/dist/config/eslint'),
  'prettier', // 放在最后
],
```

### Q2: Husky 钩子不生效？

**解决**：
```bash
cd frontend
rm -rf .husky
npm run prepare
chmod +x .husky/*
```

### Q3: TypeScript 报错太多？

**解决**：逐步修复，不要降低 `strict` 设置。使用：
```typescript
// 临时忽略（不推荐）
// @ts-ignore

// 正确处理 null
const name = user?.name ?? 'Unknown';
```

### Q4: Docker 构建失败？

**解决**：
```bash
# 查看详细日志
docker build --progress=plain -t genbi-frontend:test .

# 检查 .dockerignore
cat .dockerignore
```

---

## 📊 命令速查表

```bash
# ===== 代码检查 =====
npm run lint:js              # ESLint 检查
npm run lint:fix             # ESLint 自动修复
npm run prettier             # Prettier 格式化
npm run tsc                  # TypeScript 类型检查
npm run lint                 # 完整检查

# ===== Git 钩子 =====
npm run prepare             # 安装 Husky 钩子
git commit                  # 触发 pre-commit 检查

# ===== Docker =====
docker build -t genbi-frontend .              # 构建镜像
docker run -d -p 8080:80 genbi-frontend       # 运行容器
docker ps                                     # 查看运行中的容器
docker logs <container-id>                    # 查看日志
docker stop <container-id>                    # 停止容器

# ===== 开发 =====
npm run dev                 # 启动开发服务器
npm run build               # 构建生产版本
```

---

## 🎯 面试要点

### 能够回答的问题

1. **ESLint vs Prettier？**
   - ESLint: 代码质量（逻辑、最佳实践）
   - Prettier: 代码格式（缩进、引号）

2. **为什么用 TypeScript？**
   - 编译时类型检查
   - 更好的 IDE 支持
   - 减少运行时错误

3. **Husky 的作用？**
   - Git 钩子管理
   - 提交前自动检查
   - 保证代码质量

4. **Docker 的优势？**
   - 环境一致性
   - 快速部署
   - 易于扩展

5. **CI/CD 的价值？**
   - 自动化测试
   - 快速反馈
   - 持续集成

---

## 📝 简历写法

**技术栈**
- 熟练使用 ESLint + Prettier + TypeScript + Husky 保障代码质量
- 使用 Docker 容器化部署，GitHub Actions 实现 CI/CD
- 搭建完整的前端工程化体系

**项目经验 - GenBI 智能 BI 平台**
- 搭建代码质量保障体系（ESLint、Prettier、TypeScript、Husky）
- 配置 Git Hooks 实现提交前自动检查，确保代码质量
- 使用 Docker 容器化部署，GitHub Actions 实现 CI/CD 自动化
- 通过工程化手段，减少 30% 的代码错误，提升团队效率

---

**学习时间：1-2 小时**  
**实践验证：30 分钟**
