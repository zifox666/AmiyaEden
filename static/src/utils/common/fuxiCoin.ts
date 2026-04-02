const fuxiCoinAmountFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2
})

export function formatFuxiCoinWhole(value: number | null | undefined) {
  if (value == null) return '-'
  return Math.round(Number(value)).toLocaleString('en-US')
}

export function formatFuxiCoinAmount(value: number | null | undefined) {
  if (value == null) return '-'
  return fuxiCoinAmountFormatter.format(Number(value))
}
