# 问题解决记录

---

### npm run openapi 执行失败，连接被拒绝 (ECONNREFUSED)

**日期**：2026-03-05

**问题描述**：
运行 `npm run openapi` 报错 `FetchError: request to http://localhost:8101/api/v2/api-docs failed, reason: ECONNREFUSED`，无法生成前端接口代码。

**原因分析**：
两个原因：1) openapi 插件配置的 `schemaPath` 端口是 8101，但后端 dev 环境实际运行在端口 12345；2) 后端服务需要先启动才能拉取 API 文档。

**解决方案**：
1. 将 `config/config.ts` 中 `schemaPath` 从 `http://localhost:8101/api/v2/api-docs` 改为 `http://localhost:12345/api/v2/api-docs`
2. 先通过 `mvn spring-boot:run` 启动后端，再运行 `npm run openapi`

**相关文件**：
- `yubi-frontend-1/config/config.ts`

---

### Knife4j 测试 post 接口返回"系统错误" (code: 50000)

**日期**：2026-03-05

**问题描述**：
在 Knife4j 接口文档页面测试 `POST /api/post/add` 接口，返回 `{"code":50000,"data":null,"message":"系统错误"}`。

**原因分析**：
项目是基于 Spring Boot 万用模板开发的 BI 项目，建表 SQL 只有 `user` 和 `chart` 两张表，没有 `post` 表。模板自带的 post 相关 controller 操作不存在的表导致 RuntimeException。

**解决方案**：
这是正常现象，post 相关接口是模板遗留的，BI 项目只需关注 `user-controller` 和 `chart-controller`。

**相关文件**：
- `yubi-backend-1/sql/create_table.sql`

---

### npm run start 能登录但 npm run dev 不能登录

**日期**：2026-03-05

**问题描述**：
使用 `npm run start` 可以用 `admin/ant.design` 登录，切换到 `npm run dev` 后同样的账号密码无法登录。

**原因分析**：
三层原因：
1. `npm run start` 开启了 Mock，`admin/ant.design` 是 Mock 假数据中的预设账号
2. `npm run dev`（即 `start:dev`）设置了 `MOCK=none` 关闭 Mock，请求通过 proxy 转发到真实后端
3. 前端登录页调用的是 Mock 接口 `POST /api/login/account`，而后端真实接口是 `POST /api/user/login`，路径完全不匹配；表单字段名也不同（Mock 用 `username/password`，后端用 `userAccount/userPassword`）

**解决方案**：
1. 运行 `npm run openapi` 生成后端真实接口代码到 `src/services/yubi/`
2. 修改登录页 `Login/index.tsx`：引入 `userLoginUsingPost` 替换 Mock 的 `login`，表单字段改为 `userAccount` 和 `userPassword`，登录成功判断从 `msg.status === 'ok'` 改为 `res.code === 0`
3. 修改 `app.tsx`：引入 `getLoginUserUsingGet` 替换 Mock 的 `queryCurrentUser`，类型从 `API.CurrentUser` 改为 `API.LoginUserVO`，字段名从 `avatar/name` 改为 `userAvatar/userName`
4. 修改 `AvatarDropdown.tsx`：退出登录接口改为 `userLogoutUsingPost`，用户名字段改为 `userName`

**相关文件**：
- `yubi-frontend-1/src/pages/User/Login/index.tsx`
- `yubi-frontend-1/src/app.tsx`
- `yubi-frontend-1/src/components/RightContent/AvatarDropdown.tsx`
- `yubi-frontend-1/src/services/yubi/userController.ts`
- `yubi-frontend-1/config/proxy.ts`

---

### 后端启动报端口 12345 被占用

**日期**：2026-03-05

**问题描述**：
再次运行 `mvn spring-boot:run` 报 `Port 12345 was already in use`。

**原因分析**：
之前已经启动过后端服务，Java 进程还在运行中占用着端口。

**解决方案**：
不需要重复启动。用 `lsof -i :12345` 确认是自己的 Java 进程即可。如需重启，先 `kill <PID>` 再重新启动。

**相关文件**：
- `yubi-backend-1/src/main/resources/application.yml`
