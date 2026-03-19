---
status: active
doc_type: api
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/pkg/response/response.go
  - static/src/types/api/api.d.ts
---

# API 约定

## Base URL

```text
/api/v1
```

## 认证方式

需要登录的接口通过以下任一方式携带 JWT：

- `Authorization: Bearer <token>`
- `?token=<token>`

## 统一响应

成功：

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

分页成功：

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "pageSize": 20
  }
}
```

说明：

- 401 / 403 使用对应 HTTP 状态码
- 许多业务失败仍会返回 HTTP 200，但 `code != 200`

## 合同同步顺序

改动接口时，按以下顺序同步：

1. 后端 handler / service / response shape
2. 前端 `static/src/api/*`
3. `static/src/types/api/api.d.ts`
4. UI 使用处
5. `docs/api/route-index.md`
6. 对应 feature doc

## 文档分工

- `docs/api/conventions.md`: 统一规则
- `docs/api/route-index.md`: 当前已注册路由面
- 更细的字段结构：以代码和类型文件为准

`docs/api/route-index.md` 中的权限边界应尽量按路由显式写出，不要只依赖章节上下文让读者猜测是否需要登录或额外角色。

## 禁止事项

- 前后端偷偷改字段名不更新类型
- 在多个 markdown 文件复制同一份路由表
- 把未来计划接口写进 route index 冒充已实现
