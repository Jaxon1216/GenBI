# 快速开始 - 5 分钟上手

## 📋 第一步：复制配置文件

### 1. Docker 配置

**创建 `frontend/Dockerfile`**：

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run lint
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**创建 `frontend/.dockerignore`**：

```
node_modules
dist
.umi
.umi-production
.DS_Store
*.log
```

---

### 2. GitHub Actions 配置

**创建 `.github/workflows/frontend-ci.yml`**：

```yaml
name: Frontend CI/CD

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
  pull_request:
    branches: [ main, develop ]

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: 安装依赖
        working-directory: ./frontend
        run: npm ci
      
      - name: 代码检查
        working-directory: ./frontend
        run: |
          npm run lint:js
          npm run lint:prettier
          npm run tsc
      
      - name: 构建
        working-directory: ./frontend
        run: npm run build
```

---

### 3. VSCode 配置

**创建 `frontend/.vscode/settings.json`**：

```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  }
}
```

---

## 🚀 第二步：运行命令

### 本地开发

```bash
cd frontend

# 1. 安装依赖（如果还没安装）
npm install

# 2. 安装 Git 钩子
npm run prepare

# 3. 运行代码检查
npm run lint

# 预期：显示检查结果，可能有一些警告
```

---

### 测试 Git 提交钩子

```bash
# 1. 修改一个文件
echo "// test" >> src/app.tsx

# 2. 尝试提交
git add src/app.tsx
git commit -m "test: 测试提交钩子"

# 预期：
# ✔ Preparing lint-staged...
# ✔ Running tasks for staged files...
# ✔ Applying modifications...
# ✔ Cleaning up...
# [main xxx] test: 测试提交钩子

# 如果代码有错误，会被拦截：
# ✖ npm run lint-staged:js:
#   error  'xxx' is defined but never used
```

---

### 测试 Docker

```bash
cd frontend

# 1. 构建镜像（会自动运行代码检查）
docker build -t genbi-frontend:test .

# 预期：
# [+] Building 45.2s (15/15) FINISHED
# => [builder 7/7] RUN npm run build
# => [stage-1 2/2] COPY --from=builder /app/dist /usr/share/nginx/html
# => exporting to image

# 2. 运行容器
docker run -d -p 8080:80 --name genbi-test genbi-frontend:test

# 3. 测试访问
curl http://localhost:8080
# 或
open http://localhost:8080

# 预期：看到前端页面

# 4. 清理
docker stop genbi-test && docker rm genbi-test
```

---

### 测试 GitHub Actions

```bash
# 1. 提交配置文件
git add .github/workflows/frontend-ci.yml
git add frontend/Dockerfile
git add frontend/.dockerignore
git commit -m "ci: 添加 CI/CD 配置"

# 2. 推送到 GitHub
git push origin main

# 3. 查看 Actions 运行
# 访问：https://github.com/你的用户名/GenBI/actions

# 预期：
# ✅ 代码检查和测试 - 通过
# ✅ 构建项目 - 通过
```

---

## ✅ 第三步：验证效果

### 验证 ESLint

```bash
# 创建一个有问题的文件
cat > src/test-error.ts << 'EOF'
const unusedVar = 'test';
console.log('debug');
function badName() {}
EOF

# 运行检查
npm run lint:js

# 预期输出：
# src/test-error.ts
#   1:7   error  'unusedVar' is assigned but never used  @typescript-eslint/no-unused-vars
#   2:1   error  Unexpected console statement            no-console
#   3:10  error  Function name should be camelCase       camelcase

# 删除测试文件
rm src/test-error.ts
```

---

### 验证 Prettier

```bash
# 创建格式混乱的文件
cat > src/test-format.ts << 'EOF'
const obj={name:"test",age:20,city:"Beijing"}
const arr=[1,2,3,4,5]
EOF

# 运行格式化
npx prettier --write src/test-format.ts

# 查看结果
cat src/test-format.ts

# 预期输出：
# const obj = { name: 'test', age: 20, city: 'Beijing' };
# const arr = [1, 2, 3, 4, 5];

# 删除测试文件
rm src/test-format.ts
```

---

### 验证 TypeScript

```bash
# 创建类型错误的文件
cat > src/test-type.ts << 'EOF'
function greet(name) {
  return 'Hello ' + name;
}
const result: number = greet('World');
EOF

# 运行类型检查
npm run tsc

# 预期输出：
# src/test-type.ts:1:16 - error TS7006: Parameter 'name' implicitly has an 'any' type.
# src/test-type.ts:4:7 - error TS2322: Type 'string' is not assignable to type 'number'.

# 删除测试文件
rm src/test-type.ts
```

---

## 📊 命令速查

```bash
# 代码检查
npm run lint:js              # ESLint
npm run lint:fix             # 自动修复
npm run prettier             # Prettier
npm run tsc                  # TypeScript
npm run lint                 # 全部检查

# Git
npm run prepare             # 安装钩子
git commit                  # 触发检查

# Docker
docker build -t genbi-frontend .
docker run -d -p 8080:80 genbi-frontend
docker ps
docker stop <id>

# 开发
npm run dev                 # 开发服务器
npm run build               # 构建
```

---

## 🎯 预期效果总结

### ✅ 本地开发
- 保存文件时自动格式化（VSCode）
- 提交代码时自动检查（Husky）
- 类型错误实时提示（TypeScript）

### ✅ Docker
- 构建时自动运行代码检查
- 代码有问题时构建失败
- 生产环境使用 Nginx 部署

### ✅ GitHub Actions
- 推送代码时自动运行 CI
- 检查失败时阻止合并
- 构建产物自动上传

---

## 🐛 常见问题

### Husky 不工作？
```bash
cd frontend
rm -rf .husky
npm run prepare
chmod +x .husky/*
```

### Docker 构建失败？
```bash
# 查看详细日志
docker build --progress=plain -t genbi-frontend .
```

### GitHub Actions 失败？
- 检查 Node.js 版本（需要 18）
- 检查路径配置（frontend/**）
- 查看 Actions 日志

---

## 📝 完成检查清单

- [ ] 复制了 Dockerfile 和 .dockerignore
- [ ] 复制了 GitHub Actions 配置
- [ ] 复制了 VSCode 配置
- [ ] 运行 `npm run lint` 成功
- [ ] 测试 Git 提交钩子成功
- [ ] Docker 构建和运行成功
- [ ] GitHub Actions 运行成功

---

**总耗时：5-10 分钟**  
**详细说明：查看 CODE_QUALITY_GUIDE.md**
