---
status: active
doc_type: standard
owner: frontend
last_reviewed: 2026-03-20
source_of_truth:
  - static/src/hooks/core/useTable
  - static/src/components/core
---

# 前端表格页面标准

## 适用范围

适用于后台管理类、带分页的列表类、标准 CRUD 表格页面。

## 默认模式

新增标准表格页时，优先遵循：

- 搜索区在卡片外
- `ElCard.art-table-card` 承载表格
- `ArtTableHeader` 承载刷新、列设置、主操作按钮
- `ArtTable` 承载列表与分页
- `useTable` 管理 `loading / data / pagination / searchParams`
- 对话框放在 `ElCard` 外，作为同级节点

## 必须遵守

- 需要分页的标准管理页，默认使用 `useTable`
- 视图层不要直接 `axios` / `fetch`
- 列标题、按钮、空态、校验提示必须走 i18n
- 权限控制优先通过路由、`v-auth`、store / hooks 处理
- 重复搜索区、编辑弹窗、列定义应抽到 `modules/`

## 推荐结构

```text
views/<module>/<page>/
├── index.vue
└── modules/
    ├── <page>-search.vue
    ├── <page>-dialog.vue
    └── columns.ts
```

## 允许例外

以下场景可以直接使用原生 `ElTable`，但应在页面注释或文档中说明原因：

- 详情页内的只读子表格
- 多块数据混排的分析页或 dashboard
- `ArtTable` 难以表达的树表 / 高度定制展开行
- 第三方数据导入页、临时预览页

即使使用 `ElTable`，也仍需遵守：

- API 调用放在 `static/src/api`
- 用户可见文本本地化
- 权限不写成页面局部硬编码

## 提交前检查

- 是否真的需要分页
- 是否复用了 `useTable`
- 是否把对话框与搜索区拆出
- 是否所有用户可见字符串都已本地化
- 是否没有在页面里直接创建 HTTP client
