# 前端工程化与代码规范学习指南

## 🎯 学习目标
通过实践掌握 ESLint + Prettier + TypeScript + Husky 工具链

---

## 📖 第一部分：理解工具的作用

### 1. ESLint - 代码质量守护者
**作用**：检查代码中的潜在问题和不规范写法
- 未使用的变量
- 可能的逻辑错误
- 不符合团队规范的代码风格
- React Hooks 使用错误

**命令**：
```bash
npm run lint:js          # 检查代码
npm run lint:fix         # 自动修复
```

### 2. Prettier - 代码格式化工具
**作用**：统一代码格式，避免格式争论
- 缩进（空格/Tab）
- 引号（单引号/双引号）
- 分号
- 换行
- 行宽

**命令**：
```bash
npm run prettier         # 格式化所有文件
npm run lint:prettier    # 检查并格式化
```

### 3. TypeScript - 类型检查
**作用**：在编译时发现类型错误
- 函数参数类型
- 返回值类型
- 对象属性类型
- 避免运行时类型错误

**命令**：
```bash
npm run tsc             # 类型检查
```

### 4. Husky + lint-staged - 自动化守门员
**作用**：在 git commit 前自动运行检查
- pre-commit: 提交前检查代码质量
- commit-msg: 检查提交信息格式
- 只检查暂存的文件（提高效率）

---

## 🚀 第二部分：动手实践

### 练习 1: ESLint 基础（15分钟）

#### 步骤：
1. 打开 `frontend/learning-examples/eslint-practice.tsx`
2. 运行检查命令：
   ```bash
   cd frontend
   npm run lint:js
   ```
3. 观察报错信息，理解每个错误的含义
4. 尝试手动修复一些错误
5. 使用自动修复：
   ```bash
   npm run lint:fix
   ```
6. 对比自动修复前后的差异

#### 你会遇到的问题：
- ❌ 未使用的变量
- ❌ console.log 不应该出现在生产代码
- ❌ any 类型滥用
- ❌ React Hooks 依赖项缺失
- ❌ 变量命名不规范

---

### 练习 2: Prettier 格式化（10分钟）

#### 步骤：
1. 打开 `frontend/learning-examples/prettier-practice.tsx`
2. 观察代码格式混乱的地方
3. 运行格式化：
   ```bash
   npm run prettier
   ```
4. 观察文件变化，理解 Prettier 做了什么

#### 你会看到的变化：
- ✅ 双引号变单引号
- ✅ 缺失的分号被添加
- ✅ 缩进统一
- ✅ 行宽超过 100 自动换行
- ✅ 对象尾部逗号统一

---

### 练习 3: TypeScript 类型检查（20分钟）

#### 步骤：
1. 打开 `frontend/learning-examples/typescript-practice.tsx`
2. 运行类型检查：
   ```bash
   npm run tsc
   ```
3. 阅读错误信息，理解类型不匹配的原因
4. 逐个修复类型错误
5. 再次运行检查，直到没有错误

#### 你会遇到的类型错误：
- ❌ 参数类型不匹配
- ❌ 返回值类型错误
- ❌ 对象属性缺失
- ❌ null/undefined 处理不当
- ❌ 泛型使用错误

---

### 练习 4: Husky 提交钩子（15分钟）

#### 步骤：
1. 确保 Husky 已安装：
   ```bash
   cd frontend
   npm run prepare
   ```
2. 修改 `learning-examples/eslint-practice.tsx`，故意引入错误
3. 尝试提交：
   ```bash
   git add learning-examples/eslint-practice.tsx
   git commit -m "test: 测试 husky"
   ```
4. 观察 Husky 拦截提交，显示错误
5. 修复错误后再次提交

#### 体验的流程：
```
git commit 
  ↓
pre-commit 钩子触发
  ↓
lint-staged 运行（只检查暂存文件）
  ↓
ESLint 检查
  ↓
Prettier 格式化
  ↓
TypeScript 类型检查
  ↓
全部通过 → 允许提交
有错误 → 拒绝提交
```

---

## 🔧 第三部分：配置文件详解

### ESLint 配置 (`.eslintrc.js`)

```javascript
module.exports = {
  // 继承 UmiJS 官方配置
  extends: [require.resolve('@umijs/lint/dist/config/eslint')],
  
  // 全局变量
  globals: {
    page: true,
    REACT_APP_ENV: true,
  },
  
  // 自定义规则（可以添加）
  rules: {
    // 'no-console': 'warn',  // console 警告
    // '@typescript-eslint/no-unused-vars': 'error',  // 未使用变量报错
  }
};
```

**常用规则**：
- `no-console`: 禁止 console
- `no-debugger`: 禁止 debugger
- `no-unused-vars`: 禁止未使用变量
- `react-hooks/rules-of-hooks`: React Hooks 规则
- `@typescript-eslint/no-explicit-any`: 禁止 any 类型

---

### Prettier 配置 (`.prettierrc.js`)

```javascript
module.exports = {
  singleQuote: true,        // 使用单引号
  trailingComma: 'all',     // 尾部逗号
  printWidth: 100,          // 行宽 100
  proseWrap: 'never',       // 不换行
  endOfLine: 'lf',          // 换行符 LF
};
```

**配置说明**：
- `singleQuote`: 单引号 vs 双引号
- `semi`: 是否加分号
- `tabWidth`: 缩进宽度
- `trailingComma`: 尾部逗号（all/es5/none）
- `printWidth`: 每行最大字符数

---

### TypeScript 配置 (`tsconfig.json`)

```json
{
  "compilerOptions": {
    "strict": true,                    // 严格模式
    "target": "esnext",                // 编译目标
    "module": "esnext",                // 模块系统
    "jsx": "preserve",                 // JSX 处理
    "esModuleInterop": true,           // 模块互操作
    "skipLibCheck": true,              // 跳过库检查
    "baseUrl": "./",                   // 基础路径
    "paths": {                         // 路径映射
      "@/*": ["./src/*"]
    }
  }
}
```

**重要选项**：
- `strict`: 启用所有严格类型检查
- `noImplicitAny`: 禁止隐式 any
- `strictNullChecks`: 严格 null 检查
- `noUnusedLocals`: 检查未使用的局部变量

---

### Husky 配置

**pre-commit** (`.husky/pre-commit`):
```bash
#!/bin/sh
npx --no-install lint-staged
```

**lint-staged** (`package.json`):
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

## 🎓 第四部分：进阶配置

### 任务 1: 添加自定义 ESLint 规则

编辑 `frontend/.eslintrc.js`，添加：

```javascript
module.exports = {
  extends: [require.resolve('@umijs/lint/dist/config/eslint')],
  globals: {
    page: true,
    REACT_APP_ENV: true,
  },
  rules: {
    // 警告级别的 console
    'no-console': ['warn', { allow: ['warn', 'error'] }],
    
    // 禁止使用 any
    '@typescript-eslint/no-explicit-any': 'error',
    
    // 未使用变量报错
    '@typescript-eslint/no-unused-vars': ['error', {
      argsIgnorePattern: '^_',  // 忽略 _开头的参数
    }],
    
    // React Hooks 依赖检查
    'react-hooks/exhaustive-deps': 'warn',
  },
};
```

### 任务 2: 配置 VSCode 自动格式化

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

### 任务 3: 添加提交信息检查

编辑 `.husky/commit-msg`，确保提交信息符合规范：
- feat: 新功能
- fix: 修复
- docs: 文档
- style: 格式
- refactor: 重构
- test: 测试
- chore: 构建

---

## 📝 第五部分：常见问题解决

### 问题 1: ESLint 和 Prettier 冲突

**现象**：ESLint 要求加分号，Prettier 自动删除分号

**解决**：使用 `eslint-config-prettier` 禁用冲突规则
```bash
npm install --save-dev eslint-config-prettier
```

在 `.eslintrc.js` 中：
```javascript
extends: [
  require.resolve('@umijs/lint/dist/config/eslint'),
  'prettier',  // 放在最后
],
```

### 问题 2: Husky 钩子不生效

**解决步骤**：
```bash
cd frontend
rm -rf .husky
npm run prepare
chmod +x .husky/*
```

### 问题 3: TypeScript 报错太多

**临时方案**：在 `tsconfig.json` 中降低严格度
```json
{
  "compilerOptions": {
    "strict": false,
    "noImplicitAny": false
  }
}
```

**推荐方案**：逐步修复，不要降低标准

---

## 🎯 第六部分：实战检验

### 最终测试：修改真实代码

1. 打开 `frontend/src/pages/Dashboard/index.tsx`
2. 故意引入几个问题：
   - 添加未使用的变量
   - 使用 any 类型
   - 删除必要的类型注解
   - 格式化混乱
3. 运行完整检查：
   ```bash
   npm run lint
   ```
4. 修复所有问题
5. 提交代码，体验完整流程

---

## 📚 学习资源

- [ESLint 官方文档](https://eslint.org/docs/latest/)
- [Prettier 官方文档](https://prettier.io/docs/en/)
- [TypeScript 官方文档](https://www.typescriptlang.org/docs/)
- [Husky 官方文档](https://typicode.github.io/husky/)

---

## ✅ 学习检查清单

完成后，你应该能够：

- [ ] 理解 ESLint、Prettier、TypeScript 的区别和作用
- [ ] 能够读懂 ESLint 错误信息并修复
- [ ] 能够配置 Prettier 格式化规则
- [ ] 能够理解 TypeScript 类型错误并修复
- [ ] 能够配置 Husky 提交钩子
- [ ] 能够自定义 ESLint 规则
- [ ] 能够解决工具冲突问题
- [ ] 能够在团队中推广代码规范

---

## 🚀 下一步

学完这些后，你可以：
1. 在简历中写：熟练使用 ESLint + Prettier + TypeScript + Husky 保障代码质量
2. 面试时能够讲解每个工具的作用和配置
3. 在新项目中从零搭建代码规范体系
4. 优化现有项目的代码质量工具链

---

**现在开始第一个练习吧！** 👇
