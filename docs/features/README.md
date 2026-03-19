---
status: active
doc_type: index
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
---

# Feature Docs

## 说明

`docs/features/current/` 只描述当前仓库已经落地、可以从代码直接找到入口的模块行为。

如果一个想法还没有完整接入路由、页面、任务或服务，请写进 `docs/specs/draft/`，不要写进这里。

## 当前模块

- [auth-and-characters.md](current/auth-and-characters.md)
- [operation.md](current/operation.md)
- [info-and-reporting.md](current/info-and-reporting.md)
- [srp.md](current/srp.md)
- [commerce.md](current/commerce.md)
- [administration.md](current/administration.md)
- [esi-refresh.md](current/esi-refresh.md)
- [sde.md](current/sde.md)

## Feature Doc 最少要回答的问题

- 这个模块当前对用户提供什么能力
- 入口页面和后端路由在哪里
- 需要什么角色 / 权限
- 哪些行为是必须保持的
- 真实代码文件在哪里
