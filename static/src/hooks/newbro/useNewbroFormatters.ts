import { formatIskPlain, formatTime } from '@utils/common'

const decimalFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2
})

const percentageFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 0,
  maximumFractionDigits: 2
})

export function formatNewbroDateTime(value?: string | null): string {
  return formatTime(value)
}

export function formatNewbroIsk(value: number): string {
  return formatIskPlain(value)
}

export function formatNewbroCredit(value: number): string {
  return decimalFormatter.format(value)
}

export function formatNewbroPercentage(value: number): string {
  return `${percentageFormatter.format(value)}%`
}

export function useNewbroFormatters() {
  return {
    formatDateTime: formatNewbroDateTime,
    formatIsk: formatNewbroIsk,
    formatCredit: formatNewbroCredit,
    formatPercentage: formatNewbroPercentage
  }
}
