# EVE 舰队管理

## 概述

该模块负责管理EVE舰队的创建/邀请/记录参与过的成员等功能

## 前置需求

为每一个user_id创建一个system_wallet负责发放/兑换奖励等
创建PAP记录表/日志

## 功能

1. 创建舰队

    - admin/fc 权限组可用

    - ```sql
        create table main.fleets
        (
            id                text
                primary key,
            title             text,
            start_at          datetime,
            end_at            datetime,
            importance        text,  # strat op/cta/other
            pap_count         real,
            fc_user_id   text,
            fc_character_id text,
            created_at        datetime,
            updated_at        datetime
        );
      ```

      ```text
        标题 *
        行动的名称
        等级 *
        strat op | CTA | other
        PAP 数量 *
        0
        起止时间 *
        2026/02/24 19:18
        ～
        2026/02/24 20:18
        舰队指挥*
        (选择一个登记的角色作为舰队指挥)
        详细信息
        关于此行动的额外信息。这个字段支持

      ```

2. 已创建舰队操作（admin/创建该条目的fc）

    1. 编辑舰队信息
    2. 发送集结通知

        - 需要admin可视化设置webhook才可用

    3. 删除舰队
    4. 查看舰队组成
    5. 发放pap到成员

        - 需要admin可视化设置pap * wallet 基数才可用 否则只记录pap发放记录
        - 允许多次发放，但是只记录最后一次更新情况，不要重复发放（用于补充后加入舰队的成员pap/奖励

    6. 创建舰队邀请链接

        - 提供给普通用户选择名下指定角色加入舰队使用

3. PAP 功能（admin/创建该条目的fc）

    1. 调用pkg/eve/esi获取acess_token(请注册scope[esi-fleets.read_fleet.v1, esi-fleets.write_fleet.v1esi-fleets.write_fleet.v1] )
    2. Get character fleet info -> Get fleet members -> 记录并发放PAP
    3. Update fleet."motd" += "\n- {pap_count} 已发放 {time} -"
    4. ESI操作模型请参照[esi_openai.json](./esi_openai.json)

4. 舰队邀请链接功能（admin/创建该条目的fc）

    1. 普通用户点击邀请链接 -> 选择名下角色 -> 加入舰队
    2. Create fleet invitation -> 默认 `squad_member`
