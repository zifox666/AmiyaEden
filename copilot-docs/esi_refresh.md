# EVE ESI 数据刷新队列

## 概述

EVE ESI 数据刷新队列模块，负责管理和调度 ESI 数据的刷新任务。支持可视化任务列表，任务优先级以及不活跃角色重试，其他模块添加指定任务/角色刷新数据。

## 功能

- 通过EVE SSO获得access_tokken
- 定时刷新ESI数据
- 可视化任务列表，显示待刷新角色和数据类型
- 任务优先级设置，优先刷新重要数据
- 管理页面可以添加需要的权限 根据权限scope添加需要刷新的数据
- 创建一个readme.md保证后续可以按照格式添加新的刷新任务

## 可能用到的信息

1. ESI的全部数据 [openai.json](https://esi.evetech.net/meta/openapi.json) [本地版本](./esi_openai.json)
2. 活跃角色判断

    > 使用 `https://esi.evetech.net/characters/{character_id}/online` 根据last_login 七天未登录表明不活跃 部分高频任务放缓刷新

    ```json
    {
        "last_login": "2019-08-24T14:15:22Z",
        "last_logout": "2019-08-24T14:15:22Z",
        "logins": 0,
        "online": true
    }
    ```

## 任务及刷新频率(未列出的任务不设计)

1. Character affiliation

    - 默认刷新间隔 2 Hours
    - 统一任务 可以每次拉取数据库中1000个角色id 角色IDS[<1000] 2hours

2. 角色资产刷新

    - Get character assets / Get character asset locations / Get character asset names
    - 默认刷新间隔 1 Day / 不活跃角色刷新间隔 7 Days

3. 角色通知消息

    - Get character notifications
    - 默认刷新间隔 1 Day / 不活跃角色刷新间隔 7 Days

4. 角色Title

    - Get character corporation titles
    - 默认刷新间隔 6 Hours / 不活跃角色刷新间隔 7 Days

5. 角色克隆相关

    - Get jump fatigue / Get clones / Get active implants
    - 默认刷新间隔 6 Hours / 不活跃角色刷新间隔 7 Days
    - 角色的克隆植入体和跳跃疲劳

6. 角色合同类

    - Get contracts / Get contract bids / Get contract items
    - 默认刷新间隔 1 Day / 不活跃角色刷新间隔 7 Days

7. 角色killmail

    - Get a character's recent kills and losses
    - 默认刷新间隔 20 Minutes / 不活跃角色刷新间隔 3 Days
    - 该接口获取到的是 killmail_id + killmail_hash 还需要调取 `https://esi.evetech.net/killmails/{killmail_id}/{killmail_hash}` 对km详情进行入库
