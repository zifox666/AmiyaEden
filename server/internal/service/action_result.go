package service

import "amiya-eden/internal/model"

type MailActionResult struct {
	MailError string `json:"mail_error,omitempty"`
}

type ShopOrderActionResult struct {
	model.ShopOrder
	MailError string `json:"mail_error,omitempty"`
}

type SrpPayoutActionResult struct {
	model.SrpApplication
	MailError string `json:"mail_error,omitempty"`
}

type SrpBatchPayoutActionResult struct {
	SrpBatchPayoutSummaryResponse
	MailError string `json:"mail_error,omitempty"`
}

type SrpBatchFuxiPayoutActionResult struct {
	SrpBatchFuxiPayoutSummary
	MailError string `json:"mail_error,omitempty"`
}
