# killmail表结构

```sql
create table eve_killmail_list
(
    id                bigint auto_increment
        primary key,
    kill_mail_id      int                              null comment '击杀ID',
    kill_mail_hash    varchar(200)                     null comment '击杀Hash',
    kill_mail_time    timestamp                        null comment '击杀时间',
    solar_system_id   int                              null comment '所在星系ID',
    ship_type_id      int                              null comment '舰船ID',
    character_id      int                              null comment '角色ID',
    corporation_id    int                              null comment '军团ID',
    alliance_id       int                              null comment '联盟ID',
    janice_amount     decimal(20, 2)                   null comment '估价(留空)',
    create_time       timestamp                        null comment '创建时间',
    constraint uq_kill_mail_id
        unique (kill_mail_id)
);
create table eve_killmail_item
(
    id           bigint auto_increment
        primary key,
    kill_mail_id bigint                           null comment 'KMID',
    item_id      int                              null comment '物品类型ID',
    item_num     bigint                           null comment '数量',
    drop_type    tinyint(1)                       null comment '类型，True=掉落 False=损毁',
    flag    int                              null comment 'flag',
    create_time  timestamp                        null comment '创建时间',

);

create table eve_character_killmail
(
    id           bigint unsigned auto_increment primary key,
    character_id bigint not null,
    killmail_id  bigint not null,
    srped       tinyint(1) default 0,
    victim       tinyint(1) default 0,
    create_time  timestamp                        null comment '创建时间',
    constraint idx_char_km
        unique (character_id, killmail_id)
);

```
