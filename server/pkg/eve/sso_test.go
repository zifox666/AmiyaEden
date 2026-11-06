package eve

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// fakeSSOResponse 为 nil err 时返回 status/body，否则模拟网络错误
type fakeSSOResponse struct {
	status int
	body   string
	err    error
}

type fakeSSOTransport struct {
	responses []fakeSSOResponse
	calls     int
}

func (t *fakeSSOTransport) RoundTrip(*http.Request) (*http.Response, error) {
	r := t.responses[t.calls]
	t.calls++
	if r.err != nil {
		return nil, r.err
	}
	return &http.Response{StatusCode: r.status, Body: io.NopCloser(strings.NewReader(r.body)), Header: http.Header{}}, nil
}

func newFakeClient(responses ...fakeSSOResponse) (*Client, *fakeSSOTransport) {
	transport := &fakeSSOTransport{responses: responses}
	client := NewClient("id", "secret", "http://callback")
	client.HTTPClient = &http.Client{Transport: transport}
	return client, transport
}

func TestRefreshAccessToken(t *testing.T) {
	originalBackoff := refreshRetryBackoff
	refreshRetryBackoff = time.Millisecond
	defer func() { refreshRetryBackoff = originalBackoff }()

	success := fakeSSOResponse{status: 200, body: `{"access_token":"new","refresh_token":"rt","expires_in":1199}`}
	badGateway := fakeSSOResponse{status: 502, body: "<!DOCTYPE HTML><HTML>Bad Gateway</HTML>"}
	networkErr := fakeSSOResponse{err: errors.New("connection reset by peer")}
	invalidGrant := fakeSSOResponse{status: 400, body: `{"error":"invalid_grant","error_description":"Invalid refresh token."}`}

	tests := []struct {
		name        string
		responses   []fakeSSOResponse
		wantCalls   int
		wantErr     bool
		wantRevoked bool
	}{
		{"成功不重试", []fakeSSOResponse{success}, 1, false, false},
		{"502 后重试成功", []fakeSSOResponse{badGateway, success}, 2, false, false},
		{"网络错误后重试成功", []fakeSSOResponse{networkErr, networkErr, success}, 3, false, false},
		{"持续 5xx 不算失效", []fakeSSOResponse{badGateway, badGateway, badGateway}, 3, true, false},
		{"持续网络错误不算失效", []fakeSSOResponse{networkErr, networkErr, networkErr}, 3, true, false},
		{"invalid_grant 算失效且不重试", []fakeSSOResponse{invalidGrant}, 1, true, true},
		{"invalid_client 是配置问题不算失效", []fakeSSOResponse{{status: 401, body: `{"error":"invalid_client"}`}}, 1, true, false},
		{"429 不重试也不算失效", []fakeSSOResponse{{status: 429, body: `{"error":"too many requests"}`}}, 1, true, false},
		{"非 JSON 的 400 不算失效", []fakeSSOResponse{{status: 400, body: "bad request"}}, 1, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, transport := newFakeClient(tt.responses...)
			_, err := client.RefreshAccessToken(context.Background(), "old")

			if transport.calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", transport.calls, tt.wantCalls)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got := IsRefreshTokenRevoked(err); got != tt.wantRevoked {
				t.Errorf("IsRefreshTokenRevoked = %v, want %v (err: %v)", got, tt.wantRevoked, err)
			}
		})
	}
}

func TestRefreshAccessTokenStopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, transport := newFakeClient(fakeSSOResponse{err: context.Canceled})
	_, err := client.RefreshAccessToken(ctx, "old")

	if err == nil || IsRefreshTokenRevoked(err) {
		t.Fatalf("err = %v, want non-revoked error", err)
	}
	if transport.calls > 1 {
		t.Errorf("calls = %d, want at most 1", transport.calls)
	}
}

func TestTokenErrorTruncatesLongBody(t *testing.T) {
	err := newTokenError(502, []byte(strings.Repeat("x", tokenErrorBodyLimit+100)))
	if len(err.Body) != tokenErrorBodyLimit {
		t.Errorf("len(Body) = %d, want %d", len(err.Body), tokenErrorBodyLimit)
	}
}
