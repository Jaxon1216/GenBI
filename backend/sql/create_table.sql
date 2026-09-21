-- ===== GenBI Go 后端建表脚本 =====
-- 字段命名采用「前端口径」（见 GenBI/frontend/src/services/yubi/typings.d.ts）：
--   chart 表用 name / execMessage（而非 Java 的 chartName / Message）。
-- 关闭 MyBatis-Plus 的驼峰下划线转换问题在此不存在：这里直接用列名与前端 JSON key 对齐。

-- 创建库
create database if not exists genbi character set utf8mb4 collate utf8mb4_unicode_ci;
use genbi;

-- 用户表
create table if not exists user
(
    id           bigint auto_increment comment 'id' primary key,
    userAccount  varchar(256)                           not null comment '账号',
    userPassword varchar(512)                           not null comment '密码（MD5+固定盐）',
    userName     varchar(256)                           null comment '用户昵称',
    userAvatar   varchar(1024)                          null comment '用户头像',
    userRole     varchar(256) default 'user'            not null comment '用户角色：user/admin/ban',
    createTime   datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime   datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete     tinyint      default 0                 not null comment '是否删除',
    index idx_userAccount (userAccount)
) comment '用户' collate = utf8mb4_unicode_ci;

-- 图表信息表
create table if not exists chart
(
    id          bigint auto_increment comment 'id' primary key,
    userId      bigint                             not null comment '创建用户的id',
    name        varchar(256)                       null comment '图表名称',
    goal        text                               null comment '分析目标',
    chartData   text                               null comment '图表数据（CSV）',
    chartType   varchar(256)                       null comment '图表类型',
    genChart    text                               null comment '生成的图表配置（ECharts JSON）',
    genResult   text                               null comment '生成的分析结论',
    status      varchar(256) default 'wait'        not null comment '任务状态：wait/running/succeed/failed',
    execMessage text                               null comment '执行信息',
    createTime  datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime  datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete    tinyint      default 0             not null comment '是否删除',
    index idx_userId (userId)
) comment '图表信息' collate = utf8mb4_unicode_ci;
