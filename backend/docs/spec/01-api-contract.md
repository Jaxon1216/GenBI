# 01 · API 契约（API Contract）

> 所有接口挂在 `/api` 前缀下，服务监听 `:8080`（前端 dev 代理 `config/proxy.ts` 将 `/api` 转发到 `http://localhost:8080`）。

## 全局响应封装

所有接口返回统一封装（对齐 Java `BaseResponse`/`ResultUtils`）：

```json
{ "code": 0, "data": <T>, "message": "ok" }
```

- `code === 0` 表示成功。
- 前端页面主要判断 `res.data` 是否存在，并用 `try/catch` 兜底。
- 出错时 `data` 为 `null`，`code`/`message` 取自错误码表。**HTTP 状态码统一返回 200**（业务错误码放在 body 里，与 Java 行为一致）。

## 错误码表（沿用 Java `ErrorCode`）

| code | 含义 |
|---|---|
| 0 | ok |
| 40000 | 请求参数错误 |
| 40100 | 未登录 |
| 40101 | 无权限 |
| 40300 | 禁止访问 |
| 40400 | 请求数据不存在 |
| 42900 | 提交过于频繁（限流） |
| 50000 | 系统内部异常 |
| 50001 | 操作失败 |

## 与 Java `Genbi-backend` 的 3 处纠偏（重要）

1. **同步生成返回字段**：前端 `AddChart` 读 `res.data.genChart`，故 `/chart/gen` 的 `BiResponse` 用 **`genChart`**（Java 误用了 `genChartStr`）。
2. **状态值**：前端只识别 `wait / running / succeed / failed`，故异步流程中间态用 **`running`**（Java 误用了 `processing`，前端不识别）。
3. **Chart 字段名**：采用前端口径 **`name` / `execMessage`**（Java 实体用 `chartName` / `Message`）。

---

## 用户模块 `/api/user`

| 方法 | 路径 | 入参 | 出参 `data` | 鉴权 | 优先级 |
|---|---|---|---|---|---|
| POST | /register | `{userAccount,userPassword,checkPassword}` | `userId:number` | 否 | MUST |
| POST | /login | `{userAccount,userPassword}` | `LoginUserVO` + Set-Cookie | 否 | MUST |
| POST | /logout | — | `boolean` | 是 | MUST |
| GET | /get/login | — | `LoginUserVO` | 是 | MUST |
| POST | /add | `UserAddRequest` | `userId` | admin | SHOULD |
| POST | /delete | `{id}` | `boolean` | admin | SHOULD |
| POST | /update | `UserUpdateRequest` | `boolean` | admin | SHOULD |
| POST | /update/my | `UserUpdateMyRequest` | `boolean` | 登录 | SHOULD |
| GET | /get | `?id=` | `User` | admin | SHOULD |
| GET | /get/vo | `?id=` | `UserVO` | 登录 | SHOULD |
| POST | /list/page | `UserQueryRequest` | `Page<User>` | admin | SHOULD |
| POST | /list/page/vo | `UserQueryRequest` | `Page<UserVO>` | 登录 | SHOULD |

**校验规则（对齐 Java）**：账号长度 ≥ 4，密码长度 ≥ 8，两次密码一致；账号唯一。密码存储 `md5Hex("tanter" + 明文)`。

**LoginUserVO**：`{ id, userName, userAvatar, userProfile, userRole, createTime, updateTime }`（脱敏，不含密码）。

## 图表模块 `/api/chart`

| 方法 | 路径 | 入参 | 出参 `data` | 鉴权 | 优先级 |
|---|---|---|---|---|---|
| POST | /add | `ChartAddRequest{name,goal,chartData,chartType}` | `chartId:number` | 登录 | MUST |
| POST | /delete | `{id}` | `boolean`（本人或 admin） | 登录 | MUST |
| GET | /get | `?id=` | `Chart` | 登录 | MUST |
| POST | /list/page | `ChartQueryRequest` | `Page<Chart>`（size≤20） | 登录 | MUST |
| POST | /my/list/page | `ChartQueryRequest` | `Page<Chart>`（限本人，size≤20） | 登录 | MUST |
| POST | /gen | multipart：`file` + form/query `goal,name,chartType` | `BiResponse{chartId,genChart,genResult}` | 登录+限流 | MUST |
| POST | /gen/async | 同上 | `chartId:number` | 登录+限流 | MUST |
| POST | /update | `ChartUpdateRequest` | `boolean` | admin | SHOULD |
| POST | /edit | `ChartEditRequest` | `boolean` | 本人 | SHOULD |

**ChartQueryRequest**：`{ current, pageSize, name, goal, chartType, id, userId, sortField, sortOrder }`。

**Chart**（返回给前端的字段）：`{ id, userId, name, goal, chartData, chartType, genChart, genResult, status, execMessage, createTime, updateTime, isDelete }`。

**分页 Page<T>**：至少包含 `{ records: T[], total, current, size }`（前端读 `records` 与 `total`）。

**限流**：`/gen` 与 `/gen/async` 按 `gen_chart:{userId}` 令牌桶 3/秒，超限返回 `42900`。

**文件校验**：仅接受 `.xlsx` / `.csv`；其余（含老式 `.xls`）返回 `40000`。

## 通用

| 方法 | 路径 | 出参 | 说明 |
|---|---|---|---|
| GET | /api/health | `{code:0,data:"ok",message:"ok"}` | 存活探针 |
