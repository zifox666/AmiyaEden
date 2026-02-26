// Package esi 提供 EVE ESI API 客户端及数据刷新队列
package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// BaseURL ESI API 基础地址
	BaseURL = "https://esi.evetech.net"
	// DefaultTimeout HTTP 默认超时
	DefaultTimeout = 30 * time.Second
)

// Client ESI HTTP 客户端
type Client struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewClient 创建 ESI 客户端
func NewClient() *Client {
	return &Client{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		rateLimiter: NewRateLimiter(),
	}
}

// Get 发起带认证的 GET 请求并将响应 JSON 解码到 dest
func (c *Client) Get(ctx context.Context, path string, accessToken string, dest interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ESI request %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ESI response: %w", err)
	}

	if resp.StatusCode == http.StatusNotModified {
		return ErrNotModified
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESI error %d on %s: %s", resp.StatusCode, path, string(body))
	}

	if dest != nil {
		if err := json.Unmarshal(body, dest); err != nil {
			return fmt.Errorf("decode ESI response: %w", err)
		}
	}
	return nil
}

// GetRaw 发起带认证的 GET 请求并返回原始 JSON 字节
func (c *Client) GetRaw(ctx context.Context, path string, accessToken string) ([]byte, int, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("ESI request %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read ESI response: %w", err)
	}
	return body, resp.StatusCode, nil
}

// PostJSON 发起带认证的 POST 请求（JSON body）并将响应解码到 dest
func (c *Client) PostJSON(ctx context.Context, path string, accessToken string, reqBody interface{}, dest interface{}) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(
		io.Reader(readerFromBytes(bodyBytes)),
	))
	if err != nil {
		return fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ESI POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ESI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESI error %d on POST %s: %s", resp.StatusCode, path, string(respBody))
	}

	if dest != nil {
		if err := json.Unmarshal(respBody, dest); err != nil {
			return fmt.Errorf("decode ESI response: %w", err)
		}
	}
	return nil
}

// PutJSON 发起带认证的 PUT 请求（JSON body）
func (c *Client) PutJSON(ctx context.Context, path string, accessToken string, reqBody interface{}) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, io.NopCloser(
		io.Reader(readerFromBytes(bodyBytes)),
	))
	if err != nil {
		return fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ESI PUT %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ESI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ESI error %d on PUT %s: %s", resp.StatusCode, path, string(respBody))
	}
	return nil
}

// PostNoContent 发起带认证的 POST 请求，期望 2xx 响应（如 201 / 204）
func (c *Client) PostNoContent(ctx context.Context, path string, accessToken string, reqBody interface{}) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(
		io.Reader(readerFromBytes(bodyBytes)),
	))
	if err != nil {
		return fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ESI POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ESI response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ESI error %d on POST %s: %s", resp.StatusCode, path, string(respBody))
	}
	return nil
}

// readerFromBytes 创建一个 bytes reader
func readerFromBytes(b []byte) io.Reader {
	return &bytesReader{data: b}
}

type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
