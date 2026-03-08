# SQL 基础知识笔记

> 以本项目 `create_table.sql` 为例，讲解 SQL 建表语法和常用知识。

---

## 一、建表语句逐行解析

### 1. 创建数据库

```sql
create database if not exists yubi;
use yubi;
```

- `create database` — 创建数据库
- `if not exists` — 如果已存在就不重复创建（防止报错）
- `use yubi` — 切换到 yubi 数据库，之后的操作都在这个库里执行

### 2. 用户表完整解析

```sql
create table if not exists user
(
    id           bigint auto_increment comment 'id' primary key,
    userAccount  varchar(256)                           not null comment '账号',
    userPassword varchar(512)                           not null comment '密码',
    userName     varchar(256)                           null comment '用户昵称',
    userAvatar   varchar(1024)                          null comment '用户头像',
    userRole     varchar(256) default 'user'            not null comment '用户角色：user/admin',
    createTime   datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime   datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete     tinyint      default 0                 not null comment '是否删除',
    index idx_userAccount (userAccount)
) comment '用户' collate = utf8mb4_unicode_ci;
```

拆解每一列的格式：

```
列名  数据类型  [约束条件...]  comment '注释'
```

逐列解释：

| 列 | 含义 |
|---|---|
| `id bigint auto_increment primary key` | 主键，bigint 类型，自增（每插入一行自动+1） |
| `userAccount varchar(256) not null` | 账号，最长 256 字符的字符串，不允许为空 |
| `userPassword varchar(512) not null` | 密码（存的是加密后的哈希值，所以比较长） |
| `userName varchar(256) null` | 昵称，允许为空（null 可以省略，因为默认就是允许空） |
| `userRole varchar(256) default 'user' not null` | 角色，默认值是 'user' |
| `createTime datetime default CURRENT_TIMESTAMP not null` | 创建时间，默认为当前时间 |
| `updateTime datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP` | 更新时间，每次修改记录时自动更新为当前时间 |
| `isDelete tinyint default 0 not null` | 逻辑删除标志，0=未删除，1=已删除 |
| `index idx_userAccount (userAccount)` | 为 userAccount 列创建索引（加速查询） |

末尾部分：

| 语法 | 含义 |
|---|---|
| `comment '用户'` | 表的注释/说明 |
| `collate = utf8mb4_unicode_ci` | 字符集排序规则，utf8mb4 支持 emoji，unicode_ci 表示不区分大小写 |

---

## 二、建表语法格式总结

```sql
create table [if not exists] 表名
(
    列名1  数据类型  [约束条件]  [comment '注释'],
    列名2  数据类型  [约束条件]  [comment '注释'],
    ...
    [index 索引名 (列名)]
) [comment '表注释'] [collate = 字符集];
```

---

## 三、常用数据类型

### 数值类型

| 类型 | 大小 | 范围 | 用途 |
|------|------|------|------|
| `tinyint` | 1 字节 | -128 ~ 127 | 小数字，如状态值、布尔值 (0/1) |
| `int` | 4 字节 | 约 ±21 亿 | 常规整数 |
| `bigint` | 8 字节 | 非常大 | 主键、用户ID 等 |
| `float` | 4 字节 | — | 单精度小数 |
| `double` | 8 字节 | — | 双精度小数 |
| `decimal(M,D)` | 可变 | 精确小数 | 金额（如 `decimal(10,2)` 表示最多 10 位，其中 2 位小数） |

### 字符串类型

| 类型 | 最大长度 | 用途 |
|------|----------|------|
| `char(N)` | 固定 N 个字符 | 固定长度的字符串（如性别 M/F） |
| `varchar(N)` | 最多 N 个字符 | **最常用**，可变长度字符串 |
| `text` | ~65,535 字符 | 大段文本（如文章内容、图表数据） |
| `longtext` | ~4GB | 超大文本 |

> `varchar` vs `text`：短字符串用 `varchar`（可以建索引），长文本用 `text`。
> 本项目中 `goal`、`chartData`、`genChart` 这些可能很长的字段用了 `text`。

### 日期时间类型

| 类型 | 格式 | 用途 |
|------|------|------|
| `date` | 2026-03-05 | 只存日期 |
| `time` | 14:30:00 | 只存时间 |
| `datetime` | 2026-03-05 14:30:00 | **最常用**，日期+时间 |
| `timestamp` | 同 datetime | 自动时区转换，范围到 2038 年 |

---

## 四、常用约束条件

| 约束 | 含义 | 示例 |
|------|------|------|
| `primary key` | 主键，唯一标识每一行 | `id bigint primary key` |
| `auto_increment` | 自增，每次插入自动 +1 | 通常和主键一起用 |
| `not null` | 不允许为空 | `userAccount varchar(256) not null` |
| `null` | 允许为空（默认） | `userName varchar(256) null` |
| `default 值` | 默认值 | `userRole varchar(256) default 'user'` |
| `unique` | 唯一约束，不能重复 | `email varchar(256) unique` |
| `comment '文字'` | 注释说明 | `comment '用户昵称'` |
| `on update CURRENT_TIMESTAMP` | 更新时自动刷新时间 | 用于 updateTime 字段 |

---

## 五、索引 (Index)

```sql
index idx_userAccount (userAccount)
```

- **索引**就像书的目录，让数据库按某列快速查找，不用全表扫描
- `idx_userAccount` 是索引名（命名惯例：`idx_列名`）
- 适合给**经常用来查询**的列加索引（如用户名、手机号）
- 主键自动有索引，不用额外加

### 索引的代价

- 加速读取（SELECT），但会略微减慢写入（INSERT/UPDATE）
- 占用额外存储空间
- 原则：查询多的列加索引，很少查的列不加

---

## 六、逻辑删除 vs 物理删除

本项目用了**逻辑删除**模式：

```sql
isDelete tinyint default 0 not null comment '是否删除'
```

| 方式 | 做法 | 特点 |
|------|------|------|
| 物理删除 | `DELETE FROM user WHERE id = 1` | 数据真正从表里消失，不可恢复 |
| 逻辑删除 | `UPDATE user SET isDelete = 1 WHERE id = 1` | 数据还在表里，只是标记为已删除 |

逻辑删除的好处：数据可恢复、可审计。查询时加条件 `WHERE isDelete = 0` 只查未删除的记录。
在 Spring Boot 中 MyBatis-Plus 配置了 `logic-delete-field: isDelete`，框架自动处理。

---

## 七、CRUD 基础操作

### 插入数据 (Create)

```sql
INSERT INTO user (userAccount, userPassword, userName)
VALUES ('zhangsan', 'abc123hash', '张三');
```

### 查询数据 (Read)

```sql
-- 查询所有用户
SELECT * FROM user;

-- 条件查询
SELECT id, userName FROM user WHERE userRole = 'admin';

-- 模糊查询
SELECT * FROM user WHERE userName LIKE '%张%';

-- 排序
SELECT * FROM user ORDER BY createTime DESC;

-- 分页（第 1 页，每页 10 条）
SELECT * FROM user LIMIT 10 OFFSET 0;
-- 第 2 页
SELECT * FROM user LIMIT 10 OFFSET 10;

-- 统计
SELECT COUNT(*) FROM user WHERE isDelete = 0;
```

### 更新数据 (Update)

```sql
UPDATE user SET userName = '李四' WHERE id = 1;
```

### 删除数据 (Delete)

```sql
-- 物理删除
DELETE FROM user WHERE id = 1;

-- 逻辑删除（本项目的做法）
UPDATE user SET isDelete = 1 WHERE id = 1;
```

---

## 八、多表查询（JOIN）

假设要查"某用户创建的所有图表"：

```sql
SELECT u.userName, c.goal, c.chartType
FROM user u
JOIN chart c ON u.id = c.userId
WHERE u.id = 1;
```

- `JOIN ... ON` — 把两张表按条件连接
- `u` 和 `c` 是表的别名，写起来更短
- chart 表的 `userId` 关联 user 表的 `id`，这就是**外键关系**

### JOIN 类型

| 类型 | 含义 |
|------|------|
| `JOIN` (INNER JOIN) | 只返回两表都匹配的行 |
| `LEFT JOIN` | 返回左表所有行，右表没匹配的显示 NULL |
| `RIGHT JOIN` | 返回右表所有行，左表没匹配的显示 NULL |

---

## 九、常用函数

| 函数 | 用途 | 示例 |
|------|------|------|
| `COUNT(*)` | 计数 | `SELECT COUNT(*) FROM user` |
| `SUM(列)` | 求和 | `SELECT SUM(amount) FROM orders` |
| `AVG(列)` | 平均值 | `SELECT AVG(score) FROM exam` |
| `MAX(列)` / `MIN(列)` | 最大/最小值 | `SELECT MAX(createTime) FROM user` |
| `NOW()` | 当前时间 | `SELECT NOW()` |
| `CONCAT(a, b)` | 拼接字符串 | `SELECT CONCAT(userName, '-', userRole) FROM user` |
| `GROUP_CONCAT()` | 分组拼接 | `SELECT userId, GROUP_CONCAT(goal) FROM chart GROUP BY userId` |

---

## 十、GROUP BY 和 HAVING

```sql
-- 统计每个用户创建了多少图表
SELECT userId, COUNT(*) AS chartCount
FROM chart
WHERE isDelete = 0
GROUP BY userId
HAVING chartCount > 5;
```

- `GROUP BY` — 按某列分组
- `HAVING` — 对分组后的结果筛选（类似 WHERE，但用于聚合后）
- `AS chartCount` — 给计算结果起别名

---

## 十一、SQL 执行顺序

写的顺序和实际执行顺序不同：

```
FROM → WHERE → GROUP BY → HAVING → SELECT → ORDER BY → LIMIT
```

所以 `WHERE` 不能用 `SELECT` 里定义的别名，但 `HAVING` 可以。

---

## 十二、本项目数据表关系

```
user (用户表)
  ├── id (主键)
  ├── userAccount
  ├── userPassword
  └── ...

chart (图表表)
  ├── id (主键)
  ├── userId ──→ 关联 user.id
  ├── goal
  ├── chartData
  └── ...
```

一个用户可以有多个图表 → **一对多关系**（user 1:N chart）
