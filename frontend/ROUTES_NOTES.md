# Ant Design Pro 路由配置笔记

## 问题记录

**问题**：打开网站后侧边栏没有显示任何菜单项。

**原因**：Ant Design Pro 的 `ProLayout` 组件会根据 `config/routes.ts` 自动生成侧边栏菜单，但**只有配置了 `name` 字段的路由才会显示在侧边栏**。没有 `name` 的路由仍然可以通过 URL 正常访问，只是不会出现在菜单中。

**解决**：给需要显示在侧边栏的路由添加 `name` 字段。

---

## 路由配置语法

文件路径：`config/routes.ts`

```ts
export default [
  // 1. 基础路由
  { path: '/welcome', component: './Welcome' },
  //   path: URL 路径
  //   component: 对应 src/pages/ 下的组件文件

  // 2. 带菜单名称的路由（会显示在侧边栏）
  { name: '欢迎页面', path: '/welcome', icon: 'smile', component: './Welcome' },
  //   name: 侧边栏显示的文字 ← 这是菜单是否显示的关键
  //   icon: 侧边栏图标（使用 antd 图标名）

  // 3. 不使用全局布局的路由（没有侧边栏和顶栏）
  { path: '/user', layout: false, routes: [...] },
  //   layout: false → 该页面不显示侧边栏/顶栏，适用于登录页

  // 4. 嵌套路由（子菜单）
  {
    path: '/admin',
    icon: 'crown',
    name: '管理页面',
    access: 'canAdmin',       // 权限控制：只有 canAdmin 权限的用户能看到
    routes: [
      { path: '/admin', redirect: '/admin/sub-page' },  // 重定向
      { path: '/admin/sub-page', name: '子页面', component: './Admin' },
    ],
  },

  // 5. 重定向
  { path: '/', redirect: '/welcome' },
  //   访问 / 时自动跳转到 /welcome

  // 6. 404 兜底路由（必须放最后）
  { path: '*', layout: false, component: './404' },
];
```

## 各字段说明

| 字段        | 类型    | 作用                                                              |
| ----------- | ------- | ----------------------------------------------------------------- |
| `path`      | string  | URL 路径                                                          |
| `name`      | string  | **侧边栏菜单名称，不设置则不显示在菜单中**                        |
| `icon`      | string  | 侧边栏图标，值为 antd Icon 组件名（如 `smile`, `crown`, `table`） |
| `component` | string  | 页面组件路径，相对于 `src/pages/`                                 |
| `layout`    | boolean | 是否使用全局布局，`false` 表示独立页面（无侧边栏）                |
| `routes`    | array   | 子路由，用于嵌套菜单                                              |
| `redirect`  | string  | 重定向目标路径                                                    |
| `access`    | string  | 权限标识，对应 `src/access.ts` 中定义的权限函数                   |

## 当前项目路由结构

```
/user/login      → 登录页（layout: false，独立页面）
/welcome         → 欢迎页面
/admin           → 管理页面（需要 canAdmin 权限）
  /admin/sub-page → 管理子页面
/list            → 表格页面
/                → 重定向到 /welcome
*                → 404 页面
```

整个链路很清楚了。让我给你解释：

这个项目用的是 **Umi.js**（蚂蚁金服出品的 React 框架），Ant Design Pro 是基于 Umi 的上层模板。

整个流程是这样的：

**1. `config/routes.ts`** — 定义路由数组

```1:17:session/session1/yubi-frontend-1/config/routes.ts
export default [
  { name:'登录',path: '/user', layout: false, routes: [{ path: '/user/login', component: './User/Login' }] },
  // ...
];
```

**2. `config/config.ts`** — 主配置文件，导入并使用 routes

```6:6:session/session1/yubi-frontend-1/config/config.ts
import routes from './routes';
```

```28:29:session/session1/yubi-frontend-1/config/config.ts
  // umi routes: https://umijs.org/docs/routing
  routes,
```

这里 `routes` 作为 `defineConfig()` 的一个配置项传给 Umi 框架。

**3. Umi 框架拿到 routes 后做两件事：**

- **生成路由映射**：根据 `path` 和 `component` 建立 URL → 页面组件的对应关系（比如访问 `/welcome` 就渲染 `src/pages/Welcome` 组件）
- **生成侧边栏菜单**：`layout` 插件（第 76-80 行配置的）会遍历 routes，把**有 `name` 字段的路由**渲染成侧边栏菜单项

**4. 同理，`proxy.ts` 也是这样被使用的：**

```5:5:session/session1/yubi-frontend-1/config/config.ts
import proxy from './proxy';
```

```54:54:session/session1/yubi-frontend-1/config/config.ts
  proxy: proxy[REACT_APP_ENV as keyof typeof proxy],
```

这里根据环境变量 `REACT_APP_ENV` 来选择用哪套代理配置（`dev`/`test`/`pre`）。

**总结一下调用链：**

```
config/routes.ts  ──┐
config/proxy.ts   ──┤
config/defaultSettings.ts ──┤──→ config/config.ts (defineConfig) ──→ Umi 框架
                            │         ↓                    ↓
                            │    生成路由映射          ProLayout 插件
                            │    (URL→组件)          (routes→侧边栏菜单)
```

所以 `routes.ts` 不需要你在代码里手动 import，它通过 `config.ts` 交给 Umi 框架，框架自动处理路由匹配和菜单生成。
