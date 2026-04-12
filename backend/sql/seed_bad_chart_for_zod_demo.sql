-- 用于验证：Zod Schema + JSON 容错 + sanitize + SafeChart / Error Boundary
-- 库名：与 application.yml 中 spring.datasource.url 一致（默认 yubi；Docker/其他环境若为 genbi 请改 USE）
-- 执行：mysql -u root -p < seed_bad_chart_for_zod_demo.sql
-- 重复执行可先清理：DELETE FROM chart WHERE name LIKE '[校验演示]%';

USE yubi;

-- 绑定到「已有第一个用户」；若没有用户请先注册一个账号
SET @uid = (SELECT id FROM user WHERE isDelete = 0 ORDER BY id LIMIT 1);

-- 1) Zod 校验失败：series 为空数组（违反 min(1)）→ 前端 parseChartOption 走失败分支 + fallbackOption
INSERT INTO chart (goal, name, chartData, chartType, genChart, genResult, status, userId, isDelete)
VALUES (
  '演示：Schema 要求至少一条 series',
  '[校验演示] Zod失败-emptySeries',
  '{}',
  'bar',
  '{"title":{"text":"故意坏配置"},"series":[]}',
  '期望：卡片内出现「图表数据异常」告警 + 占位柱状图，页面不白屏',
  'succeed',
  @uid,
  0
);

-- 2) JSON 容错：前缀噪声 + 合法 JSON（走 extractJsonObject → JSON.parse）→ 校验通过则正常渲染
INSERT INTO chart (goal, name, chartData, chartType, genChart, genResult, status, userId, isDelete)
VALUES (
  '演示：模拟 AI 在 JSON 前后夹废话',
  '[校验演示] JSON容错-带前缀',
  '{}',
  'bar',
  '以下是图表配置：{"series":[{"type":"bar","data":[5,3,8],"name":"系列A"}],"xAxis":{"type":"category"},"yAxis":{}} 配置结束',
  '期望：能解析出 bar 图（若 userId 匹配可在看板添加）',
  'succeed',
  @uid,
  0
);

-- 3) 非 JSON / 无法提取 {}：走「解析失败」分支 + 统一 fallback
INSERT INTO chart (goal, name, chartData, chartType, genChart, genResult, status, userId, isDelete)
VALUES (
  '演示：完全不是 JSON',
  '[校验演示] 非JSON文本',
  '{}',
  'bar',
  'this is not json at all',
  '期望：告警 + 占位图，不抛未捕获异常',
  'succeed',
  @uid,
  0
);

-- 4) 可选：series 类型非法（不在 z.enum 内）→ Zod 失败
INSERT INTO chart (goal, name, chartData, chartType, genChart, genResult, status, userId, isDelete)
VALUES (
  '演示：series.type 不在允许枚举中',
  '[校验演示] Zod失败-非法type',
  '{}',
  'bar',
  '{"series":[{"type":"not_a_real_type","data":[1,2]}],"xAxis":{},"yAxis":{}}',
  '期望：Schema 校验失败提示 + fallback',
  'succeed',
  @uid,
  0
);
