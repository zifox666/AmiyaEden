// Package jwt 提供基于 HMAC-SHA256 的 JWT Token 生成与解析（无外部依赖）
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("令牌无效")
	ErrExpiredToken = errors.New("令牌已过期")
)

var secret []byte

// Init 初始化 JWT 密钥，须在应用启动时调用
func Init(s string) {
	secret = []byte(s)
}

// header JWT 头部（固定）
var headerEncoded = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Claims 我们系统的 JWT 载荷
type Claims struct {
	UserID      uint  `json:"uid"`
	CharacterID int64 `json:"cid"` // 登录时使用的 EVE 角色 ID
	ExpiresAt   int64 `json:"exp"`
	IssuedAt    int64 `json:"iat"`
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint, characterID int64, expireDays int) (string, error) {
	claims := Claims{
		UserID:      userID,
		CharacterID: characterID,
		ExpiresAt:   time.Now().Add(time.Duration(expireDays) * 24 * time.Hour).Unix(),
		IssuedAt:    time.Now().Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)

	message := headerEncoded + "." + payloadEnc
	sig := sign(message)

	return message + "." + sig, nil
}

// ParseToken 解析并验证 JWT Token，返回 Claims
func ParseToken(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	message := parts[0] + "." + parts[1]
	expected := sign(message)
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func sign(message string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
