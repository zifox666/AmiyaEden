// Package esi 提供 EVE ESI API 客户端及数据刷新队列
package esi

import (
	"amiya-eden/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL   = config.DefaultESIBaseURL
	APIPrefix = config.DefaultESIAPIPrefix
	// DefaultTimeout HTTP 默认超时
	DefaultTimeout = 30 * time.Second
)

// Client ESI HTTP 客户端
type Client struct {
	baseURL     string
	apiPrefix   string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewClient 创建 ESI 客户端
func NewClient() *Client {
	return NewClientWithConfig(BaseURL, APIPrefix)
}

func NewClientWithConfig(baseURL, apiPrefix string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiPrefix: normalizePrefix(apiPrefix),
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		rateLimiter: NewRateLimiter(),
	}
}

func normalizePrefix(prefix string) string {
	p := strings.TrimSpace(prefix)
	p = strings.TrimRight(p, "/")
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func (c *Client) buildURL(path string) string {
	p := strings.TrimSpace(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if c.apiPrefix != "" && !strings.HasPrefix(p, c.apiPrefix+"/") && p != c.apiPrefix {
		p = c.apiPrefix + p
	}
	return c.baseURL + p
}

func (c *Client) newAuthorizedRequest(method, path string, accessToken string, body io.Reader, contentType string) (*http.Request, error) {
	url := c.buildURL(path)
	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build ESI request: %w", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func (c *Client) performRequest(req *http.Request, path, errPrefix string, maxBytes int64) ([]byte, *http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s %s: %w", errPrefix, path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := readResponseBody(resp, maxBytes)
	if err != nil {
		return nil, resp, fmt.Errorf("read ESI response: %w", err)
	}

	return body, resp, nil
}

func readResponseBody(resp *http.Response, maxBytes int64) ([]byte, error) {
	reader := io.Reader(resp.Body)
	if maxBytes > 0 {
		reader = io.LimitReader(resp.Body, maxBytes+1)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if maxBytes > 0 && int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return body, nil
}

// Get 发起带认证的 GET 请求并将响应 JSON 解码到 dest
func (c *Client) Get(ctx context.Context, path string, accessToken string, dest interface{}) error {
	req, err := c.newAuthorizedRequest(http.MethodGet, path, accessToken, nil, "")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	body, resp, err := c.performRequest(req, path, "ESI request", 0)
	if err != nil {
		return err
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
	req, err := c.newAuthorizedRequest(http.MethodGet, path, accessToken, nil, "")
	if err != nil {
		return nil, 0, err
	}
	req = req.WithContext(ctx)

	body, resp, err := c.performRequest(req, path, "ESI request", 0)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		}
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}

// PostJSON 发起带认证的 POST 请求（JSON body）并将响应解码到 dest
func (c *Client) PostJSON(ctx context.Context, path string, accessToken string, reqBody interface{}, dest interface{}) error {
	return c.postJSON(ctx, path, accessToken, reqBody, dest, 0)
}

// PostJSONWithLimit 发起带认证的 POST 请求（JSON body），并限制最大响应字节数
func (c *Client) PostJSONWithLimit(ctx context.Context, path string, accessToken string, reqBody interface{}, dest interface{}, maxBytes int64) error {
	return c.postJSON(ctx, path, accessToken, reqBody, dest, maxBytes)
}

func (c *Client) postJSON(ctx context.Context, path string, accessToken string, reqBody interface{}, dest interface{}, maxBytes int64) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	req, err := c.newAuthorizedRequest(http.MethodPost, path, accessToken, bytes.NewReader(bodyBytes), "application/json")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	respBody, resp, err := c.performRequest(req, path, "ESI POST", maxBytes)
	if err != nil {
		return err
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

// PostCreatedJSON 发起带认证的 POST 请求（JSON body），期望 201 Created 并解码响应
func (c *Client) PostCreatedJSON(ctx context.Context, path string, accessToken string, reqBody interface{}, dest interface{}) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	req, err := c.newAuthorizedRequest(http.MethodPost, path, accessToken, bytes.NewReader(bodyBytes), "application/json")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	respBody, resp, err := c.performRequest(req, path, "ESI POST", 0)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
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

	req, err := c.newAuthorizedRequest(http.MethodPut, path, accessToken, bytes.NewReader(bodyBytes), "application/json")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	respBody, resp, err := c.performRequest(req, path, "ESI PUT", 0)
	if err != nil {
		return err
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

	req, err := c.newAuthorizedRequest(http.MethodPost, path, accessToken, bytes.NewReader(bodyBytes), "application/json")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	respBody, resp, err := c.performRequest(req, path, "ESI POST", 0)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ESI error %d on POST %s: %s", resp.StatusCode, path, string(respBody))
	}
	return nil
}

// Delete 发起带认证的 DELETE 请求
func (c *Client) Delete(ctx context.Context, path string, accessToken string) error {
	req, err := c.newAuthorizedRequest(http.MethodDelete, path, accessToken, nil, "")
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	respBody, resp, err := c.performRequest(req, path, "ESI DELETE", 0)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ESI error %d on DELETE %s: %s", resp.StatusCode, path, string(respBody))
	}
	return nil
}
