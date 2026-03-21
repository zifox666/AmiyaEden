---
status: active
doc_type: api
owner: engineering
last_reviewed: 2026-03-22
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

需要认证的接口通过以下任一方式携带 JWT：

- `Authorization: Bearer <token>`
- `?token=<token>`

说明：

- 持有有效 JWT 的请求方可能仍是 `guest`
- 只有显式标为 `Login` 的接口才要求“已认证且非 `guest`”， 如 `user` or `admin`

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

## 权限标注规则

- `Public`: 无需登录
- `JWT`: 任意持有有效 JWT 的已认证用户都可访问，包含 `guest`
- `Login`: 任意已认证且非 `guest` 的产品用户都可访问
- `RequireRole(...)`: 只有显式列出的角色边界可访问
- `RequirePermission(...)`: 只有显式列出的权限边界可访问

说明：

- 当真实边界是“只要 JWT 有效即可访问”时，统一写成 `JWT`
- 不要用 `RequireRole(..., user)` 作为“任意登录用户”的文档缩写
- 当真实含义是“所有非 guest 登录用户都能访问”时，统一写成 `Login`
- 当真实边界是具体角色白名单时，继续写 `RequireRole(...)`
- guest onboarding / self-service 路由如 `/me`、`/sso/eve/characters`、`/menu/list` 应明确标注为 `JWT`

## 禁止事项

- 前后端偷偷改字段名不更新类型
- 在多个 markdown 文件复制同一份路由表
- 把未来计划接口写进 route index 冒充已实现
