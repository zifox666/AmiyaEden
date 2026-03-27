export interface EligibleRowLike {
  canApplyNow: boolean
}

export function sortEligibleRows<T extends EligibleRowLike>(rows: T[]): T[] {
  return [...rows].sort((a, b) => Number(b.canApplyNow) - Number(a.canApplyNow))
}
