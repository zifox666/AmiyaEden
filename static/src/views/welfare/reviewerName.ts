interface WelfareHistoryReviewerNameInput {
  reviewerName: string | null | undefined
  reviewedBy: number | null | undefined
  status: string | null | undefined
  reviewedAt: string | null | undefined
  systemLabel: string
}

export function formatWelfareHistoryReviewerName({
  reviewerName,
  reviewedBy,
  status,
  reviewedAt,
  systemLabel
}: WelfareHistoryReviewerNameInput) {
  const trimmedReviewerName = reviewerName?.trim() ?? ''
  if (trimmedReviewerName) return trimmedReviewerName

  if (reviewedBy === 0 && status === 'delivered' && reviewedAt) {
    return systemLabel
  }

  return '-'
}
