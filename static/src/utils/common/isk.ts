const MILLION_ISK = 1_000_000
const plainFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2
})

const SMART_UNITS = [
  { threshold: 1_000_000_000_000, suffix: 'T' },
  { threshold: 1_000_000_000, suffix: 'B' },
  { threshold: 1_000_000, suffix: 'M' },
  { threshold: 1_000, suffix: 'K' }
] as const

export function formatIskPlain(value: number | null | undefined) {
  if (value == null) return '-'
  return plainFormatter.format(Number(value))
}

export function formatIskSmart(value: number | null | undefined) {
  if (value == null) return '-'

  const numericValue = Number(value)
  const sign = numericValue < 0 ? '-' : ''
  const absoluteValue = Math.abs(numericValue)

  for (let index = 0; index < SMART_UNITS.length; index += 1) {
    const unit = SMART_UNITS[index]
    if (absoluteValue < unit.threshold) continue

    let scaledValue = Number((absoluteValue / unit.threshold).toFixed(2))
    let scaledUnit = unit

    if (scaledValue >= 1000 && index > 0) {
      scaledUnit = SMART_UNITS[index - 1]
      scaledValue = Number((absoluteValue / scaledUnit.threshold).toFixed(2))
    }

    return `${sign}${scaledValue.toFixed(2)} ${scaledUnit.suffix}`
  }

  return `${sign}${absoluteValue.toFixed(2)}`
}

export function iskToMillionInput(value: number | null | undefined) {
  return Number((Number(value ?? 0) / MILLION_ISK).toFixed(2))
}

export function millionInputToIsk(value: number | null | undefined) {
  return Math.round(Number(Number(value ?? 0).toFixed(2)) * MILLION_ISK)
}
