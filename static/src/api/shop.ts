import request from '@/utils/http'

// ─── 用户端商店 ───

/** 获取上架商品列表 */
export function fetchProducts(data?: Partial<Api.Common.CommonSearchParams & { type: string }>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.Product>>({
    url: '/api/v1/shop/products',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 获取商品详情 */
export function fetchProductDetail(productId: number) {
  return request.post<Api.Shop.Product>({
    url: '/api/v1/shop/product/detail',
    data: { product_id: productId }
  })
}

/** 购买商品 */
export function buyProduct(data: Api.Shop.BuyParams) {
  return request.post<Api.Shop.Order>({
    url: '/api/v1/shop/buy',
    data
  })
}

/** 获取我的订单 */
export function fetchMyOrders(data?: Partial<Api.Common.CommonSearchParams & { status: string }>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.Order>>({
    url: '/api/v1/shop/orders',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 获取我的兑换码 */
export function fetchMyRedeemCodes(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.RedeemCode>>({
    url: '/api/v1/shop/redeem/list',
    data: data ?? { current: 1, size: 20 }
  })
}

// ─── 管理员商店管理 ───

/** 管理员查询商品列表 */
export function adminListProducts(data?: Api.Shop.ProductSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.Product>>({
    url: '/api/v1/system/shop/product/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员创建商品 */
export function adminCreateProduct(data: Api.Shop.ProductCreateParams) {
  return request.post<Api.Shop.Product>({
    url: '/api/v1/system/shop/product/add',
    data
  })
}

/** 管理员更新商品 */
export function adminUpdateProduct(data: Api.Shop.ProductUpdateParams) {
  return request.post<Api.Shop.Product>({
    url: '/api/v1/system/shop/product/edit',
    data
  })
}

/** 管理员删除商品 */
export function adminDeleteProduct(id: number) {
  return request.post({
    url: '/api/v1/system/shop/product/delete',
    data: { id }
  })
}

/** 管理员查询订单列表 */
export function adminListOrders(data?: Api.Shop.OrderSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.Order>>({
    url: '/api/v1/system/shop/order/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员审批通过订单 */
export function adminApproveOrder(data: Api.Shop.OrderReviewParams) {
  return request.post<Api.Shop.Order>({
    url: '/api/v1/system/shop/order/approve',
    data
  })
}

/** 管理员拒绝订单 */
export function adminRejectOrder(data: Api.Shop.OrderReviewParams) {
  return request.post<Api.Shop.Order>({
    url: '/api/v1/system/shop/order/reject',
    data
  })
}

/** 管理员查询兑换码 */
export function adminListRedeemCodes(data?: Api.Shop.RedeemSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.RedeemCode>>({
    url: '/api/v1/system/shop/redeem/list',
    data: data ?? { current: 1, size: 20 }
  })
}

// ─── 抽奖（用户端）───

/** 获取进行中的抽奖活动列表 */
export function fetchLotteryActivities(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.LotteryActivity>>({
    url: '/api/v1/shop/lottery/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 执行抽奖 */
export function drawLottery(activityId: number) {
  return request.post<Api.Shop.DrawResult>({
    url: '/api/v1/shop/lottery/draw',
    data: { activity_id: activityId }
  })
}

/** 获取我的抽奖记录 */
export function fetchMyLotteryRecords(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.LotteryRecord>>({
    url: '/api/v1/shop/lottery/records',
    data: data ?? { current: 1, size: 20 }
  })
}

// ─── 抽奖管理（管理员）───

/** 管理员查询抽奖活动列表 */
export function adminListLotteryActivities(data?: Partial<Api.Common.CommonSearchParams>) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.LotteryActivity>>({
    url: '/api/v1/system/shop/lottery/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员创建抽奖活动 */
export function adminCreateLotteryActivity(data: Api.Shop.LotteryActivityCreateParams) {
  return request.post<Api.Shop.LotteryActivity>({
    url: '/api/v1/system/shop/lottery/add',
    data
  })
}

/** 管理员更新抽奖活动 */
export function adminUpdateLotteryActivity(data: Api.Shop.LotteryActivityUpdateParams) {
  return request.post<Api.Shop.LotteryActivity>({
    url: '/api/v1/system/shop/lottery/edit',
    data
  })
}

/** 管理员删除抽奖活动 */
export function adminDeleteLotteryActivity(id: number) {
  return request.post({
    url: '/api/v1/system/shop/lottery/delete',
    data: { id }
  })
}

/** 管理员新增奖品 */
export function adminCreateLotteryPrize(data: Api.Shop.LotteryPrizeCreateParams) {
  return request.post<Api.Shop.LotteryPrize>({
    url: '/api/v1/system/shop/lottery/prize/add',
    data
  })
}

/** 管理员更新奖品 */
export function adminUpdateLotteryPrize(data: Api.Shop.LotteryPrizeUpdateParams) {
  return request.post<Api.Shop.LotteryPrize>({
    url: '/api/v1/system/shop/lottery/prize/edit',
    data
  })
}

/** 管理员删除奖品 */
export function adminDeleteLotteryPrize(id: number) {
  return request.post({
    url: '/api/v1/system/shop/lottery/prize/delete',
    data: { id }
  })
}

/** 管理员查询抽奖记录 */
export function adminListLotteryRecords(data?: {
  current?: number
  size?: number
  activity_id?: number
}) {
  return request.post<Api.Common.PaginatedResponse<Api.Shop.LotteryRecord>>({
    url: '/api/v1/system/shop/lottery/records',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员更新抽奖记录发放状态 */
export function adminUpdateLotteryRecordDelivery(
  id: number,
  deliveryStatus: 'pending' | 'delivered'
) {
  return request.post({
    url: '/api/v1/system/shop/lottery/records/deliver',
    data: { id, delivery_status: deliveryStatus }
  })
}

// ─── 图片上传 ───

/** 上传商店图片，返回访问 URL */
export function uploadShopImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<{ url: string }>({
    url: '/api/v1/upload/image',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
