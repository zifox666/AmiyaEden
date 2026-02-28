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
