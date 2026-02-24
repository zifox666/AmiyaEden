// Package eve 提供 EVE Online SSO OAuth 2.0 客户端实现
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

const (
	// AuthorizeURL EVE SSO 授权地址
	AuthorizeURL = "https://login.eveonline.com/v2/oauth/authorize"
	// TokenURL EVE SSO Token 交换地址
	TokenURL = "https://login.eveonline.com/v2/oauth/token"
	// PortraitURLFmt EVE 角色头像 URL 模板
	PortraitURLFmt = "https://images.evetech.net/characters/%d/portrait?size=128"
)

// Client EVE SSO OAuth 客户端
type Client struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	HTTPClient   *http.Client
}

// NewClient 创建 EVE SSO 客户端
func NewClient(clientID, clientSecret, callbackURL string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// BuildAuthURL 生成 EVE SSO 授权跳转 URL
func (c *Client) BuildAuthURL(state string, scopes []string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("redirect_uri", c.CallbackURL)
	params.Set("client_id", c.ClientID)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	return AuthorizeURL + "?" + params.Encode()
}

// TokenResponse EVE SSO Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ExchangeCode 用授权码换取 Token
func (c *Client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.CallbackURL)
	return c.doTokenRequest(ctx, data)
}

// RefreshAccessToken 刷新 Access Token
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	return c.doTokenRequest(ctx, data)
}

func (c *Client) doTokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.SetBasicAuth(c.ClientID, c.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("EVE SSO error %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	return &tokenResp, nil
}

// JWTClaims EVE SSO v2 JWT access_token 解析后的载荷
type JWTClaims struct {
	Sub    string      `json:"sub"`   // "CHARACTER:EVE:12345678"
	Name   string      `json:"name"`  // 角色名
	Owner  string      `json:"owner"` // 角色所有者哈希
	Scopes interface{} `json:"scp"`   // string 或 []interface{}
	Exp    int64       `json:"exp"`
	Iss    string      `json:"iss"`
	Azp    string      `json:"azp"`
}

// ParseAccessToken 解码 EVE SSO JWT access_token 载荷（不验签，Token 来自 EVE 服务器 HTTPS 响应，安全可信）
func ParseAccessToken(accessToken string) (*JWTClaims, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// 补全 Base64URL padding
	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// 尝试 RawURLEncoding
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("decode JWT payload: %w", err)
		}
	}

	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("parse JWT claims: %w", err)
	}
	return &claims, nil
}

// GetCharacterID 从 sub 字段 "CHARACTER:EVE:12345678" 中提取角色 ID
func (c *JWTClaims) GetCharacterID() (int64, error) {
	parts := strings.Split(c.Sub, ":")
	if len(parts) < 3 || parts[0] != "CHARACTER" {
		return 0, fmt.Errorf("invalid EVE sub format: %s", c.Sub)
	}
	var id int64
	if _, err := fmt.Sscanf(parts[2], "%d", &id); err != nil {
		return 0, fmt.Errorf("parse character id: %w", err)
	}
	return id, nil
}

// GetScopes 将 scp 字段（可能是 string 或 []interface{}）统一返回 []string
func (c *JWTClaims) GetScopes() []string {
	if c.Scopes == nil {
		return nil
	}
	switch v := c.Scopes.(type) {
	case string:
		return []string{v}
	case []interface{}:
		scopes := make([]string, 0, len(v))
		for _, s := range v {
			if str, ok := s.(string); ok {
				scopes = append(scopes, str)
			}
		}
		return scopes
	}
	return nil
}

// PortraitURL 生成角色头像 URL
func PortraitURL(characterID int64) string {
	return fmt.Sprintf(PortraitURLFmt, characterID)
}
