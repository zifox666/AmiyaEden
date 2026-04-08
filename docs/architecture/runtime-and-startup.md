---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-04-09
source_of_truth:
  - server/main.go
  - server/bootstrap
---

# 运行与启动

本文档描述后端启动顺序与运行时行为。依赖要求与本地启动流程见 `docs/guides/local-development.md`。

## 后端启动顺序

`server/main.go` 当前启动流程：

1. 初始化配置
2. 初始化日志
3. 初始化 JWT
4. 初始化数据库
5. 初始化 Redis
6. 初始化 cron
7. 异步检查 SDE
8. 注册 ESI scopes
9. 初始化 HTTP 路由
10. 启动服务

SDE 检查更新当前行为：

- 通过 `sde.download_url` 获取最新 release 信息
- 若本地配置了 `sde.proxy`，会优先尝试通过代理下载
- 若代理连接失败，会自动回退为直连重试
- 导入成功后在 `sde_versions` 中记录当前版本

## 数据库初始化副作用

数据库初始化不仅建立连接，还会执行：

- `AutoMigrate`
- 自定义索引补齐
- schema 规范化与兼容处理

## 运行时提示

- 新人物 SSO 成功后，后台会触发 ESI 全量刷新与自动权限同步
- ESI 刷新队列按 cron 调度，不要求手工逐个任务启动
- SDE 缺失会直接影响舰队配置 EFT 解析、名称翻译、搜索等共享能力
- `register` 页面源码仍在仓库中，但不是当前支持的登录架构；`forget-password` 页面已移除
