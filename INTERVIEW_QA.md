# GenBI 智能数据分析平台 — 面试 Q&A 完整版

> 答案遵循 STAR 法则（场景 → 归因 → 动作 → 结果），体现排错能力和总结能力。
> 末尾附【数据验证清单】，标注哪些量化数据需要实测。

---

## 一、项目总览

### Q1: 说一下本项目的完整业务流程

**A:**

这是一个 AI 驱动的智能 BI 平台，核心解决的问题是：**让非技术人员也能通过上传 Excel 得到专业的数据可视化图表**。

完整链路分 5 步：

1. **用户上传**：前端表单提交分析目标 + Excel 文件，通过 Umi OpenAPI 生成的 service 层发送 multipart 请求到后端
2. **后端处理**：Spring Boot 将 Excel 转为 CSV，拼接 system prompt + 用户数据，调用百度千帆 OpenAI 兼容 API
3. **AI 生成**：大模型返回用特殊分隔符（`【【【【`）分割的两段内容——ECharts option JSON 和文字分析结论
4. **前端渲染**：前端拿到 JSON 字符串后经过 Zod Schema 校验 → sanitize 过滤 → SafeChart 组件渲染，外层包裹 Error Boundary 兜底
5. **看板展示**：用户可以在拖拽看板中组合多个图表，自由布局并持久化

其中同步模式直接等 AI 返回结果；异步模式后端先入库一条 `wait` 状态的记录，用线程池异步执行 AI 调用，前端通过轮询感知状态变化。

---

## 二、核心亮点深入（大概率会问）

---

### 亮点 1：Zod Schema 校验层

#### Q2: 你在项目中为什么要做 Schema 校验？遇到了什么问题？

**A:**

**场景**：上线初期，AI 生成的图表有大约 10%-15% 的概率出现页面白屏或报错。我打开 DevTools 排查，发现 ECharts 的 `option` 传入了 `undefined` 或者 JSON 格式有问题——AI 有时候会在 JSON 前面加 ` ```json` 标记，有时候 `series` 字段直接缺失。

**归因**：根本原因是我们把 AI 当作一个"稳定的 API"来用了，直接 `JSON.parse(res.genChart)` 然后喂给 ECharts，中间没有任何防御层。AI 本质上是一个不确定性输出源。

**动作**：我设计了一套三层防御机制：
1. **JSON 容错提取**：AI 返回非标准 JSON 时，用 `extractJsonObject` 找到第一个 `{` 和最后一个 `}` 的子串再解析
2. **Zod Schema 校验**：定义 `EChartsOptionSchema`，要求 `series` 必须是非空数组、`type` 必须是合法图表类型，用 `safeParse` 做安全校验
3. **sanitize 过滤**：在 `JSON.stringify` 的 replacer 中检测含 `function` 或 `=>` 的字符串值并移除，防止 XSS

最外层再包一个 Error Boundary，即使 Schema 校验通过但 ECharts 渲染时仍然报错（比如 data 值不合理），也能优雅降级。

**结果**：图表渲染崩溃率降为 0，任何异常格式都有对应的 fallback 展示。

#### Q2.1 追问: 为什么选 Zod 不选 JSON Schema 或手写校验？

**A:** 三个原因：
1. **TS 类型推导**：`z.infer<typeof Schema>` 直接推导出类型，不用手维护两份代码。JSON Schema 需要 ajv + json-schema-to-typescript 两个库才能做到。
2. **safeParse 不抛异常**：返回 `{ success, data/error }` 联合类型，天然适合"校验失败走兜底"的流程，不需要 try/catch。
3. **体积小**：gzip 约 13KB，比 ajv 轻量很多，对前端 bundle size 友好。

#### Q2.2 追问: Error Boundary 为什么只能用 Class 组件？它能捕获什么？

**A:** `getDerivedStateFromError` 和 `componentDidCatch` 是 Class 组件独有的生命周期，React 团队至今没有提供对应的 Hook。

能捕获的：子组件渲染阶段的同步异常（render 方法 / JSX 中的报错）。

不能捕获的：
- 事件处理函数中的异常（需要 try/catch）
- 异步代码（setTimeout、Promise.reject）
- Error Boundary 自身的异常

在我的项目中，ECharts 组件在接收到格式正确但值不合理的 option 时会在渲染阶段抛错，Error Boundary 正好能兜住这种场景。

---

### 亮点 2：前端轮询（指数退避）

#### Q3: 你为什么要做轮询？怎么发现需要优化的？

**A:**

**场景**：异步模式下，用户提交后跳转到"我的图表"页面，但图表卡在 `wait` 或 `running` 状态，用户需要手动刷新才能看到结果。这个体验很差。

**归因**：异步接口提交后前端没有任何主动获取状态更新的机制。最简单的方案是定时轮询，但如果用固定 3 秒间隔，一个 30 秒完成的任务在生成完后还会继续发请求——我在 Network 面板观察到即使所有图表都已经 succeed，请求还在不停发。

**动作**：封装了 `usePolling` Hook，做了三层优化：
1. **指数退避**：`interval = min(current × 1.5, 30000)`，数据没变化就逐渐放慢，有变化立刻重置为 3 秒
2. **Page Visibility API**：`visibilitychange` 事件监听，用户切走标签页暂停轮询，切回来立即请求一次并重置间隔
3. **自动启停**：用 `useMemo` 计算 `hasPendingCharts`（列表中是否有 wait/running 的图表），没有就完全停止轮询

**结果**：相比固定 3 秒间隔，在一个典型 30 秒任务周期内请求数从约 10 次降到约 4 次，减少约 60%。

#### Q3.1 追问: useRef 和 useState 什么时候用哪个？轮询里为什么用 ref？

**A:** 核心区别：**useState 的更新会触发重渲染，useRef 不会**。

在轮询 Hook 中：
- `timerRef`（定时器 ID）：只是为了清除定时器，不需要触发 UI 更新
- `currentIntervalRef`（当前间隔）：频繁变化，如果用 state 每次退避都会触发整个组件重渲染，浪费性能
- `isMountedRef`（组件是否存活）：用于防止卸载后 setState 报警告
- `previousDataRef`（上次数据快照）：对比用，不需要反映到 UI

总结：**跟 UI 展示有关的用 state，跟副作用/内部逻辑有关的用 ref**。

#### Q3.2 追问: useEffect 的清理函数什么时候执行？不清理会怎样？

**A:** 两个时机：
1. **依赖变化时**：先执行上一次的 cleanup，再执行新的 effect
2. **组件卸载时**：执行最后一次的 cleanup

在轮询中如果不 `clearTimeout`，组件卸载后定时器继续运行，回调里的 `setState` 会尝试更新一个已卸载的组件，React 会报 "Can't perform a state update on an unmounted component" 警告。更严重的是，多次切换页面会积累多个定时器同时运行，造成请求洪峰。

---

### 亮点 3：虚拟列表

#### Q4: 虚拟列表是怎么想到要做的？怎么排查出性能问题的？

**A:**

**场景**：用户上传的 Excel 可能有几千行数据，我们有一个"查看原始数据"的功能，最初用普通 `map` 渲染所有行。有一次测试用 3000 行数据点击查看，Modal 弹出后明显感觉到 1-2 秒的卡顿，滚动也不流畅。

**排查**：我打开 Chrome DevTools 的 Performance 面板录制了一次，发现 Rendering 阶段耗了 2000ms+，Long Task 占满了主线程。切到 Elements 面板一看，DOM 树里有 3000 个行节点——这就是瓶颈所在。

**归因**：浏览器对大量 DOM 节点的布局计算（Layout）和绘制（Paint）开销是线性增长的，3000 个节点每次布局都要逐一计算位置。

**动作**：引入 react-window 的 `FixedSizeList`，核心原理是**只渲染可视区域内的约 15 个 DOM 节点**：
- 容器固定高度 + `overflow: auto` 产生滚动条
- 内部一个占位元素 `height = itemCount × itemSize` 撑起真实滚动高度
- 每行用 `position: absolute; top: index × 40px` 定位
- 滚动时根据 `scrollTop` 计算可见范围，只挂载/卸载进出视口的行

因为行高固定（40px），定位是 O(1) 的，不需要测量。

**结果**：DOM 节点从 3000 个降到约 15 个，弹窗打开从 2 秒卡顿变为瞬间响应，滚动帧率稳定 60fps。

#### Q4.1 追问: 如果行高不固定怎么办？

**A:** 用 `VariableSizeList`，传入 `itemSize` 函数 `(index) => height`。它内部维护一个前缀和数组来缓存每行的偏移量，定位需要二分查找是 O(log n)。首次渲染时需要预估行高，滚动过程中动态修正。

在我的项目中数据都是 CSV 单行文本，高度固定，所以用 `FixedSizeList` 就够了。

#### Q4.2 追问: 虚拟列表快速滚动会白屏吗？怎么优化？

**A:** 会。极速滚动时新行的渲染跟不上滚动速度，会出现短暂空白。react-window 默认有 `overscanCount` 参数可以设置预渲染行数（比如在可视范围上下各多渲染 5 行），用空间换时间。

更极端的场景（比如飞书文档的大表格），可以用 **Canvas 渲染**替代 DOM——不需要创建 DOM 节点，直接在画布上绘制单元格内容，彻底绕开 DOM 的布局和绘制开销。但实现复杂度高很多，需要自己处理文本测量、事件委托、选区等。

---

### 亮点 4：拖拽看板

#### Q5: react-grid-layout 的布局系统是怎么工作的？

**A:**

核心是一个 **12 列网格系统**，每个 item 用 `{ i, x, y, w, h }` 描述位置：
- `i`：唯一标识
- `x, y`：左上角在网格中的坐标（列/行）
- `w, h`：占几列、几行

拖拽时有**碰撞检测和自动压缩**——如果 A 移到 B 的位置，B 会被自动推到下方。

响应式通过 `Responsive` 组件实现，在不同 breakpoint 下切换列数（lg: 12, md: 10, sm: 6）。`WidthProvider` 是一个 HOC，监听容器 resize 事件把宽度传给 grid 组件计算列宽。

我用 `draggableHandle=".drag-handle"` 限制了只有卡片标题栏能触发拖拽，避免用户在图表区域操作时误触拖拽。

#### Q5.1 追问: 布局持久化怎么做的？有什么局限？

**A:**

当前用 `localStorage` 存储 `{ items, layouts }` JSON 对象。优点是简单无后端依赖，缺点是只在本地浏览器有效，清缓存会丢失，不能跨设备同步。

如果要做得更好：后端新增 `dashboard` 表，存看板名称 + layout JSON + chart IDs，用 debounce 在拖拽结束后自动保存，支持多看板切换。但当时后端不是我负责，就先用 localStorage 实现了 MVP。

---

### 亮点 5：骨架屏 + 路由懒加载

#### Q6: 骨架屏和 Loading Spinner 有什么区别？为什么用骨架屏？

**A:**

**场景**：我的图表页面首次加载时，原来用 Ant Design 的 List loading（一个 Spinner），数据返回后内容突然出现，页面会"跳一下"。

**归因**：这就是 CLS（Cumulative Layout Shift）——Spinner 占位和实际内容的高度不同，内容出现时挤开了其他元素。

**动作**：用 Ant Design 的 Skeleton 组件组合（Avatar + Input + Node）模拟真实卡片布局，确保骨架屏和最终内容的尺寸一致，加载完成后"原地替换"。同时区分**首次加载**（没有旧数据，显示骨架屏）和**翻页加载**（有旧数据，用 List 自带的半透明 overlay）。

Umi 4 默认对 config routes 做代码分割，每个路由是独立 chunk，`src/loading.tsx` 是约定的全局加载组件，路由 chunk 下载期间自动展示。

**结果**：首屏 JS 体积减少约 40%（因为不再一次性加载所有页面的代码），CLS 消除。

---

## 三、通用前端问题（面试鸭题目）

---

#### Q7: React 框架的优势和适用场景？

**A:** 五个核心优势：
1. **组件化**：UI 拆分为独立可复用的组件，比如我项目里的 SafeChart、ChartErrorBoundary 都可以跨页面复用
2. **虚拟 DOM**：通过 diff 算法最小化真实 DOM 操作，比直接操作 DOM 性能更好
3. **单向数据流**：数据从父组件流向子组件，状态变化可预测，调试方便
4. **生态丰富**：Ant Design、ECharts-for-React、react-window 等成熟库直接可用
5. **跨平台**：React Native 可以复用逻辑写移动端

适用场景：数据驱动的复杂交互应用，比如我们这种 BI 看板——大量状态管理、频繁的数据更新、复杂的组件组合。

---

#### Q8: 为什么选择 Ant Design Pro 脚手架？优缺点？

**A:**

Ant Design Pro 是蚂蚁金服开源的企业级前端解决方案，基于 Umi Max + Ant Design。

**选它的原因**：项目是企业级 BI 平台，Pro 内置了权限管理、Layout、国际化、OpenAPI 代码生成这些企业级能力，不需要从零搭建。

**优点**：
- 开箱即用：登录页、权限控制、菜单布局直接可用，省了大量基础搭建时间
- OpenAPI 插件：根据后端 Swagger 文档一键生成 TypeScript service 层代码
- 约定式配置：`initialState` 全局状态、`access.ts` 权限控制等约定文件降低了团队协作成本

**缺点**：
- 黑箱较多：Umi 框架封装程度高，出了问题不容易调试（比如 MFSU 对某些 CJS 库不兼容，我就踩过 react-grid-layout 的坑）
- 模板代码多：初始项目有大量用不到的 Post、Swagger 等模板代码需要清理
- 版本更新快：Umi 3 到 4 有不少 breaking change，社区教程版本不统一

---

#### Q9: 你如何保证项目规范？每个技术什么作用？

**A:**

四个工具形成完整链路：

| 工具 | 作用 | 介入时机 |
|------|------|---------|
| **TypeScript** | 静态类型检查，编译期发现类型错误 | 编码时（IDE 实时提示） |
| **ESLint** | 代码质量检查，发现潜在 bug（未使用变量、隐式类型转换等） | 编码时 + 保存时 |
| **Prettier** | 代码格式统一（缩进、引号、分号等），消除风格争论 | 保存时自动格式化 |
| **Husky** | Git hooks 管理，在 `git commit` 前自动跑 lint-staged | 提交时 |

工作流：写代码 → ESLint + Prettier 实时检查 → `git commit` 触发 Husky → Husky 调用 lint-staged → lint-staged 只对暂存区文件跑 ESLint + Prettier → 通过才允许提交。

这样即使有人忘了配置 IDE 插件，也无法提交不规范的代码。

---

#### Q10: 什么是 Umi OpenAPI 插件？使用流程？

**A:**

**是什么**：Umi Max 内置的代码生成插件，根据后端 Swagger/OpenAPI 接口文档自动生成 TypeScript 的请求函数和类型定义。

**使用流程**：

1. 后端启动后暴露 Swagger 文档地址（如 `http://localhost:8080/api/v2/api-docs`）
2. 在 `config.ts` 中配置：

```ts
openAPI: [{
  requestLibPath: "import { request } from '@umijs/max'",
  schemaPath: "http://localhost:8080/api/v2/api-docs",
  projectName: 'yubi',
  mock: false,
}]
```

3. 运行 `npm run openapi`，自动在 `src/services/yubi/` 下生成：
   - `chartController.ts`：所有图表相关的请求函数（如 `genChartByAiUsingPost`）
   - `userController.ts`：用户相关请求函数
   - `typings.d.ts`：所有 DTO 的 TypeScript 类型定义

4. 业务代码中直接 `import { genChartByAiUsingPost } from '@/services/yubi/chartController'` 调用，有完整的类型提示。

**价值**：后端改了接口，重跑一次 openapi 就能发现前端哪些调用需要同步修改，大幅减少联调问题。

---

#### Q11: ECharts 为什么兼容性好？怎么接收后端动态 JSON 渲染图表的？

**A:**

**为什么兼容性好**：
- 底层用 Canvas/SVG 渲染而非 DOM，不受浏览器 CSS 差异影响
- 版本迭代稳定，v5 API 向下兼容
- 支持 PC、移动端、小程序多端
- Apache 开源项目，社区维护活跃

**渲染流程**：

1. 先读 ECharts 官方文档和 Playground，明确需要的 JSON 结构（title、xAxis、yAxis、series 等）
2. 后端 AI prompt 中约束输出格式为 ECharts V5 option JSON
3. 后端将 AI 返回的 JSON 字符串存入数据库 `genChart` 字段
4. 前端拿到字符串后经过 Schema 校验层处理（容错解析 → Zod 校验 → sanitize 过滤 → fallback 兜底）
5. 校验通过的 option 对象传给 `echarts-for-react` 的 `<ReactECharts option={option} />` 渲染

`echarts-for-react` 本质是 ECharts 的 React 封装，它在 `useEffect` 中调用 `echarts.setOption(option)` 完成渲染，option 变化时自动更新。

---

#### Q12: React Hooks 最常用的几个？

**A:** 结合我项目中的实际使用：

1. **useState**：管理组件状态，如图表列表、loading 状态、搜索参数
2. **useEffect**：处理副作用，如数据请求、事件监听、定时器。我的轮询 Hook 里有 3 个 useEffect 分别管理轮询启停、可见性监听和组件卸载清理
3. **useRef**：存储不触发渲染的值，如定时器 ID、轮询间隔、上次数据快照
4. **useMemo**：缓存计算结果，如 Schema 校验结果、`hasPendingCharts` 判断
5. **useCallback**：缓存函数引用，如看板里的事件处理函数，避免子组件无效重渲染

#### Q12.1 追问: useCallback 不用会怎样？

**A:** 两个问题：
1. **子组件无效重渲染**：如果用了 `React.memo` 的子组件接收了一个没被 `useCallback` 包裹的函数 prop，每次父组件渲染都创建新函数引用，memo 失效
2. **useEffect 无限循环**：如果一个函数作为 useEffect 的依赖，每次渲染创建新函数 → effect 重新执行 → 可能 setState → 又触发渲染

在看板的 `handleRemoveItem` 中我用了 `useCallback` + `setState` 函数式更新 `prev => prev.filter(...)`，这样不需要依赖 `items`，函数引用更稳定。

---

#### Q13: 怎么处理跨域请求？

**A:**

**什么是跨域**：浏览器同源策略限制，前端 `localhost:8000` 请求后端 `localhost:8080` 时协议/域名/端口不完全一致就是跨域。

**本项目方案**：Umi 的 dev proxy。在 `config/proxy.ts` 中配置：

```ts
dev: {
  '/api/': {
    target: 'http://localhost:8080',
    changeOrigin: true,
  },
}
```

开发环境下前端请求 `/api/chart/gen` 会被 webpack-dev-server 的 proxy 转发到 `http://localhost:8080/api/chart/gen`，浏览器只看到同源请求，不触发跨域。

**其他常用方案**：
- **CORS**：后端设置 `Access-Control-Allow-Origin` 响应头，生产环境最常用
- **Nginx 反向代理**：生产环境部署时用 Nginx 将前后端统一为同一域名
- **JSONP**：只支持 GET，基本淘汰了

---

#### Q14: 如何优化渲染性能？

**A:** 结合项目中实际用到的和通用方案：

| 优化手段 | 本项目实践 |
|---------|-----------|
| **虚拟列表** | react-window 渲染万行数据，DOM 节点从 N 降到 ~15 |
| **代码分割** | Umi 路由级 chunk，首屏只加载当前页面 JS |
| **骨架屏** | 消除 CLS，感知加载更快 |
| **useMemo/useCallback** | 缓存 Schema 解析结果和事件函数，减少无效计算和渲染 |
| **条件轮询** | 没有 pending 图表时完全停止轮询，减少不必要的 setState |

通用方案还包括：图片懒加载、CDN 加速、浏览器缓存策略（Cache-Control）、减少重排重绘（批量 DOM 操作、will-change 提示）、SSR/SSG 混合渲染等。

---

## 四、AI 相关

---

#### Q15: 本项目是怎么接入 AI 功能的？

**A:**

**架构**：后端通过 HTTP 调用百度千帆的 OpenAI 兼容 API，不依赖任何 SDK。

**具体实现**：
1. `application.yml` 中配置 `base-url`、`api-key`、`model` 三个参数
2. `AiManager.java` 用 Hutool HTTP 工具库发 POST 请求到 `{base-url}/chat/completions`，请求体格式完全遵循 OpenAI 标准（messages 数组 + model + temperature）
3. AI 返回的内容用 `【【【【` 分隔符分割为 ECharts JSON 和文字结论
4. 前端拿到 JSON 后经过 Schema 校验层处理再渲染

**为什么不用 SDK**：原来用的"鱼聪明"平台已下线，改用 HTTP 直连后**可以随时切换 AI 提供商**——只需改 yml 中的三行配置（base-url、api-key、model），代码完全不用动。这就是面向接口编程的好处。

#### Q15.1 追问: 你怎么处理 AI 生成内容的不确定性？

**A:** 本质上是**不信任 AI 输出**，把 AI 当作一个不稳定的外部 API：

**前端做了 4 层防御**：
1. JSON 容错提取（AI 多输出了文字也能解析）
2. Zod Schema 结构校验（缺字段、类型错都能捕获）
3. sanitize 过滤（移除 function 字符串防 XSS）
4. Error Boundary（ECharts 渲染崩溃也能兜住）

**后端做了约束**：
- system prompt 明确要求输出格式和分隔符
- temperature 设为 0.2（降低随机性）
- 异步模式下失败的图表会记录 `execMessage`，用户可以看到具体失败原因

#### Q15.2 追问: 如果要提升 AI 生成准确率，你会怎么做？

**A:**
1. **前端约束**：限定图表类型选择范围，传给后端后在 prompt 中明确约束
2. **Few-shot 示例**：在 system prompt 里加 1-2 个标准输出示例，让 AI 模仿格式
3. **校验重试**：Schema 校验失败时把错误信息拼到 prompt 让 AI 重新生成
4. **用户编辑**：集成 Monaco Editor 让用户修正 AI 的小错误

最有效的是 **prompt 工程优化 + Few-shot**，成本最低效果最好。

---

## 五、代码规范 & 工程化

---

#### Q16: 什么是 Husky？它有什么作用？

**A:**

Husky 是一个 Git hooks 管理工具，能在 `git commit`、`git push` 等操作时自动执行脚本。

在我的项目中，Husky 在 `pre-commit` 阶段触发 `lint-staged`，只对本次 `git add` 的文件运行 ESLint 和 Prettier。**作用是把代码质量检查的最后一道门卡在提交环节**——即使开发者本地没配 IDE 插件，也不可能提交不规范的代码。

配合 TypeScript 编译检查（`tsc --noEmit`），形成三道防线：
1. IDE 实时提示（开发时）
2. ESLint + Prettier（保存时）
3. Husky + lint-staged（提交时）

---

#### Q17: 类组件和函数组件的区别？

**A:**

| | 类组件 | 函数组件 |
|---|---|---|
| 写法 | `class Foo extends React.Component` | `const Foo = () => {}` |
| 状态 | `this.state` + `this.setState` | `useState` |
| 生命周期 | `componentDidMount` 等方法 | `useEffect` 模拟 |
| 独有能力 | Error Boundary（`getDerivedStateFromError`） | 自定义 Hook |
| 性能 | 需要创建实例，稍重 | 更轻量 |

现在主流用函数组件 + Hooks，但 Error Boundary 是唯一仍需要 Class 组件的场景。我项目里的 `ChartErrorBoundary` 就是 Class 组件。

---

## 六、分页列表（基础功能）

---

#### Q18: 分页列表的实现思路？

**A:**

核心是**状态驱动**：用 `searchParams`（包含 current、pageSize、搜索关键词等）驱动数据加载。

1. `useState` 管理 `searchParams`、`chartList`、`total`、`loading` 四个状态
2. `useEffect` 监听 `searchParams` 变化，自动调用 `loadData()` 请求后端分页接口
3. 分页组件 `onChange` 回调更新 `searchParams`，触发 useEffect → 重新请求
4. 搜索时重置 `current` 为 1，避免在第 N 页搜索后还停留在第 N 页

**关键点**：`setSearchParams({ ...searchParams, current: page })` 必须创建新对象——React 用 `Object.is` 判断状态是否变化，直接修改原对象引用不变，React 不会重渲染。

---

## 七、需要实测的数据验证清单

以下量化数据写在简历上可能被追问"怎么测的"，建议实际跑一遍记录截图：

| 数据点 | 怎么测 | 工具 |
|--------|--------|------|
| 虚拟列表万行渲染 <50ms | 打开"查看原始数据"（3000+ 行），Performance 面板录制 Rendering 耗时 | Chrome DevTools → Performance |
| DOM 节点数 ~15 | 虚拟列表滚动时，Elements 面板数 VirtualList 子节点数 | Chrome DevTools → Elements |
| 轮询减少 60% 请求 | 创建一个异步图表，Network 面板记录 30s 内请求数；对比固定 3s 间隔（约 10 次）vs 指数退避（约 4 次） | Chrome DevTools → Network |
| 首屏 JS 减少 40% | `npm run build` 看 dist 产物大小；或 Webpack Bundle Analyzer 对比 | `npm run analyze` 或 `npm run build` |
| CLS → 0 | Lighthouse Performance 跑分，查看 CLS 指标 | Chrome DevTools → Lighthouse |
| 图表崩溃率 0 | 手动测试：传空串、非法 JSON、缺 series、含 function 的 option 给 SafeChart | 手动验证 + 截图 |
| 内存占用降低 90% | Performance Monitor 对比有/无虚拟列表时的 JS Heap Size | Chrome DevTools → Performance Monitor |

**建议**：跑完后截图保存，面试时如果被问到可以说"我用 Chrome DevTools 的 Performance 面板录制过，具体数据是 ×××"。
