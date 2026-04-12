# 前端工程化：学习笔记 + 面试话术

聚焦 **ESLint、Prettier、TypeScript、Husky、Docker、GitHub Actions** 的职责与协作关系。下文以本仓库（Ant Design Pro + Umi Max + 独立后端）为**架构参照**，不讲具体业务 bug。

---

## 一、整体链路（先建立心智模型）

```
本地编辑
  → ESLint / Prettier / tsc（编辑器 + 命令行）
  → git commit：Husky + lint-staged（只检查暂存文件）
  → git push：GitHub Actions（固定 Node 版本下再跑 lint + build）
  → 部署：Docker 构建镜像，Nginx 提供静态资源与网关能力
```

目标：**质量可重复、协作有门禁、环境可复现**。

---

## 二、ESLint：约束「怎么写才对」

**做什么**：把源码解析成 AST，按规则检查语义与最佳实践（未使用变量、Hooks 规则、TS 相关规则等）。

**和 Prettier 的分工**：ESLint 管**质量与约定**；不管「缩进用几个空格」这类纯格式（交给 Prettier）。

**本仓库接法**：`extends` Umi 官方封装，减少从零配规则的成本。

```javascript
// .eslintrc.js（节选）
module.exports = {
  extends: [require.resolve('@umijs/lint/dist/config/eslint')],
  globals: { page: true, REACT_APP_ENV: true },
};
```

**常用命令**：

```bash
npm run lint:js
npm run lint:fix
```

**面试一句话**：ESLint 在本地和 CI 里统一团队规则，把可维护性问题和常见错误前置到写代码/提交阶段。

---

## 三、Prettier：统一「长什么样」

**做什么**：按配置文件重排代码样式（引号、换行、尾逗号等），**不**做类型或逻辑推断。

**本仓库**：配置在 `.prettierrc.js`（如单引号、`printWidth`）。

```bash
npm run prettier
npm run lint:prettier
```

**与 ESLint 并存**：用 `eslint-config-prettier` 关闭与 Prettier **重叠**的 ESLint 格式规则，避免同一类问题被两个工具各管一半。

**面试一句话**：Prettier 消灭格式争论，CR 时只看语义；ESLint 管对错，Prettier 管好看。

---

## 四、TypeScript：`tsc --noEmit`

**做什么**：在**不产出 JS** 的前提下做全量类型检查，与「能打包成功」独立——有的项目会把 `tsc` 单独放进 `npm run lint` 或 CI。

**本仓库**：`npm run lint` = `lint:js` + `lint:prettier` + `tsc`。

**面试一句话**：构建工具链有时对类型较宽松；单独跑 `tsc` 是把类型契约固化到流水线里的一种常见做法。

---

## 五、Husky + lint-staged：提交门禁

**Husky**：管理 Git hooks（如 `pre-commit`），在特定 Git 事件时执行脚本。

**lint-staged**：只对 **git 暂存区**涉及的文件跑命令，避免每次全仓库扫描，速度适合日常提交。

**本仓库接线**：

- `package.json` 里 `prepare` 安装 husky 到 `frontend/.husky`
- `lint-staged` 配置对 `*.{ts,tsx,...}` 跑 ESLint，对多类扩展名跑 Prettier
- `pre-commit` 里执行 `lint-staged`

**面试要点**：

- 为什么用 lint-staged：**增量**检查，大仓库友好。
- CI 为什么还要再跑一遍：本地可跳过 hook；远程 runner 环境统一，是团队共同底线。

---

## 六、Docker：构建可复现、运行环境隔离

**前端为什么常见 Node 构建 + Nginx 运行**：

1. `npm run build` 产出的是**静态资源**（`dist/`），生产环境需要 HTTP Server；Nginx 适合托管静态文件。
2. **多阶段构建**：一阶段用 Node 安装依赖并 build，二阶段只拷贝 `dist` 到 `nginx` 镜像，减小体积、减少攻击面。

**SPA + 网关层面的两件标配**（与是否用 Umi/React 无关，属于部署常识）：

1. **History 路由**：服务器上对非文件路径回退到 `index.html`（如 `try_files ... /index.html`），由前端路由接管路径。
2. **同源 `/api`**：开发时由 devServer **proxy**；生产里没有 dev server 时，由 **Nginx（或网关）反向代理**到后端服务，浏览器仍请求当前站点下的 `/api`。

```nginx
# 概念示例：反代 + SPA 回退（具体以仓库 nginx.conf 为准）
location /api/ {
    proxy_pass http://backend:8081;
}
location / {
    try_files $uri $uri/ /index.html;
}
```

**Compose**：多容器编排（前端、后端、数据库、缓存等）在同一网络内用**服务名**互访，宿主机通过端口映射访问。

**面试一句话**：Docker 解决「在我机器上能跑」；Compose 描述多服务拓扑；前端镜像里 Nginx 负责静态资源与反向代理是常见生产形态。

---

## 七、GitHub Actions：远程 CI/CD

**本仓库常见模式**（见 `.github/workflows/frontend-ci.yml`）：

- **触发范围**：`paths: frontend/**`，仅前端变更时跑前端任务，节省 runner 时间。
- **阶段划分**：先 `lint`（含 ESLint / Prettier / tsc），再 `build`；可用 `needs` 串联 job。
- **环境固定**：`actions/setup-node` 指定 Node 版本，与本地 `engines` / 团队约定一致。

**面试一句话**：Actions 是仓库里的可版本化流水线；`paths` 做变更检测，`needs` 表达依赖顺序，和本地 Husky 互补而不是重复废话。

---

## 八、工具对照表（背诵用）

| 工具             | 解决什么             | 典型出现在      |
| ---------------- | -------------------- | --------------- |
| ESLint           | 代码质量与约定       | 本地、husky、CI |
| Prettier         | 代码风格一致         | 本地、husky、CI |
| TypeScript / tsc | 类型契约             | 本地、CI        |
| Husky            | Git 钩子自动化       | 本地 commit     |
| lint-staged      | 只检查暂存文件       | pre-commit      |
| Docker           | 环境一致、交付物镜像 | 构建/部署       |
| GitHub Actions   | 远程自动化与门禁     | push / PR       |

---

## 九、高频面试简答（纯工程化）

1. **ESLint 和 Prettier 区别？**  
   ESLint：逻辑与最佳实践；Prettier：格式化。配合 `eslint-config-prettier` 避免规则冲突。

2. **为什么要 husky + lint-staged？**  
   提交前自动执行检查；lint-staged 控制范围，缩短反馈时间。

3. **CI 里已有 lint，本地还要 husky 吗？**  
   本地更早反馈；CI 防绕过 hook、统一环境。二者互补。

4. **生产环境为什么常用 Nginx 托管前端？**  
   静态资源高效；可顺带做压缩、缓存、路由回退与 API 反向代理。

5. **GitHub Actions 的 `paths` 有什么意义？**  
   路径过滤，无关变更不跑对应 job，节省时间与成本。

---

## 十、简历可用表述（按需裁剪）

- 使用 **ESLint + Prettier + TypeScript** 建立代码规范与类型门禁，`npm run lint` 串联检查命令。
- 通过 **Husky + lint-staged** 在提交阶段对变更文件执行 lint 与格式化。
- 使用 **Docker 多阶段构建** 前端镜像，**Nginx** 托管构建产物并配置 **SPA 与 API 转发**；配合 **Docker Compose** 编排多服务。
- 使用 **GitHub Actions** 在 PR/主干上自动执行前端 lint、tsc 与 build，并用 **paths** 限制触发范围。

---

## 十一、本仓库命令速查

```bash
cd frontend
npm run lint:js
npm run lint:fix
npm run lint:prettier
npm run tsc
npm run lint
npm run prepare    # 安装 husky
```

```bash
# 根目录（若使用 Compose）
docker-compose up -d --build
```

---

文档只覆盖**工程化工具链**与**通用部署知识**；具体端口、环境变量以当前仓库 `docker-compose.yml`、`Dockerfile`、`nginx.conf` 为准。
