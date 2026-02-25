package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/cache"
	"amiya-eden/pkg/eve"
	"amiya-eden/pkg/jwt"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────
//  Scope 注册机制（供其他模块使用）
// ─────────────────────────────────────────────

// RegisteredScope 其他模块注册的 ESI Scope
type RegisteredScope struct {
	Module      string // 注册模块名
	Scope       string // ESI scope 字符串
	Description string // 描述（向用户展示）
	Required    bool   // 是否为必选 scope
}

var (
	scopeMu          sync.RWMutex
	registeredScopes []RegisteredScope
)

// RegisterScope 供其他模块调用，注册所需的 ESI scope
//
//	module:      模块标识，如 "killmail"
//	scope:       ESI scope，如 "esi-killmails.read_killmails.v1"
//	description: 向用户展示的说明
//	required:    是否为必选（false 时用户可选择跳过）
func RegisterScope(module, scope, description string, required bool) {
	scopeMu.Lock()
	defer scopeMu.Unlock()
	registeredScopes = append(registeredScopes, RegisteredScope{
		Module:      module,
		Scope:       scope,
		Description: description,
		Required:    required,
	})
}

// GetRegisteredScopes 获取所有已注册的 scope 列表
func GetRegisteredScopes() []RegisteredScope {
	scopeMu.RLock()
	defer scopeMu.RUnlock()
	result := make([]RegisteredScope, len(registeredScopes))
	copy(result, registeredScopes)
	return result
}

// buildAllScopes 构建登录时使用的完整 scope 列表（publicData + 所有已注册 scope）
func buildLoginScopes(extraScopes []string) []string {
	scopeSet := map[string]struct{}{
		"publicData": {},
	}
	scopeMu.RLock()
	for _, rs := range registeredScopes {
		scopeSet[rs.Scope] = struct{}{}
	}
	scopeMu.RUnlock()

	for _, s := range extraScopes {
		if s != "" {
			scopeSet[s] = struct{}{}
		}
	}

	scopes := make([]string, 0, len(scopeSet))
	for s := range scopeSet {
		scopes = append(scopes, s)
	}
	return scopes
}

// ─────────────────────────────────────────────
//  SSO Service
// ─────────────────────────────────────────────

const (
	stateCachePrefix = "eve:sso:state:"
	stateCacheTTL    = 10 * time.Minute
)

// stateData OAuth state 中存储的数据
type stateData struct {
	ExtraScopes  []string `json:"extra_scopes,omitempty"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	BindToUserID uint     `json:"bind_to_user_id,omitempty"` // >0 时表示「绑定角色」流程，而非登录
}

// EveSSOService EVE SSO 业务逻辑层
type EveSSOService struct {
	charRepo  *repository.EveCharacterRepository
	userRepo  *repository.UserRepository
	eveClient *eve.Client
}

func NewEveSSOService() *EveSSOService {
	cfg := global.Config.EveSSO
	return &EveSSOService{
		charRepo:  repository.NewEveCharacterRepository(),
		userRepo:  repository.NewUserRepository(),
		eveClient: eve.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
	}
}

// GetAuthURL 生成 EVE SSO 授权 URL，并将 state 存入 Redis
//
//	extraScopes: 额外需要的 scope，传 nil 则使用所有已注册 scope
//	redirectURL: 登录成功后前端跳转地址
func (s *EveSSOService) GetAuthURL(ctx context.Context, extraScopes []string, redirectURL string) (string, error) {
	// 生成随机 state
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	// 存入 Redis
	data := stateData{ExtraScopes: extraScopes, RedirectURL: redirectURL}
	if err := cache.Set(ctx, stateCachePrefix+state, data, stateCacheTTL); err != nil {
		global.Logger.Warn("存储 SSO state 失败", zap.Error(err))
		// Redis 不可用时仍允许继续（降级）
	}

	scopes := buildLoginScopes(extraScopes)
	return s.eveClient.BuildAuthURL(state, scopes), nil
}

// GetBindAuthURL 生成「绑定新角色」的 EVE SSO 授权 URL
// 与 GetAuthURL 不同的是，state 中会记录当前登录用户 ID，回调时将角色绑到该用户下
func (s *EveSSOService) GetBindAuthURL(ctx context.Context, userID uint, extraScopes []string, redirectURL string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	data := stateData{ExtraScopes: extraScopes, RedirectURL: redirectURL, BindToUserID: userID}
	if err := cache.Set(ctx, stateCachePrefix+state, data, stateCacheTTL); err != nil {
		global.Logger.Warn("存储 SSO state 失败", zap.Error(err))
	}

	scopes := buildLoginScopes(extraScopes)
	return s.eveClient.BuildAuthURL(state, scopes), nil
}

// CallbackResult EVE SSO 回调处理结果
type CallbackResult struct {
	Token       string              `json:"token"` // 我们系统颁发的 JWT
	User        *model.User         `json:"user"`
	Character   *model.EveCharacter `json:"character"`
	RedirectURL string              `json:"redirect_url"` // 前端跳转地址（可能为空）
}

// HandleCallback 处理 EVE SSO 回调，完成 Token 交换、用户创建/更新，颁发本系统 JWT
func (s *EveSSOService) HandleCallback(ctx context.Context, code, state, clientIP string) (*CallbackResult, error) {
	if code == "" {
		return nil, errors.New("授权码不能为空")
	}

	// 读取并删除 state（防重放）
	var sd stateData
	if err := cache.Get(ctx, stateCachePrefix+state, &sd); err != nil {
		// state 不存在或已过期——记录但不强制失败（降级兼容）
		global.Logger.Warn("EVE SSO state 未找到或已过期", zap.String("state", state))
	} else {
		_ = cache.Del(ctx, stateCachePrefix+state)
	}

	// 1. 用授权码换取 Token
	tokenResp, err := s.eveClient.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 2. 解析 JWT access_token 获取角色信息
	claims, err := eve.ParseAccessToken(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}
	characterID, err := claims.GetCharacterID()
	if err != nil {
		return nil, err
	}

	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	scopesStr := strings.Join(claims.GetScopes(), " ")
	portraitURL := eve.PortraitURL(characterID)

	// 3. 查找或创建 EveCharacter
	char, err := s.charRepo.GetByCharacterID(characterID)
	now := time.Now()

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// 该角色第一次出现

		// ── 绑定流程：将新角色绑定到已登录的用户 ──
		if sd.BindToUserID > 0 {
			user, err := s.userRepo.GetByID(sd.BindToUserID)
			if err != nil {
				return nil, errors.New("绑定目标用户不存在")
			}

			char = &model.EveCharacter{
				CharacterID:   characterID,
				CharacterName: claims.Name,
				PortraitURL:   portraitURL,
				UserID:        user.ID,
				AccessToken:   tokenResp.AccessToken,
				RefreshToken:  tokenResp.RefreshToken,
				TokenExpiry:   tokenExpiry,
				Scopes:        scopesStr,
			}
			if err := s.charRepo.Create(char); err != nil {
				return nil, err
			}

			// 如果用户还没有主角色，自动设为主角色
			if user.PrimaryCharacterID == 0 {
				user.PrimaryCharacterID = characterID
				if err := s.userRepo.Update(user); err != nil {
					return nil, err
				}
			}

			jwtToken, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
			if err != nil {
				return nil, err
			}
			return &CallbackResult{Token: jwtToken, User: user, Character: char, RedirectURL: sd.RedirectURL}, nil
		}

		// ── 登录流程：首次登录，创建新用户 + 新角色 ──
		user := &model.User{
			Nickname:           claims.Name,
			Avatar:             portraitURL,
			Status:             1,
			Role:               "user",
			PrimaryCharacterID: characterID,
			LastLoginAt:        &now,
			LastLoginIP:        clientIP,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}

		char = &model.EveCharacter{
			CharacterID:   characterID,
			CharacterName: claims.Name,
			PortraitURL:   portraitURL,
			UserID:        user.ID,
			AccessToken:   tokenResp.AccessToken,
			RefreshToken:  tokenResp.RefreshToken,
			TokenExpiry:   tokenExpiry,
			Scopes:        scopesStr,
		}
		if err := s.charRepo.Create(char); err != nil {
			return nil, err
		}

		jwtToken, err := jwt.GenerateToken(user.ID, characterID, user.Role, global.Config.JWT.ExpireDay)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{Token: jwtToken, User: user, Character: char, RedirectURL: sd.RedirectURL}, nil
	}

	// 已有角色

	// ── 绑定流程：该角色已存在 ──
	if sd.BindToUserID > 0 {
		if char.UserID != sd.BindToUserID {
			return nil, errors.New("该角色已绑定到其他账号，无法再次绑定")
		}
		// 角色已属于当前用户，更新 Token 即可
		char.AccessToken = tokenResp.AccessToken
		char.RefreshToken = tokenResp.RefreshToken
		char.TokenExpiry = tokenExpiry
		char.Scopes = scopesStr
		char.CharacterName = claims.Name
		char.PortraitURL = portraitURL
		if err := s.charRepo.Update(char); err != nil {
			return nil, err
		}
		user, err := s.userRepo.GetByID(sd.BindToUserID)
		if err != nil {
			return nil, err
		}
		jwtToken, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{Token: jwtToken, User: user, Character: char, RedirectURL: sd.RedirectURL}, nil
	}

	// ── 登录流程：已有角色重新登录 ──
	char.AccessToken = tokenResp.AccessToken
	char.RefreshToken = tokenResp.RefreshToken
	char.TokenExpiry = tokenExpiry
	char.Scopes = scopesStr
	char.CharacterName = claims.Name
	char.PortraitURL = portraitURL
	if err := s.charRepo.Update(char); err != nil {
		return nil, err
	}

	// 更新用户最后登录信息
	user, err := s.userRepo.GetByID(char.UserID)
	if err != nil {
		return nil, err
	}
	user.LastLoginAt = &now
	user.LastLoginIP = clientIP
	// 如果用户还没有主角色，自动设为当前登录角色
	if user.PrimaryCharacterID == 0 {
		user.PrimaryCharacterID = characterID
	}
	// 同步头像和昵称为主角色
	user.Avatar = portraitURL
	user.Nickname = claims.Name
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	jwtToken, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return nil, err
	}
	return &CallbackResult{Token: jwtToken, User: user, Character: char, RedirectURL: sd.RedirectURL}, nil
}

// GetValidToken 获取指定角色的有效 access_token（如即将过期则自动刷新）
// 供其他模块调用，用于发起 ESI 请求
func (s *EveSSOService) GetValidToken(ctx context.Context, characterID int64) (string, error) {
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return "", err
	}

	// Token 有效期剩余 < 5 分钟则刷新
	if time.Until(char.TokenExpiry) < 5*time.Minute {
		if err := s.refreshCharacterToken(ctx, char); err != nil {
			return "", err
		}
	}

	return char.AccessToken, nil
}

// refreshCharacterToken 刷新角色 Token 并持久化
func (s *EveSSOService) refreshCharacterToken(ctx context.Context, char *model.EveCharacter) error {
	tokenResp, err := s.eveClient.RefreshAccessToken(ctx, char.RefreshToken)
	if err != nil {
		return err
	}

	claims, err := eve.ParseAccessToken(tokenResp.AccessToken)
	if err != nil {
		return err
	}

	char.AccessToken = tokenResp.AccessToken
	char.RefreshToken = tokenResp.RefreshToken
	char.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	char.Scopes = strings.Join(claims.GetScopes(), " ")

	return s.charRepo.Update(char)
}

// GetCharactersByUserID 获取用户绑定的所有 EVE 角色（不含 Token）
func (s *EveSSOService) GetCharactersByUserID(userID uint) ([]model.EveCharacter, error) {
	return s.charRepo.ListByUserID(userID)
}

// SetPrimaryCharacter 设置用户的主角色
func (s *EveSSOService) SetPrimaryCharacter(userID uint, characterID int64) error {
	// 验证该角色确实属于当前用户
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if char.UserID != userID {
		return errors.New("该角色不属于当前用户")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.PrimaryCharacterID = characterID
	user.Avatar = char.PortraitURL
	user.Nickname = char.CharacterName
	return s.userRepo.Update(user)
}

// UnbindCharacter 解除绑定某个 EVE 角色
func (s *EveSSOService) UnbindCharacter(userID uint, characterID int64) error {
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if char.UserID != userID {
		return errors.New("该角色不属于当前用户")
	}

	// 确保至少保留一个角色
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}
	if len(chars) <= 1 {
		return errors.New("至少需要保留一个角色，无法解绑")
	}

	// 如果要解绑的是主角色，自动切换到另一个角色
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user.PrimaryCharacterID == characterID {
		for _, c := range chars {
			if c.CharacterID != characterID {
				user.PrimaryCharacterID = c.CharacterID
				user.Avatar = c.PortraitURL
				user.Nickname = c.CharacterName
				break
			}
		}
		if err := s.userRepo.Update(user); err != nil {
			return err
		}
	}

	return s.charRepo.Delete(char.ID)
}

// GetRedirectURLFromState 仅读取 state 对应的前端 redirect URL（不删除 state）
// 用于错误场景下仍能跳回前端
func (s *EveSSOService) GetRedirectURLFromState(ctx context.Context, state string) string {
	if state == "" {
		return ""
	}
	var sd stateData
	if err := cache.Get(ctx, stateCachePrefix+state, &sd); err != nil {
		return ""
	}
	return sd.RedirectURL
}
