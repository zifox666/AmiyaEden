package esi

import (
	"amiya-eden/global"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  常量
// ─────────────────────────────────────────────

const (
	// rateLimitSafetyMargin 安全余量：剩余配额低于此值时开始节流
	rateLimitSafetyMargin = 10
	// rateLimitPauseThreshold 暂停阈值：剩余配额低于此值时暂停请求
	rateLimitPauseThreshold = 5
	// rateLimitMaxWait 单次限速等待的最长时间
	rateLimitMaxWait = 60 * time.Second
	// maxRetries 遇到 420 限速错误时的最大重试次数
	maxRetries = 3
	// retryBaseDelay 重试初始退避时间
	retryBaseDelay = 2 * time.Second
	// paginationConcurrency 分页并发拉取数
	paginationConcurrency = 10
)

// ─────────────────────────────────────────────
//  响应元数据
// ─────────────────────────────────────────────

// ResponseMeta ESI 响应头中提取的元数据
type ResponseMeta struct {
	StatusCode      int    `json:"status_code"`
	CacheStatus     string `json:"cache_status"`      // x-esi-cache-status: HIT / MISS
	RequestID       string `json:"request_id"`        // x-esi-request-id
	Pages           int    `json:"pages"`             // x-pages: 总页数（分页端点）
	RateLimitGroup  string `json:"ratelimit_group"`   // x-ratelimit-group
	RateLimitLimit  string `json:"ratelimit_limit"`   // x-ratelimit-limit (如 "150/15m")
	RateLimitRemain int    `json:"ratelimit_remain"`  // x-ratelimit-remaining
	RateLimitUsed   int    `json:"ratelimit_used"`    // x-ratelimit-used
	ETag            string `json:"etag,omitempty"`    // ETag
	Expires         string `json:"expires,omitempty"` // Expires
}

// parseResponseMeta 从 HTTP 响应头提取元数据
func parseResponseMeta(resp *http.Response) *ResponseMeta {
	meta := &ResponseMeta{
		StatusCode:     resp.StatusCode,
		CacheStatus:    resp.Header.Get("X-Esi-Cache-Status"),
		RequestID:      resp.Header.Get("X-Esi-Request-Id"),
		RateLimitGroup: resp.Header.Get("X-Ratelimit-Group"),
		RateLimitLimit: resp.Header.Get("X-Ratelimit-Limit"),
		ETag:           resp.Header.Get("ETag"),
		Expires:        resp.Header.Get("Expires"),
	}

	if pages := resp.Header.Get("X-Pages"); pages != "" {
		if n, err := strconv.Atoi(pages); err == nil {
			meta.Pages = n
		}
	}
	if remain := resp.Header.Get("X-Ratelimit-Remaining"); remain != "" {
		if n, err := strconv.Atoi(remain); err == nil {
			meta.RateLimitRemain = n
		}
	}
	if used := resp.Header.Get("X-Ratelimit-Used"); used != "" {
		if n, err := strconv.Atoi(used); err == nil {
			meta.RateLimitUsed = n
		}
	}

	return meta
}

// ─────────────────────────────────────────────
//  限速器
// ─────────────────────────────────────────────

// RateLimiter 根据 ESI 响应头（x-ratelimit-*）自动追踪各限速组的配额
// 当剩余配额不足时自动等待，避免触发 420 限速
type RateLimiter struct {
	mu     sync.Mutex
	groups map[string]*groupState
}

// groupState 单个限速组的状态
type groupState struct {
	remaining int           // 当前窗口剩余次数
	limit     int           // 窗口总限制
	window    time.Duration // 窗口时长
	resetAt   time.Time     // 估算的窗口重置时间
	updatedAt time.Time     // 上次更新时间
}

// NewRateLimiter 创建限速器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		groups: make(map[string]*groupState),
	}
}

// Update 根据响应元数据更新限速状态
func (rl *RateLimiter) Update(meta *ResponseMeta) {
	if meta == nil || meta.RateLimitGroup == "" {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	gs, ok := rl.groups[meta.RateLimitGroup]
	if !ok {
		gs = &groupState{}
		rl.groups[meta.RateLimitGroup] = gs
	}

	gs.remaining = meta.RateLimitRemain
	gs.updatedAt = time.Now()

	// 解析限制，如 "150/15m"
	if meta.RateLimitLimit != "" {
		if parts := strings.SplitN(meta.RateLimitLimit, "/", 2); len(parts) == 2 {
			if n, err := strconv.Atoi(parts[0]); err == nil {
				gs.limit = n
			}
			if dur, err := time.ParseDuration(parts[1]); err == nil {
				gs.window = dur
				// 当 used=1 时（第一次请求），推算窗口重置时间
				if meta.RateLimitUsed <= 1 {
					gs.resetAt = time.Now().Add(dur)
				}
			}
		} else {
			// 纯数字格式
			if n, err := strconv.Atoi(meta.RateLimitLimit); err == nil {
				gs.limit = n
			}
		}
	}

	// 如果没有 resetAt 但有 window，根据已用量合理推算
	if gs.resetAt.IsZero() && gs.window > 0 && gs.limit > 0 {
		// 粗略估算：假设窗口中消耗均匀
		elapsed := gs.window * time.Duration(meta.RateLimitUsed) / time.Duration(gs.limit)
		gs.resetAt = time.Now().Add(gs.window - elapsed)
	}
}

// Wait 如果对应限速组的配额不足，阻塞等待直到安全
// 如果 group 为空或无状态记录则直接返回
func (rl *RateLimiter) Wait(ctx context.Context, group string) error {
	if group == "" {
		return nil
	}

	rl.mu.Lock()
	gs, ok := rl.groups[group]
	if !ok {
		rl.mu.Unlock()
		return nil
	}

	remaining := gs.remaining
	resetAt := gs.resetAt
	rl.mu.Unlock()

	// 窗口已过重置时间，配额应已恢复
	if !resetAt.IsZero() && time.Now().After(resetAt) {
		return nil
	}

	// 配额充足
	if remaining > rateLimitSafetyMargin {
		return nil
	}

	// 配额偏低但未到暂停阈值，短暂节流
	if remaining > rateLimitPauseThreshold {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 配额非常低，需要等待窗口重置
	wait := time.Until(resetAt)
	if wait <= 0 {
		// 重置时间未知或已过期，固定等待
		wait = 5 * time.Second
	}
	if wait > rateLimitMaxWait {
		wait = rateLimitMaxWait
	}

	global.Logger.Warn("[ESI RateLimit] 配额不足，等待窗口重置",
		zap.String("group", group),
		zap.Int("remaining", remaining),
		zap.Duration("wait", wait),
	)

	select {
	case <-time.After(wait):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ─────────────────────────────────────────────
//  增强请求方法
// ─────────────────────────────────────────────

// doRequest 底层 GET 请求：发送请求、读取响应、解析元数据、限速更新
// 遇到 420 限速错误会自动指数退避重试
func (c *Client) doRequest(ctx context.Context, url string, accessToken string) ([]byte, *ResponseMeta, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 构造请求
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("build ESI request: %w", err)
		}
		if accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		}
		req.Header.Set("Accept", "application/json")

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("ESI request %s: %w", url, err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, nil, fmt.Errorf("read ESI response: %w", readErr)
		}

		meta := parseResponseMeta(resp)

		// 更新限速器
		if c.rateLimiter != nil {
			c.rateLimiter.Update(meta)
		}

		// 420 限速：指数退避重试
		if resp.StatusCode == 420 {
			delay := retryBaseDelay * time.Duration(1<<uint(attempt))
			if delay > rateLimitMaxWait {
				delay = rateLimitMaxWait
			}
			global.Logger.Warn("[ESI RateLimit] 收到 420 限速，退避重试",
				zap.String("url", url),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
			)
			lastErr = fmt.Errorf("ESI rate limited (420) on %s", url)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, meta, ctx.Err()
			}
		}

		return body, meta, nil
	}

	return nil, nil, fmt.Errorf("ESI rate limited after %d retries: %w", maxRetries, lastErr)
}

// GetWithMeta 发起 GET 请求并返回解码后的数据和响应元数据
// 相比 Get，额外返回 ResponseMeta 以便调用方感知缓存/限速状态
func (c *Client) GetWithMeta(ctx context.Context, path string, accessToken string, dest interface{}) (*ResponseMeta, error) {
	url := c.baseURL + path
	body, meta, err := c.doRequest(ctx, url, accessToken)
	if err != nil {
		return meta, fmt.Errorf("ESI GET %s: %w", path, err)
	}

	if meta.StatusCode == http.StatusNotModified {
		return meta, ErrNotModified
	}
	if meta.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("ESI error %d on %s: %s", meta.StatusCode, path, string(body))
	}

	if dest != nil {
		if err := json.Unmarshal(body, dest); err != nil {
			return meta, fmt.Errorf("decode ESI response: %w", err)
		}
	}

	return meta, nil
}

// GetPaginated 自动处理分页的 GET 请求
//
// dest 必须是指向切片的指针（如 *[]AssetItem），会自动获取所有页面并合并为完整切片。
// 内部流程：
//  1. 请求第 1 页，从 x-pages 响应头获取总页数
//  2. 并发拉取剩余页面（受限速器约束）
//  3. 合并所有页面的 JSON 数组
//
// 返回第一页的 ResponseMeta 和任何错误
func (c *Client) GetPaginated(ctx context.Context, path string, accessToken string, dest interface{}) (*ResponseMeta, error) {
	// 1. 请求第一页
	page1URL := c.buildPageURL(path, 1)
	body, meta, err := c.doRequest(ctx, page1URL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("ESI GET %s page 1: %w", path, err)
	}

	if meta.StatusCode == http.StatusNotModified {
		return meta, ErrNotModified
	}
	if meta.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("ESI error %d on %s: %s", meta.StatusCode, path, string(body))
	}

	totalPages := meta.Pages
	if totalPages <= 0 {
		totalPages = 1
	}

	// 单页直接解码返回
	if totalPages == 1 {
		if dest != nil {
			if err := json.Unmarshal(body, dest); err != nil {
				return meta, fmt.Errorf("decode ESI response: %w", err)
			}
		}
		return meta, nil
	}

	// 2. 多页：并发拉取剩余页面
	global.Logger.Debug("[ESI Paginated] 检测到多页响应，开始并发拉取",
		zap.String("path", path),
		zap.Int("total_pages", totalPages),
		zap.String("ratelimit_group", meta.RateLimitGroup),
		zap.Int("ratelimit_remaining", meta.RateLimitRemain),
	)

	allBodies := make([][]byte, totalPages)
	allBodies[0] = body

	var (
		mu       sync.Mutex
		fetchErr error
	)
	sem := make(chan struct{}, paginationConcurrency)
	var wg sync.WaitGroup

	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()

			// 限速等待（使用第 1 页获知的 group）
			if c.rateLimiter != nil && meta.RateLimitGroup != "" {
				if waitErr := c.rateLimiter.Wait(ctx, meta.RateLimitGroup); waitErr != nil {
					mu.Lock()
					if fetchErr == nil {
						fetchErr = fmt.Errorf("rate limit wait for page %d: %w", p, waitErr)
					}
					mu.Unlock()
					return
				}
			}

			pageURL := c.buildPageURL(path, p)
			pageBody, pageMeta, reqErr := c.doRequest(ctx, pageURL, accessToken)
			if reqErr != nil {
				mu.Lock()
				if fetchErr == nil {
					fetchErr = fmt.Errorf("fetch page %d of %s: %w", p, path, reqErr)
				}
				mu.Unlock()
				return
			}

			if pageMeta.StatusCode != http.StatusOK {
				mu.Lock()
				if fetchErr == nil {
					fetchErr = fmt.Errorf("ESI error %d on page %d of %s: %s",
						pageMeta.StatusCode, p, path, string(pageBody))
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			allBodies[p-1] = pageBody
			mu.Unlock()
		}(page)
	}

	wg.Wait()

	if fetchErr != nil {
		return meta, fetchErr
	}

	// 3. 合并所有 JSON 数组
	if dest != nil {
		merged, mergeErr := mergeJSONArrays(allBodies)
		if mergeErr != nil {
			return meta, fmt.Errorf("merge paginated results for %s: %w", path, mergeErr)
		}
		if err := json.Unmarshal(merged, dest); err != nil {
			return meta, fmt.Errorf("decode merged ESI response for %s: %w", path, err)
		}
	}

	global.Logger.Debug("[ESI Paginated] 分页数据合并完成",
		zap.String("path", path),
		zap.Int("total_pages", totalPages),
	)

	return meta, nil
}

// ─────────────────────────────────────────────
//  辅助方法
// ─────────────────────────────────────────────

// buildPageURL 构建带分页参数的完整 URL
func (c *Client) buildPageURL(path string, page int) string {
	base := c.baseURL + path
	if strings.Contains(path, "?") {
		return fmt.Sprintf("%s&page=%d", base, page)
	}
	return fmt.Sprintf("%s?page=%d", base, page)
}

// mergeJSONArrays 将多个 JSON 数组合并为一个
func mergeJSONArrays(arrays [][]byte) ([]byte, error) {
	var merged []json.RawMessage
	for i, data := range arrays {
		if data == nil {
			return nil, fmt.Errorf("page %d data is nil", i+1)
		}
		var page []json.RawMessage
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("unmarshal page %d: %w", i+1, err)
		}
		merged = append(merged, page...)
	}
	return json.Marshal(merged)
}
