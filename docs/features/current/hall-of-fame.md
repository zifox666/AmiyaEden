---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-10
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/hall_of_fame.go
  - static/src/api/hall-of-fame.ts
  - static/src/views/hall-of-fame
---

# 伏羲名人堂 / Fuxi Hall of Heroes

## 当前能力

- 提供一个独立于用户/人物主数据的名人堂模块，管理员可独立维护英雄条目
- 圣殿页面向所有 `Login` 用户展示名人堂画布、背景图与全部可见卡片
- 管理页面向 `admin` 开放，支持新增、编辑、删除、隐藏、拖拽摆放、调整层级与修改卡片尺寸
- 管理页支持上传圣殿背景图，直接保存为 base64 data URL，不落盘
- 卡片样式支持 `gold`、`silver`、`bronze`、`custom` 四种主题；自定义主题允许单独配置背景色、文字色和边框色
- 画布尺寸可由管理员调整；卡片坐标按百分比持久化，便于不同尺寸画布保持相对布局
- 圣殿页面按当前画布尺寸自适应缩放显示，卡片支持悬浮发光效果

## 入口

### 前端页面

- `static/src/views/hall-of-fame/temple` — 圣殿页（所有已登录用户）
- `static/src/views/hall-of-fame/manage` — 管理页（管理员）

### 后端路由

用户端：

- `GET /api/v1/hall-of-fame/temple` — 获取圣殿配置与可见卡片

管理端：

- `GET /api/v1/system/hall-of-fame/config`
- `PUT /api/v1/system/hall-of-fame/config`
- `POST /api/v1/system/hall-of-fame/upload-background`
- `GET /api/v1/system/hall-of-fame/cards`
- `POST /api/v1/system/hall-of-fame/cards`
- `PUT /api/v1/system/hall-of-fame/cards/batch-layout`
- `PUT /api/v1/system/hall-of-fame/cards/:id`
- `DELETE /api/v1/system/hall-of-fame/cards/:id`

## 权限边界

- 左侧菜单根节点 `伏羲名人堂 / Fuxi Hall of Heroes` 对所有 `Login` 用户可见
- `圣殿 / Temple` 子页面要求 `login: true`
- `管理 / Manage` 子页面要求 `roles: ['super_admin', 'admin']`
- 后端管理接口统一挂在 `/system/hall-of-fame/*` 下，由 `admin` 路由组保护

## 数据模型

### `hall_of_fame_config`

- 单例配置表，保存：
  - `background_image`
  - `canvas_width`
  - `canvas_height`
- 若首次读取时不存在记录，服务层会自动创建默认值（1920×1080）

### `hall_of_fame_card`

- 每张卡片包含：
  - 基础文案：`name`、`title`、`description`
  - 图像：`avatar`
  - 布局：`pos_x`、`pos_y`、`width`、`height`、`z_index`
  - 视觉：`style_preset`、`custom_bg_color`、`custom_text_color`、`custom_border_color`、`font_size`
  - 显示控制：`visible`
- 采用软删除

## 关键不变量

- 管理页可以读取全部卡片；圣殿页只返回 `visible = true` 的卡片
- 画布配置始终存在；服务层负责 get-or-create 默认单例
- 背景上传最大 5MB，仅允许 `jpeg/png/webp`
- 卡片坐标在服务层与前端拖拽辅助逻辑中都会被约束在 `0–100` 范围
- 卡片布局保存走批量接口，避免拖拽时逐条请求

## 主要代码文件

- `server/internal/model/hall_of_fame.go`
- `server/internal/repository/hall_of_fame.go`
- `server/internal/service/hall_of_fame.go`
- `server/internal/handler/hall_of_fame.go`
- `server/internal/router/router.go`
- `static/src/api/hall-of-fame.ts`
- `static/src/router/modules/hall-of-fame.ts`
- `static/src/types/api/api.d.ts` (`Api.HallOfFame`)
- `static/src/views/hall-of-fame/`
