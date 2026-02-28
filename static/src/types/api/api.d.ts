/**
 * API 接口类型定义模块
 *
 * 提供所有后端接口的类型定义
 *
 * ## 主要功能
 *
 * - 通用类型（分页参数、响应结构等）
 * - 认证类型（登录、用户信息等）
 * - 系统管理类型（用户、角色等）
 * - 全局命名空间声明
 *
 * ## 使用场景
 *
 * - API 请求参数类型约束
 * - API 响应数据类型定义
 * - 接口文档类型同步
 *
 * ## 注意事项
 *
 * - 在 .vue 文件使用需要在 eslint.config.mjs 中配置 globals: { Api: 'readonly' }
 * - 使用全局命名空间，无需导入即可使用
 *
 * ## 使用方式
 *
 * ```typescript
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const response: Api.Auth.UserInfo = await fetchUserInfo()
 * ```
 *
 * @module types/api/api
 * @author Art Design Pro Team
 */

declare namespace Api {
  /** 通用类型 */
  namespace Common {
    /** 分页参数 */
    interface PaginationParams {
      /** 当前页码 */
      current: number
      /** 每页条数 */
      size: number
      /** 总条数 */
      total: number
    }

    /** 通用搜索参数 */
    type CommonSearchParams = Pick<PaginationParams, 'current' | 'size'>

    /** 分页响应基础结构 */
    interface PaginatedResponse<T = any> {
      records: T[]
      current: number
      size: number
      total: number
    }

    /** 启用状态 */
    type EnableStatus = '1' | '2'
  }

  /** 认证类型 */
  namespace Auth {
    /** 登录参数（已废弃，仅保留兼容） */
    interface LoginParams {
      userName: string
      password: string
    }

    /** 登录响应（已废弃，仅保留兼容） */
    interface LoginResponse {
      token: string
      refreshToken: string
    }

    /** EVE 角色信息 */
    interface EveCharacter {
      id: number
      character_id: number
      character_name: string
      portrait_url: string
      user_id: number
      scopes: string
      token_expiry: string
      token_valid: boolean
      corporation_id: number
      alliance_id: number
    }

    /** 已注册的 ESI Scope */
    interface RegisteredScope {
      module: string
      scope: string
      description: string
      required: boolean
    }

    /** /me 接口响应 */
    interface MeResponse {
      user: {
        id: number
        nickname: string
        avatar: string
        status: number
        role: string
        primary_character_id: number
        last_login_at: string
        last_login_ip: string
      }
      characters: EveCharacter[]
      /** 用户所有活跃角色编码列表 */
      roles: string[]
      /** 用户所有权限标识列表 */
      permissions: string[]
    }

    /** 用户信息（路由守卫和权限指令使用） */
    interface UserInfo {
      buttons: string[]
      roles: string[]
      userId: number
      userName: string
      avatar: string
      characters?: EveCharacter[]
      primaryCharacterId?: number
    }
  }

  /** 系统管理类型 */
  namespace SystemManage {
    /** 用户列表 */
    type UserList = Api.Common.PaginatedResponse<UserListItem>

    /** 用户列表项（匹配后端 model.User） */
    interface UserListItem {
      id: number
      nickname: string
      avatar: string
      status: number // 1:正常 0:禁用
      role: string // 历史兼容字段
      last_login_at: string | null
      last_login_ip: string
      created_at: string
      updated_at: string
    }

    /** 用户搜索参数 */
    type UserSearchParams = Partial<{
      nickname: string
      status: number
    }> &
      Partial<Api.Common.CommonSearchParams>

    /** 角色列表（分页） */
    type RoleList = Api.Common.PaginatedResponse<RoleItem>

    /** 角色（匹配后端 model.Role） */
    interface RoleItem {
      id: number
      code: string
      name: string
      description: string
      is_system: boolean
      sort: number
      status: number
      menu_ids?: number[]
      created_at: string
      updated_at: string
    }

    /** 创建角色请求 */
    interface CreateRoleParams {
      code: string
      name: string
      description?: string
      sort?: number
    }

    /** 更新角色请求 */
    interface UpdateRoleParams {
      name?: string
      description?: string
      sort?: number
    }

    /** 角色搜索参数 */
    type RoleSearchParams = Partial<Api.Common.CommonSearchParams>

    /** 菜单项（后端 model.Menu） */
    interface MenuItem {
      id: number
      parent_id: number
      type: 'dir' | 'menu' | 'button'
      name: string
      path: string
      component: string
      permission: string
      title: string
      icon: string
      sort: number
      is_hide: boolean
      keep_alive: boolean
      is_hide_tab: boolean
      fixed_tab: boolean
      status: number
      children?: MenuItem[]
      created_at: string
      updated_at: string
    }

    /** 创建菜单请求 */
    interface CreateMenuParams {
      parent_id?: number
      type: 'dir' | 'menu' | 'button'
      name: string
      path?: string
      component?: string
      permission?: string
      title: string
      icon?: string
      sort?: number
      is_hide?: boolean
      keep_alive?: boolean
      is_hide_tab?: boolean
      fixed_tab?: boolean
    }

    /** 更新菜单请求 */
    interface UpdateMenuParams {
      parent_id?: number
      type?: 'dir' | 'menu' | 'button'
      name?: string
      path?: string
      component?: string
      permission?: string
      title?: string
      icon?: string
      sort?: number
      is_hide?: boolean
      keep_alive?: boolean
      is_hide_tab?: boolean
      fixed_tab?: boolean
    }

    /** 用户角色关联 */
    interface UserRoleInfo {
      role_ids: number[]
      roles: RoleItem[]
    }
  }

  /** ESI 刷新队列类型 */
  namespace ESIRefresh {
    /** 任务定义信息 */
    interface TaskInfo {
      name: string
      description: string
      priority: number
      active_interval: string
      inactive_interval: string
      required_scopes: string[]
    }

    /** 任务运行时状态 */
    interface TaskStatus {
      task_name: string
      description: string
      character_id: number
      priority: number
      last_run?: string | null
      next_run?: string | null
      status: 'pending' | 'running' | 'success' | 'failed'
      error?: string
    }

    /** 手动触发任务请求参数（指定角色） */
    interface RunTaskParams {
      task_name: string
      character_id: number
    }

    /** 手动触发任务请求参数（所有角色） */
    interface RunTaskByNameParams {
      task_name: string
    }

    /** 任务状态搜索参数（分页 + 筛选） */
    type TaskStatusSearchParams = Partial<{
      task_name: string
      status: string
    }> &
      Partial<Api.Common.CommonSearchParams>

    /** 任务状态分页响应 */
    type TaskStatusList = Api.Common.PaginatedResponse<TaskStatus>
  }

  /** 舰队管理类型 */
  namespace Fleet {
    /** 舰队信息 */
    interface FleetItem {
      id: string
      title: string
      description: string
      start_at: string
      end_at: string
      importance: 'strat_op' | 'cta' | 'other'
      pap_count: number
      fc_user_id: number
      fc_character_id: number
      fc_character_name: string
      esi_fleet_id: number | null
      created_at: string
      updated_at: string
    }

    /** 舰队列表（分页） */
    type FleetList = Api.Common.PaginatedResponse<FleetItem>

    /** 舰队搜索参数 */
    type FleetSearchParams = Partial<{
      importance: string
      fc_user_id: number
    }> &
      Partial<Api.Common.CommonSearchParams>

    /** 创建舰队请求 */
    interface CreateFleetParams {
      title: string
      description?: string
      start_at: string
      end_at: string
      importance: 'strat_op' | 'cta' | 'other'
      pap_count: number
      character_id: number
    }

    /** 更新舰队请求 */
    interface UpdateFleetParams {
      title?: string
      description?: string
      start_at?: string
      end_at?: string
      importance?: string
      pap_count?: number
      character_id?: number
      esi_fleet_id?: number
    }

    /** 舰队成员 */
    interface FleetMember {
      id: number
      fleet_id: string
      character_id: number
      character_name: string
      user_id: number
      ship_type_id: number | null
      solar_system_id: number | null
      joined_at: string
      created_at: string
    }

    /** PAP 记录 */
    interface PapLog {
      id: number
      fleet_id: string
      character_id: number
      user_id: number
      pap_count: number
      issued_by: number
      created_at: string
    }

    /** 邀请链接 */
    interface FleetInvite {
      id: number
      fleet_id: string
      code: string
      active: boolean
      expires_at: string
      created_at: string
    }

    /** 加入舰队请求 */
    interface JoinFleetParams {
      code: string
      character_id: number
    }

    /** 钱包信息 */
    interface Wallet {
      id: number
      user_id: number
      balance: number
      updated_at: string
    }

    /** 钱包流水 */
    interface WalletTransaction {
      id: number
      user_id: number
      amount: number
      reason: string
      ref_type: string
      ref_id: string
      balance_after: number
      operator_id: number
      created_at: string
    }

    /** 钱包流水分页 */
    type WalletTransactionList = Api.Common.PaginatedResponse<WalletTransaction>

    /** ESI 角色舰队信息 */
    interface CharacterFleetInfo {
      fleet_id: number
      fleet_boss_id: number
      role: string
      squad_id: number
      wing_id: number
    }

    /** ESI 舰队成员 */
    interface ESIFleetMember {
      character_id: number
      join_time: string
      role: string
      role_name: string
      ship_type_id: number
      solar_system_id: number
      squad_id: number
      wing_id: number
    }
  }

  /** 系统钱包类型（独立于 EVE Wallet） */
  namespace SysWallet {
    /** 钱包信息 */
    interface Wallet {
      id: number
      user_id: number
      balance: number
      updated_at: string
    }

    /** 钱包流水 */
    interface WalletTransaction {
      id: number
      user_id: number
      amount: number
      reason: string
      ref_type: string
      ref_id: string
      balance_after: number
      operator_id: number
      created_at: string
    }

    /** 钱包操作日志 */
    interface WalletLog {
      id: number
      operator_id: number
      target_uid: number
      action: 'add' | 'deduct' | 'set'
      amount: number
      before: number
      after: number
      reason: string
      created_at: string
    }

    /** 管理员调整余额请求 */
    interface AdjustParams {
      target_uid: number
      action: 'add' | 'deduct' | 'set'
      amount: number
      reason: string
    }

    /** 流水查询参数 */
    type TransactionSearchParams = Partial<{
      current: number
      size: number
      user_id: number
      ref_type: string
    }>

    /** 操作日志查询参数 */
    type LogSearchParams = Partial<{
      current: number
      size: number
      operator_id: number
      target_uid: number
      action: string
    }>
  }

  /** SRP 补损管理类型 */
  namespace Srp {
    /** 舰船标准补损金额 */
    interface ShipPrice {
      id: number
      ship_type_id: number
      ship_name: string
      amount: number
      created_by: number
      updated_by: number
      created_at: string
      updated_at: string
    }

    /** 新增/更新舰船价格请求 */
    interface UpsertShipPriceParams {
      id?: number
      ship_type_id: number
      ship_name: string
      amount: number
    }

    /** 补损申请 */
    interface Application {
      id: number
      user_id: number
      character_id: number
      character_name: string
      killmail_id: number
      fleet_id: string | null
      note: string
      ship_type_id: number
      ship_name: string
      solar_system_id: number
      solar_system_name: string
      killmail_time: string
      corporation_id: number
      corporation_name: string
      alliance_id: number
      alliance_name: string
      recommended_amount: number
      final_amount: number
      review_status: 'pending' | 'approved' | 'rejected'
      reviewed_by: number | null
      reviewed_at: string | null
      review_note: string
      payout_status: 'pending' | 'paid'
      paid_by: number | null
      paid_at: string | null
      created_at: string
      updated_at: string
    }

    /** 申请列表分页响应 */
    type ApplicationList = Api.Common.PaginatedResponse<Application>

    /** 提交补损申请请求 */
    interface SubmitApplicationParams {
      character_id: number
      killmail_id: number
      fleet_id?: string | null
      note?: string
      final_amount?: number
    }

    /** 申请搜索参数（管理端） */
    type ApplicationSearchParams = Partial<{
      fleet_id: string
      character_id: number
      review_status: string
      payout_status: string
    }> &
      Partial<Api.Common.CommonSearchParams>

    /** 审批请求 */
    interface ReviewParams {
      action: 'approve' | 'reject'
      review_note?: string
      final_amount?: number
    }

    /** 发放请求 */
    interface PayoutParams {
      final_amount?: number
    }

    /** 快捷 KM 列表条目 */
    interface FleetKillmailItem {
      killmail_id: number
      killmail_time: string
      ship_type_id: number
      solar_system_id: number
      victim_name: string
    }
  }

  /** 商店系统类型 */
  namespace Shop {
    /** 商品 */
    interface Product {
      id: number
      name: string
      description: string
      image: string
      price: number
      stock: number
      max_per_user: number
      type: 'normal' | 'redeem'
      need_approval: boolean
      status: number
      sort_order: number
      created_at: string
      updated_at: string
    }

    /** 订单 */
    interface Order {
      id: number
      order_no: string
      user_id: number
      product_id: number
      product_name: string
      product_type: string
      quantity: number
      unit_price: number
      total_price: number
      status: string
      transaction_id: number | null
      remark: string
      reviewed_by: number | null
      reviewed_at: string | null
      review_remark: string
      created_at: string
      updated_at: string
    }

    /** 兑换码 */
    interface RedeemCode {
      id: number
      order_id: number
      product_id: number
      user_id: number
      code: string
      status: 'unused' | 'used' | 'expired'
      used_at: string | null
      expires_at: string | null
      created_at: string
      updated_at: string
    }

    /** 购买请求 */
    interface BuyParams {
      product_id: number
      quantity: number
      remark?: string
    }

    /** 商品创建请求 */
    interface ProductCreateParams {
      name: string
      description?: string
      image?: string
      price: number
      stock?: number
      max_per_user?: number
      type: 'normal' | 'redeem'
      need_approval?: boolean
      status?: number
      sort_order?: number
    }

    /** 商品更新请求 */
    interface ProductUpdateParams {
      id: number
      name?: string
      description?: string
      image?: string
      price?: number
      stock?: number
      max_per_user?: number
      type?: string
      need_approval?: boolean
      status?: number
      sort_order?: number
    }

    /** 商品查询参数 */
    type ProductSearchParams = Partial<{
      current: number
      size: number
      status: number
      type: string
      name: string
    }>

    /** 订单查询参数 */
    type OrderSearchParams = Partial<{
      current: number
      size: number
      user_id: number
      product_id: number
      status: string
    }>

    /** 订单审批请求 */
    interface OrderReviewParams {
      order_id: number
      remark?: string
    }

    /** 兑换码查询参数 */
    type RedeemSearchParams = Partial<{
      current: number
      size: number
      product_id: number
      status: string
    }>
  }

  /** 通知相关类型 */
  namespace Notification {
    /** 通知项 */
    interface NotificationItem {
      id: number
      character_id: number
      notification_id: number
      sender_id: number
      sender_type: string
      text?: string
      timestamp: string
      type: string
      is_read?: boolean
    }

    /** 通知列表请求参数 */
    interface ListParams {
      page?: number
      page_size?: number
      type?: string
      is_read?: boolean
    }

    /** 通知列表响应 */
    interface NotificationSummary {
      list: NotificationItem[]
      total: number
      page: number
      page_size: number
      unread_count: number
    }

    /** 未读数响应 */
    interface UnreadCountResponse {
      unread_count: number
    }

    /** 标记已读请求 */
    interface MarkAsReadParams {
      notification_ids: number[]
    }
  }

  /** 工作台类型 */
  namespace Dashboard {
    /** 卡片统计数据 */
    interface Cards {
      eve_wallet_balance: number
      eve_skill_points: number
      system_wallet_balance: number
      alliance_pap: number
    }

    /** 统一舰队参与记录 */
    interface FleetItem {
      source: 'internal' | 'alliance'
      id: string
      title: string
      start_at: string
      end_at?: string
      importance?: string
      pap_count: number
      ship_type_name?: string
      character_name?: string
    }

    /** 月度 PAP 统计项 */
    interface PapMonthly {
      year: number
      month: number
      total_pap: number
    }

    /** PAP 统计数据 */
    interface PapStats {
      alliance: PapMonthly[]
      internal: PapMonthly[]
    }

    /** 补损列表项 */
    interface SrpItem {
      id: number
      character_name: string
      ship_name: string
      solar_system_name: string
      killmail_time: string
      recommended_amount: number
      final_amount: number
      review_status: string
      payout_status: string
      created_at: string
    }

    /** 工作台完整响应 */
    interface DashboardResult {
      cards: Cards
      fleets: FleetItem[]
      pap_stats: PapStats
      srp_list: SrpItem[]
    }
  }
}
