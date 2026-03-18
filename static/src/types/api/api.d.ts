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
      list: T[]
      total: number
      page: number
      pageSize: number
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
      token_invalid: boolean
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

    /** ESI 军团角色 → 系统角色映射 */
    interface EsiRoleMapping {
      id: number
      esi_role: string
      role_id: number
      role_code: string
      role_name: string
      created_at: string
    }

    /** 创建 ESI 角色映射请求 */
    interface CreateEsiRoleMappingParams {
      esi_role: string
      role_id: number
    }

    /** ESI 头衔 → 系统角色映射 */
    interface EsiTitleMapping {
      id: number
      corporation_id: number
      title_id: number
      title_name: string
      role_id: number
      role_code: string
      role_name: string
      created_at: string
    }

    /** 军团头衔信息（从头衔快照获取，用于前端下拉选择） */
    interface CorpTitleInfo {
      corporation_id: number
      title_id: number
      title_name: string
    }

    /** 创建 ESI 头衔映射请求 */
    interface CreateEsiTitleMappingParams {
      corporation_id: number
      title_id: number
      title_name?: string
      role_id: number
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
      fleet_config_id: number | null
      auto_srp_mode: 'disabled' | 'submit_only' | 'auto_approve'
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
      fleet_config_id?: number | null
      send_ping?: boolean
      auto_srp_mode?: 'disabled' | 'submit_only' | 'auto_approve'
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
      fleet_config_id?: number | null
      auto_srp_mode?: 'disabled' | 'submit_only' | 'auto_approve'
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

    /** 舰队成员（含 PAP 信息）*/
    interface MemberWithPap extends FleetMember {
      pap_count: number | null
      issued_at: string | null
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
      /** 以下为富化字段（联表查询返回） */
      character_name: string
      fleet_title: string
      fleet_start_at: string
      fc_character_name: string
      fleet_importance: string
      ship_type_id: number | null
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

  /** 舰队配置类型 */
  namespace FleetConfig {
    /** 舰队配置装配条目（不含 EFT，通过专用端点获取） */
    interface FittingItem {
      id: number
      fleet_config_id: number
      ship_type_id: number
      fitting_name: string
      srp_amount: number
    }

    /** 舰队配置 */
    interface FleetConfigItem {
      id: number
      name: string
      description: string
      created_by: number
      created_at: string
      updated_at: string
      fittings: FittingItem[]
    }

    /** 舰队配置列表（分页） */
    type FleetConfigList = Api.Common.PaginatedResponse<FleetConfigItem>

    /** 创建/更新装配条目请求（输入英文 EFT，后端解析） */
    interface FittingReq {
      fitting_name: string
      eft: string
      srp_amount: number
    }

    /** 创建舰队配置请求 */
    interface CreateFleetConfigParams {
      name: string
      description?: string
      fittings: FittingReq[]
    }

    /** 更新舰队配置请求 */
    interface UpdateFleetConfigParams {
      name?: string
      description?: string
      fittings?: FittingReq[]
    }

    /** 从用户装配导入请求 */
    interface ImportFittingParams {
      character_id: number
      fitting_id: number
    }

    /** 从用户装配导入响应（英文 EFT） */
    interface ImportFittingResponse {
      fitting_name: string
      eft: string
      srp_amount: number
    }

    /** 导出到 ESI 请求 */
    interface ExportToESIParams {
      character_id: number
      fleet_config_id: number
      fitting_item_id: number
    }

    /** 单个装配的本地化 EFT */
    interface EFTFittingItem {
      id: number
      eft: string
    }

    /** GetFittingEFT 响应 */
    interface EFTResponse {
      fittings: EFTFittingItem[]
    }

    /** 装备替代品 */
    interface FittingItemReplacement {
      id: number
      type_id: number
      type_name: string
    }

    /** 装备物品详情 */
    interface FittingItemDetail {
      id: number
      type_id: number
      type_name: string
      quantity: number
      flag: string
      flag_group: string
      importance: 'required' | 'optional' | 'replaceable'
      penalty: 'none' | 'half'
      replacement_penalty: 'none' | 'half'
      replacements: FittingItemReplacement[]
    }

    /** 装配物品详情响应 */
    interface FittingItemsResponse {
      fitting_id: number
      fitting_name: string
      ship_type_id: number
      items: FittingItemDetail[]
    }

    /** 单个物品设置更新请求 */
    interface ItemSettingUpdate {
      id: number
      importance: 'required' | 'optional' | 'replaceable'
      penalty: 'none' | 'half'
      replacement_penalty: 'none' | 'half'
      replacements?: number[]
    }

    /** 批量更新装备设置请求 */
    interface UpdateItemsSettingsParams {
      items: ItemSettingUpdate[]
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
      character_name?: string
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
      character_name?: string
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
      target_character_name?: string
      operator_character_name?: string
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
      /** 关联舰队标题（后端填充） */
      fleet_title?: string
      /** 关联舰队 FC 角色名（后端填充） */
      fleet_fc_name?: string
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
      character_id: number
      victim_name: string
    }

    /** KM 装配详情请求 */
    interface KillmailDetailRequest {
      killmail_id: number
      language?: string
    }

    /** KM 槽位中的物品 */
    interface KillmailSlotItem {
      item_id: number
      item_name: string
      quantity: number
      dropped: boolean
    }

    /** KM 槽位分组 */
    interface KillmailSlotGroup {
      flag_id: number
      flag_name: string
      flag_text: string
      order_id: number
      items: KillmailSlotItem[]
    }

    /** KM 装配详情响应 */
    interface KillmailDetailResponse {
      killmail_id: number
      killmail_time: string
      ship_type_id: number
      ship_name: string
      solar_system_id: number
      system_name: string
      character_id: number
      character_name: string
      janice_amount: number | null
      slots: KillmailSlotGroup[]
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
      limit_period: 'forever' | 'daily' | 'weekly' | 'monthly'
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
      limit_period?: 'forever' | 'daily' | 'weekly' | 'monthly'
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
      limit_period?: 'forever' | 'daily' | 'weekly' | 'monthly'
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

    // ─── 抽奖活动 ───

    /** 抽奖奖品稀有度 */
    type LotteryPrizeTier = 'normal' | 'rare' | 'legendary'

    /** 抽奖奖品 */
    interface LotteryPrize {
      id: number
      activity_id: number
      name: string
      image: string
      tier: LotteryPrizeTier
      probability_weight: number
      total_stock: number
      drawn_count: number
      created_at: string
      updated_at: string
    }

    /** 抽奖活动 */
    interface LotteryActivity {
      id: number
      name: string
      description: string
      image: string
      cost_per_draw: number
      status: number
      start_at: string | null
      end_at: string | null
      sort_order: number
      prizes: LotteryPrize[]
      created_at: string
      updated_at: string
    }

    /** 抽奖记录 */
    interface LotteryRecord {
      id: number
      user_id: number
      activity_id: number
      activity_name: string
      prize_id: number
      prize_name: string
      prize_tier: LotteryPrizeTier
      prize_image: string
      cost: number
      delivery_status: 'pending' | 'delivered'
      created_at: string
      updated_at: string
    }

    /** 抽奖结果 */
    interface DrawResult {
      prize: LotteryPrize
    }

    /** 创建抽奖活动参数 */
    interface LotteryActivityCreateParams {
      name: string
      description?: string
      image?: string
      cost_per_draw?: number
      status?: number
      start_at?: string | null
      end_at?: string | null
      sort_order?: number
    }

    /** 更新抽奖活动参数 */
    interface LotteryActivityUpdateParams {
      id: number
      name?: string
      description?: string
      image?: string
      cost_per_draw?: number
      status?: number
      start_at?: string | null
      end_at?: string | null
      sort_order?: number
    }

    /** 创建奖品参数 */
    interface LotteryPrizeCreateParams {
      activity_id: number
      name: string
      image?: string
      tier?: LotteryPrizeTier
      probability_weight?: number
      total_stock?: number
    }

    /** 更新奖品参数 */
    interface LotteryPrizeUpdateParams {
      id: number
      name?: string
      image?: string
      tier?: LotteryPrizeTier
      probability_weight?: number
      total_stock?: number
    }
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

  /** EVE 角色信息类型 */
  namespace EveInfo {
    /** 钱包流水请求参数 */
    interface WalletRequest {
      character_id: number
      page: number
      page_size: number
    }

    /** 钱包流水条目 */
    interface WalletJournal {
      id: number
      amount: number
      balance: number
      date: string
      description: string
      first_party_id: number
      second_party_id: number
      ref_type: string
      reason: string
    }

    /** 钱包流水响应 */
    interface WalletResponse {
      balance: number
      journals: WalletJournal[]
      total: number
      page: number
      page_size: number
    }

    /** 技能请求参数 */
    interface SkillRequest {
      character_id: number
      language?: string
    }

    /** 技能条目 */
    interface SkillItem {
      skill_id: number
      skill_name: string
      group_id: number
      group_name: string
      active_level: number
      trained_level: number
      skillpoints_in_skill: number
      learned: boolean // 是否已注射：false = 未吸收技能书
    }

    /** 技能队列条目 */
    interface SkillQueueItem {
      queue_position: number
      skill_id: number
      skill_name: string
      finished_level: number
      level_start_sp: number
      level_end_sp: number
      training_start_sp: number
      start_date: number
      finish_date: number
    }

    /** 技能列表响应 */
    interface SkillResponse {
      total_sp: number
      unallocated_sp: number
      skill_count: number
      skills: SkillItem[]
      skill_queue: SkillQueueItem[]
    }

    /** 可用舰船请求参数 */
    interface ShipRequest {
      character_id: number
      language?: string
    }

    /** 舰船技能需求 */
    interface ShipSkillReq {
      skill_id: number
      skill_name: string
      required_level: number
      current_level: number
      met: boolean
      depth: number
    }

    /** 舰船条目 */
    interface ShipItem {
      type_id: number
      type_name: string
      group_id: number
      group_name: string
      market_group_id: number
      market_group_name: string
      race_id: number
      race_name: string
      can_fly: boolean
      skill_reqs: ShipSkillReq[]
    }

    /** 可用舰船响应 */
    interface ShipResponse {
      total_ships: number
      flyable_ships: number
      ships: ShipItem[]
    }

    /** 克隆体/植入体请求 */
    interface ImplantsRequest {
      character_id: number
      language?: string
    }

    /** 位置信息 */
    interface ImplantLocation {
      location_id: number
      location_type: string
      location_name: string
    }

    /** 植入体条目 */
    interface ImplantItem {
      implant_id: number
      implant_name: string
    }

    /** 跳跃克隆体信息 */
    interface JumpCloneInfo {
      jump_clone_id: number
      location: ImplantLocation
      implants: ImplantItem[]
    }

    /** 克隆体/植入体响应 */
    interface ImplantsResponse {
      home_location: ImplantLocation | null
      last_clone_jump_date: string | null
      last_station_change_date: string | null
      jump_fatigue_expire: string | null
      last_jump_date: string | null
      active_implants: ImplantItem[]
      jump_clones: JumpCloneInfo[]
    }

    /** 装配列表请求 */
    interface FittingsRequest {
      language?: string
    }

    /** 装配物品条目 */
    interface FittingItemResponse {
      type_id: number
      type_name: string
      quantity: number
      flag: string
    }

    /** 按槽位分组的装配物品 */
    interface FittingSlotGroup {
      flag_name: string
      flag_text: string
      order_id: number
      items: FittingItemResponse[]
    }

    /** 单个装配 */
    interface FittingResponse {
      fitting_id: number
      character_id: number
      name: string
      description: string
      ship_type_id: number
      ship_name: string
      group_id: number
      group_name: string
      race_id: number
      race_name: string
      slots: FittingSlotGroup[]
    }

    /** 装配列表响应 */
    interface FittingsListResponse {
      total: number
      fittings: FittingResponse[]
    }

    /** 保存装配请求 */
    interface SaveFittingRequest {
      character_id: number
      fitting_id?: number
      name: string
      description?: string
      ship_type_id: number
      items: {
        type_id: number
        quantity: number
        flag: string
      }[]
    }

    /** 资产查询请求 */
    interface AssetsRequest {
      language?: string
    }

    /** 资产物品节点 */
    interface AssetItemNode {
      item_id: number
      type_id: number
      type_name: string
      group_name: string
      category_id: number
      quantity: number
      location_flag: string
      is_singleton: boolean
      is_blueprint_copy?: boolean
      asset_name?: string
      character_id: number
      character_name: string
      children?: AssetItemNode[]
    }

    /** 资产位置节点 */
    interface AssetLocationNode {
      location_id: number
      location_type: string
      location_name: string
      items: AssetItemNode[]
    }

    /** 资产列表响应 */
    interface AssetsResponse {
      total_items: number
      locations: AssetLocationNode[]
    }

    /** 合同请求（含分页与过滤） */
    interface ContractsRequest {
      current: number
      size: number
      type?: string
      status?: string
      language?: string
    }

    /** 合同竞标条目 */
    interface ContractBidItem {
      amount: number
      bid_id: number
      bidder_id: number
      date_bid: string
    }

    /** 合同物品条目 */
    interface ContractItemDetail {
      type_id: number
      type_name: string
      group_name: string
      category_id: number
      quantity: number
      is_included: boolean
      is_singleton: boolean
    }

    /** 单条合同响应（列表行，不含物品/竞标） */
    interface ContractItem {
      character_id: number
      character_name: string
      contract_id: number
      acceptor_id: number
      assignee_id: number
      availability: string
      buyout?: number
      collateral?: number
      date_accepted?: string
      date_completed?: string
      date_expired: string
      date_issued: string
      days_to_complete?: number
      end_location_id?: number
      for_corporation: boolean
      issuer_corporation_id: number
      issuer_id: number
      price?: number
      reward?: number
      start_location_id?: number
      status: string
      title?: string
      type: string
      volume?: number
    }

    /** 合同列表响应（分页） */
    type ContractsResponse = Api.Common.PaginatedResponse<ContractItem>

    /** 合同详情请求 */
    interface ContractDetailRequest {
      character_id: number
      contract_id: number
      language?: string
    }

    /** 合同详情响应（物品 + 竞标） */
    interface ContractDetailResponse {
      items: ContractItemDetail[]
      bids: ContractBidItem[]
    }
  }

  /** SDE 数据查询类型 */
  namespace Sde {
    /** 模糊搜索请求 */
    interface FuzzySearchRequest {
      keyword: string
      language?: string
      category_ids?: number[]
      exclude_category_ids?: number[]
      limit?: number
      search_member?: boolean
    }

    /** 模糊搜索结果条目 */
    interface FuzzySearchItem {
      id: number
      name: string
      group_id: number
      group_name: string
      category: string // "type" | "character"
    }
  }

  /** NPC 刷怪报表类型 */
  namespace NpcKill {
    /** 个人刷怪报表请求 */
    interface NpcKillRequest {
      character_id: number
      start_date?: string
      end_date?: string
      language?: string
      page?: number
      page_size?: number
    }

    /** 个人刷怪报表请求（所有角色汇总） */
    interface NpcKillAllRequest {
      start_date?: string
      end_date?: string
      language?: string
      page?: number
      page_size?: number
    }

    /** 公司刷怪报表请求（管理员） */
    interface NpcKillCorpRequest {
      start_date?: string
      end_date?: string
      language?: string
      page?: number
      page_size?: number
    }

    /** 总览统计 */
    interface Summary {
      total_bounty: number
      total_ess: number
      total_tax: number
      actual_income: number
      total_records: number
      estimated_hours: number
    }

    /** 按 NPC 分类统计 */
    interface ByNpc {
      npc_id: number
      npc_name: string
      count: number
      amount: number
    }

    /** 按地点分类统计 */
    interface BySystem {
      solar_system_id: number
      solar_system_name: string
      count: number
      amount: number
    }

    /** 时间趋势 */
    interface Trend {
      date: string
      amount: number
      count: number
    }

    /** 刷怪流水条目 */
    interface JournalItem {
      id: number
      character_id: number
      character_name: string
      amount: number
      tax: number
      date: string
      ref_type: string
      solar_system_id: number
      solar_system_name: string
      reason: string
    }

    /** 个人刷怪报表响应 */
    interface NpcKillResponse {
      summary: Summary
      by_npc: ByNpc[]
      by_system: BySystem[]
      trend: Trend[]
      journals: JournalItem[]
      total: number
      page: number
      page_size: number
    }

    /** 公司成员刷怪统计 */
    interface CorpMemberSummary {
      character_id: number
      character_name: string
      total_bounty: number
      total_ess: number
      total_tax: number
      actual_income: number
      record_count: number
    }

    /** 公司刷怪报表响应 */
    interface NpcKillCorpResponse {
      summary: Summary
      members: CorpMemberSummary[]
      by_system: BySystem[]
      trend: Trend[]
    }
  }

  /** Webhook 配置 */
  namespace Webhook {
    interface Config {
      url: string
      enabled: boolean
      type: 'discord' | 'feishu' | 'dingtalk' | 'onebot' | string
      fleet_template: string
      ob_target_type: 'group' | 'private'
      ob_target_id: number
      ob_token: string
    }
  }
}
