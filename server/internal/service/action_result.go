package service

import "amiya-eden/internal/model"

type MailActionResult struct {
	MailAttemptSummary
}

type ShopOrderActionResult struct {
	model.ShopOrder
	MailAttemptSummary
}

type SrpPayoutActionResult struct {
	model.SrpApplication
	MailAttemptSummary
}

type SrpBatchPayoutActionResult struct {
	SrpBatchPayoutSummaryResponse
	MailAttemptSummary
}

type SrpBatchFuxiPayoutActionResult struct {
	SrpBatchFuxiPayoutSummary
	MailAttemptSummary
}
