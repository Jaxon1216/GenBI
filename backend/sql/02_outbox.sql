-- ===== GenBI 事务消息表（outbox）=====
-- 作用：GenAsync 在同一个 DB 事务里写 chart(wait) + outbox(pending)，
--       消除原先「入库成功但发 MQ 失败导致任务永远 wait」的双写不一致；
--       由 server 进程内的 relay goroutine 轮询 pending 记录补发到 RabbitMQ。
-- 投递语义为 at-least-once（可能重复发同一 chartId），依赖消费端的幂等消费兜底。
-- 列名沿用「前端口径」驼峰风格（aggregateId/retryCount/createTime/updateTime），
-- charset/collate 对齐 chart 表（utf8mb4 / utf8mb4_unicode_ci）。

use genbi;

create table if not exists outbox
(
    id          bigint auto_increment comment 'id' primary key,
    aggregateId bigint                                 not null comment '聚合根id（此处为 chartId）',
    payload     varchar(256)                           not null comment '待投递的消息体（此处为 chartId 字符串）',
    status      varchar(32)  default 'pending'         not null comment '投递状态：pending/sent',
    retryCount  int          default 0                 not null comment '投递重试次数',
    createTime  datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime  datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    index idx_status (status)
) comment '事务消息表（outbox）' collate = utf8mb4_unicode_ci;
