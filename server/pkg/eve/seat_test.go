package eve

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestGetPassthroughTokenClassifiesNotFound(t *testing.T) {
	tests := []struct {
		name         string
		response     fakeSSOResponse
		wantNotFound bool
	}{
		{"404 表示 SeAT 拒绝该角色", fakeSSOResponse{status: 404, body: `{"message":""}`}, true},
		{"500 是 SeAT 故障", fakeSSOResponse{status: 500, body: `{"message":"Server Error"}`}, false},
		{"429 是限流", fakeSSOResponse{status: 429, body: `{"message":"Too Many Attempts."}`}, false},
		{"网络错误", fakeSSOResponse{err: errors.New("connection reset by peer")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSeatClient("https://seat.example", "id", "secret", "http://callback")
			client.HTTPClient.Transport = &fakeSSOTransport{responses: []fakeSSOResponse{tt.response}}

			_, err := client.GetPassthroughToken(context.Background(), 1, "token", nil)
			if err == nil {
				t.Fatal("want error")
			}
			// 调用方会再包一层，判断必须能穿透
			wrapped := fmt.Errorf("SeAT passthrough 角色 1 失败: %w", err)
			if got := IsSeatPassthroughNotFound(wrapped); got != tt.wantNotFound {
				t.Errorf("IsSeatPassthroughNotFound = %v, want %v (err: %v)", got, tt.wantNotFound, err)
			}
		})
	}
}
