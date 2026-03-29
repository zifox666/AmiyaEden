---
status: draft
doc_type: draft
owner: engineering
last_reviewed: 2026-03-29
source_of_truth:
  - static/src/views/welfare/approval/index.vue
  - static/src/views/welfare/my/index.vue
  - docs/standards/frontend-table-pages.md
---

# 福利模块表格页面规范修复计划

## 问题描述

福利模块多个页面存在表格规范违反，包括滚动失效、ElEmpty 冗余、缺少 ledger 配置等问题。

## 根因分析

### 结构问题：ElTabs 嵌套在 art-table-card 内破坏 Flex 高度链

当前结构：

```
.art-full-height
  └─ ElCard.art-table-card
       └─ .el-card__body (height:100%; overflow:hidden)
            └─ ElTabs                    ← 无 flex 属性，断链
                 └─ .el-tabs__content    ← 无 flex 属性，断链
                      └─ .el-tab-pane    ← 无 flex 属性，断链
                           └─ ArtTable   ← 无法获取约束高度，自然撑高
```

`art-table-card` 的 `.el-card__body` 设置了 `height: 100%; overflow: hidden`，但 ElTabs 的内部 DOM 元素（`el-tabs__content`、`el-tab-pane`）缺少 `flex: 1; min-height: 0`，导致高度链断裂。ArtTable 的 `useTableHeight` hook 无法计算出正确的容器高度，表格随内容自然撑高而非内部滚动。

### 规范违反项

| 项 | 当前 | 规范要求 |
|----|------|----------|
| 待发放 tab 页大小 | `size: 50` | 审批记录为无限增长，应使用 ledger 规则 `size: 200` |
| 待发放 tab variant | 无 | 应加 `visual-variant="ledger"` |
| ElEmpty 组件 | 手动渲染 | ArtTable 内置空状态，冗余可移除 |
| ElTabs 使用方式 | TabPane 包裹完整表格内容 | 应使用 SRP manage 模式（Tab 仅作标签，表格共享） |

## 全项目表格审计结果

### 需要修复的页面

| # | 页面文件 | 问题类型 | 严重度 |
|---|----------|----------|--------|
| 1 | `welfare/approval/index.vue` | ElTabs 断链、ElEmpty 冗余、pending 缺 ledger | 高 |
| 2 | `welfare/my/index.vue` | ElTabs 断链、ElEmpty 冗余、applications 缺 ledger + size=10 | 高 |

### 已合规无需修复的页面

| 页面文件 | 说明 |
|----------|------|
| `system/wallet/index.vue` | CSS-only 修复已到位，子模块均 ledger + size:200 |
| `shop/browse/index.vue` | CSS-only 修复已到位 |
| `shop/order-manage/index.vue` | CSS-only 修复已到位 |
| `srp/manage/index.vue` | 金标准参考，ElTabs 共享模式 |
| `newbro/manage/index.vue` | ElTabs 在页面级（不在 art-table-card 内），三个表均 ledger + size:200 |
| `newbro/select-captain/index.vue` | ElTabs 在普通 ElCard 内，ArtTable 有 ledger + size:200 |
| `info/wallet/index.vue` | 无 tabs，标准表格页 |
| `welfare/settings/index.vue` | 无 tabs，标准布局；管理型表格数量有限，size:50 合理 |
| 其余 20+ 使用 ArtTable 的页面 | 无 ElTabs 嵌套，标准 art-table-card 布局 |

---

## 修复方案

两个福利页面均采用 SRP manage 页面的已验证模式（`srp/manage/index.vue`）。

---

### 页面 1：welfare/approval/index.vue

#### 当前问题

| 项 | 当前 | 规范要求 |
|----|------|----------|
| 待发放 tab 页大小 | `size: 50` | 审批记录为无限增长，应使用 ledger 规则 `size: 200` |
| 待发放 tab variant | 无 | 应加 `visual-variant="ledger"` |
| ElEmpty 组件 | 手动渲染（两个 tab） | ArtTable 内置空状态，冗余可移除 |
| ElTabs 使用方式 | TabPane 包裹完整表格内容 | 应使用 SRP manage 模式（Tab 仅作标签，表格共享） |
| Flex 高度链 | 断链（无 scoped CSS 修复） | 需补全 flex 链 |

#### 目标结构（参照 SRP manage）

```
.art-full-height
  └─ ElCard.art-table-card.welfare-approval-card
       └─ .el-card__body (flex 链修复)
            └─ .welfare-approval-content (flex 容器)
                 ├─ ElTabs (仅标签)
                 │    ├─ ElTabPane label="待发放" name="pending"
                 │    └─ ElTabPane label="历史记录" name="history"
                 ├─ ArtTableHeader (共享)
                 └─ .welfare-approval-table-shell (flex:1; min-height:0)
                      └─ ArtTable (共享，响应式 props)
```

#### Script 变更

1. **合并为单个 useTable 实例**：根据 `activeTab` 切换 `apiParams`、`columnsFactory`
2. **移除 `ElEmpty` 导入**：ArtTable 内置空状态处理
3. **两个 tab 均使用 ledger 规则**：`size: 200`，`visual-variant="ledger"`

#### Scoped CSS 新增

```scss
.welfare-approval-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.welfare-approval-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.welfare-approval-content {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.welfare-approval-table-shell {
  flex: 1;
  min-height: 0;
}
```

---

### 页面 2：welfare/my/index.vue

#### 当前问题

| 项 | 当前 | 规范要求 |
|----|------|----------|
| ElTabs 使用方式 | TabPane 包裹完整表格内容（与 approval 相同断链） | SRP manage 模式 |
| ElEmpty 组件 | 手动渲染（两个 tab） | 冗余，ArtTable 内置空状态 |
| 已领取 tab 页大小 | `size: 10` | 领取记录为无限增长，应 `size: 200` |
| 已领取 tab variant | 无 | 应加 `visual-variant="ledger"` |
| Flex 高度链 | 断链（无 scoped CSS 修复） | 需补全 flex 链 |

#### 特殊考虑

此页面两个 tab 的数据结构完全不同：
- **申请福利 tab**：`eligibleRows`（本地计算，无分页，无需 ledger）—— 可视为配置型表格
- **已领取福利 tab**：`applications`（API 分页，无限增长记录）—— 需要 ledger 规则

因此不适合直接合并为单个 ArtTable（数据来源不同），应采用 **CSS-only 修复模式**（参照 `system/wallet/index.vue`），保持两个独立 ArtTable，仅修复 flex 高度链。

#### 目标结构

```
.art-full-height
  └─ ElCard.art-table-card
       └─ .el-card__body (flex 链修复)
            └─ ElTabs (flex:1; display:flex; flex-direction:column)
                 └─ .el-tabs__content (flex:1; overflow:hidden)
                      └─ .el-tab-pane (height:100%; display:flex; flex-direction:column)
                           └─ ArtTable (各 tab 独立)
```

#### Script 变更

1. **已领取 tab 补充 ledger**：`apiParams: { current: 1, size: 200 }`，添加 `visual-variant="ledger"`
2. **移除两个 `ElEmpty`**：ArtTable 内置空状态处理
3. **移除 `ElEmpty` 导入**

#### Scoped CSS 新增

```scss
.welfare-my-page {
  :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  :deep(.el-tabs) {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  :deep(.el-tabs__content) {
    flex: 1;
    overflow: hidden;
    min-height: 0;
  }
  :deep(.el-tab-pane) {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
}
```

---

## 实施步骤

### Step 1：修复 welfare/approval/index.vue（SRP 共享模式重构）

1. **重构模板**：ElTabs 改为仅含标签，ArtTable 移到 Tabs 外部，包裹在 flex 容器中
2. **重构 Script**：合并两个 useTable 为一个，使用 computed 响应式切换参数
3. **新增 scoped CSS**：添加 Flex 高度链修复样式
4. **清理**：移除 `ElEmpty` 导入和冗余空状态渲染
5. **验证**：本地 `make dev` 确认两个 tab 切换正常，表格可滚动

### Step 2：修复 welfare/my/index.vue（CSS-only 修复）

1. **新增 scoped CSS**：添加 Flex 高度链修复样式
2. **修复已领取 tab**：`size` 改为 200，添加 `visual-variant="ledger"`
3. **清理**：移除 `ElEmpty` 导入和冗余空状态渲染
4. **验证**：本地 `make dev` 确认两个 tab 切换正常，表格可滚动

## 参考文件

- 正确模式参考：`static/src/views/srp/manage/index.vue`（ElTabs + ArtTable 共享模式）
- 简洁参考：`static/src/views/info/wallet/index.vue`（无 tabs 的标准表格页）
- 表格规范：`docs/standards/frontend-table-pages.md`
- 高度计算 hook：`static/src/hooks/core/useTableHeight.ts`
- 全局布局 CSS：`static/src/assets/styles/core/app.scss`

## 约束

- 不改动 `useTableHeight`、`useLayoutHeight`、`ArtTable` 等公共组件/hook
- 不引入新的依赖或工具函数
- 保持现有 i18n key 不变
- 保持现有 API 调用不变
