---
status: draft
doc_type: draft
owner: backend
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/service/srp.go
  - server/internal/service/auto_srp.go
---

# Killmail Schema 草案

## 说明

这份内容保留的是一个早期 killmail 表结构想法，不再代表当前数据库真实结构。

## 用途

保留这份草案是为了记录 killmail 相关设计关注点：

- killmail 基础信息
- killmail item 明细
- character 与 killmail 的关联
- SRP 是否已使用该 killmail

## 当前原则

- 真实实现以代码中的 model / repository / service 为准
- 如果需要重新设计 killmail 表结构，应基于现有模型重新整理，而不是直接照搬旧 SQL
