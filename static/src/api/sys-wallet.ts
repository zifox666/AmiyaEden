import request from '@/utils/http'

// ─── 用户端钱包 ───

/** 获取我的钱包 */
export function fetchMyWallet() {
  return request.post<Api.SysWallet.Wallet>({
    url: '/api/v1/operation/wallet/my'
  })
}

/** 获取我的钱包流水 */
export function fetchMyWalletTransactions(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.SysWallet.WalletTransaction>>({
    url: '/api/v1/operation/wallet/my/transactions',
    data: data ?? { current: 1, size: 20 }
  })
}

// ─── 管理员钱包管理 ───

/** 管理员查询所有用户钱包 */
export function adminListWallets(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.SysWallet.Wallet>>({
    url: '/api/v1/system/wallet/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员查看指定用户钱包 */
export function adminGetWallet(userId: number) {
  return request.post<Api.SysWallet.Wallet>({
    url: '/api/v1/system/wallet/detail',
    data: { user_id: userId }
  })
}

/** 管理员调整用户钱包余额 */
export function adminAdjustWallet(data: Api.SysWallet.AdjustParams) {
  return request.post<Api.SysWallet.Wallet>({
    url: '/api/v1/system/wallet/adjust',
    data
  })
}

/** 管理员查询钱包流水 */
export function adminListTransactions(data?: Api.SysWallet.TransactionSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.SysWallet.WalletTransaction>>({
    url: '/api/v1/system/wallet/transactions',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员查询操作日志 */
export function adminListWalletLogs(data?: Api.SysWallet.LogSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.SysWallet.WalletLog>>({
    url: '/api/v1/system/wallet/logs',
    data: data ?? { current: 1, size: 20 }
  })
}
