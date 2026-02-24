package model

import "time"

// OperationLog API 操作日志
type OperationLog struct {
	ID         uint      `gorm:"primarykey"               json:"id"`
	RequestID  string    `gorm:"size:64;index"            json:"request_id"`
	UserID     uint      `gorm:"default:0;index"          json:"user_id"`
	Username   string    `gorm:"size:64;default:''"       json:"username"`
	IP         string    `gorm:"size:64"                  json:"ip"`
	Method     string    `gorm:"size:16"                  json:"method"`
	Path       string    `gorm:"size:256"                 json:"path"`
	Query      string    `gorm:"size:512"                 json:"query"`
	StatusCode int       `gorm:"index"                    json:"status_code"`
	BizCode    int       `gorm:"default:0"                json:"biz_code"`
	LatencyMs  int64     `gorm:"comment:'耗时(ms)'"        json:"latency_ms"`
	UserAgent  string    `gorm:"size:256"                 json:"user_agent"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index"     json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_log"
}
