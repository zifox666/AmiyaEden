const ISK_KEYWORD = 'isk'
const ISK_PER_FUXI_COIN = 1_000_000

type OrderIskSource = {
  product_name?: string | null
  total_price?: number | null
}

export function resolveOrderIskTotal(order: OrderIskSource) {
  const productName = String(order.product_name ?? '').toLowerCase()
  if (!productName.includes(ISK_KEYWORD)) return null

  return Math.round(Number(order.total_price ?? 0) * ISK_PER_FUXI_COIN)
}
