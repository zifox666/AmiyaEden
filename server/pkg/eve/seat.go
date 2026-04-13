// Package eve 提供 SeAT OAuth 2.0 / OpenID Connect 客户端实现
package eve

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SeatAccount SeAT 用户下的 EVE 角色账号
type SeatAccount struct {
	ID    int64  `json:"id"`   // EVE character_id
	Name  string `json:"name"` // 角色名
	Valid bool   `json:"valid"`
}

// SeatContact SeAT 联系方式（如 QQ）
type SeatContact struct {
	QQ *SeatQQContact `json:"qq,omitempty"`
}

// SeatQQContact QQ 联系方式
type SeatQQContact struct {
	ID   int64  `json:"id"`
	Nick string `json:"nick"`
}

// SeatUserInfo SeAT id_token JWT 解析出的用户信息
type SeatUserInfo struct {
	Sub      string        `json:"sub"`      // SeAT 内部用户 ID
	UID      string        `json:"uid"`      // 主角色 character_id（字符串）
	Name     string        `json:"nam"`      // SeAT 用户名
	Groups   []string      `json:"groups"`   // 所属组列表
	Accounts []SeatAccount `json:"acct"`     // 关联的 EVE 角色列表
	Contacts *SeatContact  `json:"contacts"` // 联系方式
	Scopes   []string      `json:"scp"`      // 授权的 scope 列表
	ExpireAt int64         `json:"exp"`      // 过期时间戳
}

// SeatPassthroughClaims passthrough access_token JWT 中的 Claims
type SeatPassthroughClaims struct {
	Sub  string   `json:"sub"`  // "character:XXXXXXXX"
	Name string   `json:"name"` // 角色名
	Scp  []string `json:"scp"`  // ESI scopes
	Exp  int64    `json:"exp"`  // 过期时间戳
}

// CharacterID 从 sub 字段（格式 "character:12345678"）解析角色 ID
func (c *SeatPassthroughClaims) CharacterID() (int64, error) {
	parts := strings.SplitN(c.Sub, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("unexpected passthrough sub format: %s", c.Sub)
	}
	var id int64
	if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
		return 0, fmt.Errorf("parse character_id from sub: %w", err)
	}
	return id, nil
}

// SeatTokenResponse SeAT OAuth Token 响应
type SeatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
}

// SeatClient SeAT OAuth 客户端
type SeatClient struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	HTTPClient   *http.Client
}

// NewSeatClient 创建 SeAT OAuth 客户端
func NewSeatClient(baseURL, clientID, clientSecret, callbackURL string) *SeatClient {
	// 去掉末尾斜杠
	baseURL = strings.TrimRight(baseURL, "/")
	return &SeatClient{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// BuildAuthURL 生成 SeAT 授权跳转 URL
func (c *SeatClient) BuildAuthURL(state string, scopes []string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("redirect_uri", c.CallbackURL)
	params.Set("client_id", c.ClientID)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	return c.BaseURL + "/oauth/authorize?" + params.Encode()
}

// ExchangeCode 用授权码换取 Token
func (c *SeatClient) ExchangeCode(ctx context.Context, code string) (*SeatTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.CallbackURL)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	return c.doTokenRequest(ctx, data)
}

// RefreshAccessToken 刷新 Access Token
func (c *SeatClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*SeatTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	return c.doTokenRequest(ctx, data)
}

func (c *SeatClient) doTokenRequest(ctx context.Context, data url.Values) (*SeatTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SeAT OAuth error %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp SeatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	return &tokenResp, nil
}

// GetUserInfo 获取 SeAT 用户信息（通过 /oauth/userinfo）
// Deprecated: 推荐使用 ParseIDToken 直接从 id_token 解析，无需网络请求
func (c *SeatClient) GetUserInfo(ctx context.Context, accessToken string) (*SeatUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/oauth/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("build userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SeAT userinfo error %d: %s", resp.StatusCode, string(body))
	}

	var info SeatUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo: %w", err)
	}
	return &info, nil
}

// ParseIDToken 从 id_token JWT 中解析用户信息（不验证签名）
// id_token 包含 sub, uid, nam, acct, scp, groups, contacts, exp
func ParseIDToken(idToken string) (*SeatUserInfo, error) {
	payload, err := decodeJWTPayload(idToken)
	if err != nil {
		return nil, fmt.Errorf("decode id_token: %w", err)
	}
	var info SeatUserInfo
	if err := json.Unmarshal(payload, &info); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}
	return &info, nil
}

// ParsePassthroughToken 从 passthrough access_token JWT 中解析角色信息（不验证签名）
// passthrough token 包含 sub("character:XXXX"), name, scp, exp
func ParsePassthroughToken(accessToken string) (*SeatPassthroughClaims, error) {
	payload, err := decodeJWTPayload(accessToken)
	if err != nil {
		return nil, fmt.Errorf("decode passthrough token: %w", err)
	}
	var claims SeatPassthroughClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse passthrough claims: %w", err)
	}
	return &claims, nil
}

// decodeJWTPayload 解析 JWT 的 payload 部分（不验证签名）
func decodeJWTPayload(token string) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}
	// Base64url 解码（JWT 使用 RawURLEncoding，无填充）
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decode JWT payload: %w", err)
	}
	return payload, nil
}

// GetPassthroughToken 用 SeAT access_token 为指定角色换取 ESI access_token
// 返回的 token 可直接用于请求 https://esi.evetech.net
// esiScopes 为空时不传 scope 参数（SeAT 会使用该角色已授权的全部 ESI scopes）
func (c *SeatClient) GetPassthroughToken(ctx context.Context, characterID int64, seatAccessToken string, esiScopes []string) (*SeatTokenResponse, error) {
	data := url.Values{}
	if len(esiScopes) > 0 {
		// 只保留 ESI scopes（publicData 或 esi- 前缀）
		var filtered []string
		for _, s := range esiScopes {
			if s == "publicData" || strings.HasPrefix(s, "esi-") {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > 0 {
			data.Set("scope", strings.Join(filtered, " "))
		}
	}

	endpoint := fmt.Sprintf("%s/oauth/passthrough/%d", c.BaseURL, characterID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build passthrough request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+seatAccessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("passthrough request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read passthrough response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SeAT passthrough error %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp SeatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse passthrough response: %w", err)
	}
	return &tokenResp, nil
}
