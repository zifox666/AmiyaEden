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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	seatStateCachePrefix = "seat:sso:state:"
	seatStateCacheTTL    = 10 * time.Minute
)

// seatStateData SeAT OAuth state 中存储的数据
type seatStateData struct {
	RedirectURL  string `json:"redirect_url,omitempty"`
	BindToUserID uint   `json:"bind_to_user_id,omitempty"` // >0 时表示「绑定 SeAT 账号」流程
}

// SeatSSOService SeAT SSO 业务逻辑层
type SeatSSOService struct {
	seatUserRepo *repository.SeatUserRepository
	charRepo     *repository.EveCharacterRepository
	userRepo     *repository.UserRepository
	configRepo   *repository.SysConfigRepository
	roleSvc      *RoleService
}

func NewSeatSSOService() *SeatSSOService {
	return &SeatSSOService{
		seatUserRepo: repository.NewSeatUserRepository(),
		charRepo:     repository.NewEveCharacterRepository(),
		userRepo:     repository.NewUserRepository(),
		configRepo:   repository.NewSysConfigRepository(),
		roleSvc:      NewRoleService(),
	}
}

// buildSeatClient 根据 sys_config 动态构建 SeAT OAuth 客户端
func (s *SeatSSOService) buildSeatClient() (*eve.SeatClient, error) {
	enabled := s.configRepo.GetBool(model.SysConfigSeatEnabled, false)
	if !enabled {
		return nil, errors.New("SeAT 登录未启用")
	}

	baseURL, _ := s.configRepo.Get(model.SysConfigSeatBaseURL, "")
	clientID, _ := s.configRepo.Get(model.SysConfigSeatClientID, "")
	clientSecret, _ := s.configRepo.Get(model.SysConfigSeatClientSecret, "")
	callbackURL, _ := s.configRepo.Get(model.SysConfigSeatCallbackURL, "")

	if baseURL == "" || clientID == "" || clientSecret == "" || callbackURL == "" {
		return nil, errors.New("SeAT OAuth 配置不完整")
	}

	return eve.NewSeatClient(baseURL, clientID, clientSecret, callbackURL), nil
}

// getSeatScopes 获取 SeAT OAuth scopes
func (s *SeatSSOService) getSeatScopes() []string {
	scopesStr, _ := s.configRepo.Get(model.SysConfigSeatScopes, model.SysConfigDefaultSeatScopes)
	var scopes []string
	for _, sc := range strings.Fields(scopesStr) {
		if sc != "" {
			scopes = append(scopes, sc)
		}
	}
	if len(scopes) == 0 {
		scopes = strings.Fields(model.SysConfigDefaultSeatScopes)
	}
	return scopes
}

// IsSeatEnabled 检查 SeAT 登录是否已启用
func (s *SeatSSOService) IsSeatEnabled() bool {
	return s.configRepo.GetBool(model.SysConfigSeatEnabled, false)
}

// GetSeatAuthURL 生成 SeAT 授权 URL（登录流程）
func (s *SeatSSOService) GetSeatAuthURL(ctx context.Context, redirectURL string) (string, error) {
	client, err := s.buildSeatClient()
	if err != nil {
		return "", err
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	prefix, _ := s.configRepo.Get(model.SysConfigCorpID, "")
	state = prefix + "_" + state

	data := seatStateData{RedirectURL: redirectURL}
	if err := cache.Set(ctx, seatStateCachePrefix+state, data, seatStateCacheTTL); err != nil {
		global.Logger.Warn("存储 SeAT state 失败", zap.Error(err))
	}

	scopes := s.getSeatScopes()
	return client.BuildAuthURL(state, scopes), nil
}

// GetSeatBindURL 生成 SeAT 授权 URL（绑定流程，已有本系统用户）
func (s *SeatSSOService) GetSeatBindURL(ctx context.Context, userID uint, redirectURL string) (string, error) {
	client, err := s.buildSeatClient()
	if err != nil {
		return "", err
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	prefix, _ := s.configRepo.Get(model.SysConfigCorpID, "")
	state = prefix + "_" + state

	data := seatStateData{RedirectURL: redirectURL, BindToUserID: userID}
	if err := cache.Set(ctx, seatStateCachePrefix+state, data, seatStateCacheTTL); err != nil {
		global.Logger.Warn("存储 SeAT state 失败", zap.Error(err))
	}

	scopes := s.getSeatScopes()
	return client.BuildAuthURL(state, scopes), nil
}

// GetRedirectURLFromState 从 state 中恢复 redirect URL
func (s *SeatSSOService) GetRedirectURLFromState(ctx context.Context, state string) string {
	if state == "" {
		return ""
	}
	var sd seatStateData
	if err := cache.Get(ctx, seatStateCachePrefix+state, &sd); err != nil {
		return ""
	}
	return sd.RedirectURL
}

// SeatCallbackResult SeAT OAuth 回调处理结果
type SeatCallbackResult struct {
	Token           string      `json:"token"`
	User            *model.User `json:"user"`
	RedirectURL     string      `json:"redirect_url"`
	IsRawRedirect   bool        `json:"-"`
	PendingTransfer bool        `json:"-"`
}

// HandleSeatCallback 处理 SeAT OAuth 回调
func (s *SeatSSOService) HandleSeatCallback(ctx context.Context, code, state, clientIP string) (*SeatCallbackResult, error) {
	if code == "" {
		return nil, errors.New("授权码不能为空")
	}

	client, err := s.buildSeatClient()
	if err != nil {
		return nil, err
	}

	// 读取并删除 state
	var sd seatStateData
	if err := cache.Get(ctx, seatStateCachePrefix+state, &sd); err != nil {
		global.Logger.Warn("SeAT state 未找到或已过期", zap.String("state", state))
	} else {
		_ = cache.Del(ctx, seatStateCachePrefix+state)
	}

	// 1. 交换 Token
	tokenResp, err := client.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 2. 从 id_token JWT 中直接解析用户信息（无需额外网络请求）
	if tokenResp.IDToken == "" {
		return nil, fmt.Errorf("SeAT token response missing id_token")
	}
	userInfo, err := eve.ParseIDToken(tokenResp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("parse SeAT id_token: %w", err)
	}

	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	groupsJSON, _ := json.Marshal(userInfo.Groups)

	// 3. 查找已有的 SeAT 绑定
	existingSeat, seatErr := s.seatUserRepo.GetBySeatUserID(userInfo.Sub)
	now := time.Now()

	// ── 绑定流程 ──
	if sd.BindToUserID > 0 {
		return s.handleSeatBind(ctx, sd, userInfo, tokenResp, tokenExpiry, string(groupsJSON))
	}

	// ── 登录流程 ──
	if seatErr != nil && errors.Is(seatErr, gorm.ErrRecordNotFound) {
		// SeAT 用户首次出现，自动合并角色
		return s.handleSeatFirstLogin(ctx, sd, userInfo, tokenResp, tokenExpiry, string(groupsJSON), clientIP, now)
	}
	if seatErr != nil {
		return nil, seatErr
	}

	// SeAT 用户已有绑定，更新 Token 并登录
	existingSeat.AccessToken = tokenResp.AccessToken
	existingSeat.RefreshToken = tokenResp.RefreshToken
	existingSeat.TokenExpiry = tokenExpiry
	existingSeat.SeatUsername = userInfo.Name
	existingSeat.Groups = string(groupsJSON)
	if err := s.seatUserRepo.Update(existingSeat); err != nil {
		return nil, err
	}

	// 同步 SeAT 角色列表到本系统
	s.syncSeatCharacters(existingSeat.UserID, userInfo)

	user, err := s.userRepo.GetByID(existingSeat.UserID)
	if err != nil {
		return nil, err
	}

	user.LastLoginAt = &now
	user.LastLoginIP = clientIP
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	jwtToken, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return nil, err
	}
	return &SeatCallbackResult{Token: jwtToken, User: user, RedirectURL: sd.RedirectURL}, nil
}

// handleSeatFirstLogin 处理 SeAT 用户首次登录
func (s *SeatSSOService) handleSeatFirstLogin(
	ctx context.Context, sd seatStateData,
	userInfo *eve.SeatUserInfo, tokenResp *eve.SeatTokenResponse,
	tokenExpiry time.Time, groupsJSON string,
	clientIP string, now time.Time,
) (*SeatCallbackResult, error) {

	// 尝试通过 SeAT 角色列表找到已有的本系统用户
	var targetUser *model.User
	for _, acct := range userInfo.Accounts {
		if !acct.Valid {
			continue
		}
		char, err := s.charRepo.GetByCharacterID(acct.ID)
		if err != nil {
			continue // 该角色不在本系统中
		}
		user, err := s.userRepo.GetByID(char.UserID)
		if err != nil {
			continue
		}
		targetUser = user
		break
	}

	if targetUser == nil {
		// 没有任何匹配角色，创建新用户
		primaryCharID := int64(0)
		nickname := userInfo.Name
		portraitURL := ""

		// 使用 SeAT 主角色信息
		for _, acct := range userInfo.Accounts {
			if acct.Valid {
				primaryCharID = acct.ID
				nickname = acct.Name
				portraitURL = eve.PortraitURL(acct.ID)
				break
			}
		}

		targetUser = &model.User{
			Nickname:           nickname,
			Avatar:             portraitURL,
			Status:             1,
			Role:               "user",
			PrimaryCharacterID: primaryCharID,
			LastLoginAt:        &now,
			LastLoginIP:        clientIP,
		}
		if err := s.userRepo.Create(targetUser); err != nil {
			return nil, err
		}
		s.roleSvc.EnsureUserHasDefaultRole(context.Background(), targetUser.ID)
	} else {
		targetUser.LastLoginAt = &now
		targetUser.LastLoginIP = clientIP
		if err := s.userRepo.Update(targetUser); err != nil {
			return nil, err
		}
	}

	// 创建 SeAT 绑定记录
	mainCharID := int64(0)
	if userInfo.UID != "" {
		var id int64
		if _, err := parseCharacterID(userInfo.UID); err == nil {
			id, _ = parseCharacterID(userInfo.UID)
			mainCharID = id
		}
	}

	seatUser := &model.SeatUser{
		SeatUserID:   userInfo.Sub,
		SeatUsername: userInfo.Name,
		UserID:       targetUser.ID,
		MainCharID:   mainCharID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  tokenExpiry,
		Groups:       groupsJSON,
	}
	if err := s.seatUserRepo.Create(seatUser); err != nil {
		return nil, err
	}

	// 同步角色列表（自动合并，冲突提示迁移）
	s.syncSeatCharacters(targetUser.ID, userInfo)

	jwtToken, err := jwt.GenerateToken(targetUser.ID, targetUser.PrimaryCharacterID, targetUser.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return nil, err
	}
	return &SeatCallbackResult{Token: jwtToken, User: targetUser, RedirectURL: sd.RedirectURL}, nil
}

// handleSeatBind 处理 SeAT 账号绑定到已有用户
func (s *SeatSSOService) handleSeatBind(
	ctx context.Context, sd seatStateData,
	userInfo *eve.SeatUserInfo, tokenResp *eve.SeatTokenResponse,
	tokenExpiry time.Time, groupsJSON string,
) (*SeatCallbackResult, error) {

	// 检查该 SeAT 账号是否已被其他用户绑定
	existing, err := s.seatUserRepo.GetBySeatUserID(userInfo.Sub)
	if err == nil && existing.UserID != sd.BindToUserID {
		return nil, errors.New("该 SeAT 账号已被其他用户绑定")
	}

	mainCharID := int64(0)
	if userInfo.UID != "" {
		if id, err := parseCharacterID(userInfo.UID); err == nil {
			mainCharID = id
		}
	}

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新绑定
		seatUser := &model.SeatUser{
			SeatUserID:   userInfo.Sub,
			SeatUsername: userInfo.Name,
			UserID:       sd.BindToUserID,
			MainCharID:   mainCharID,
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			TokenExpiry:  tokenExpiry,
			Groups:       groupsJSON,
		}
		if err := s.seatUserRepo.Create(seatUser); err != nil {
			return nil, err
		}
	} else if err == nil {
		// 更新已有绑定
		existing.AccessToken = tokenResp.AccessToken
		existing.RefreshToken = tokenResp.RefreshToken
		existing.TokenExpiry = tokenExpiry
		existing.SeatUsername = userInfo.Name
		existing.Groups = groupsJSON
		if err := s.seatUserRepo.Update(existing); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 同步角色
	s.syncSeatCharacters(sd.BindToUserID, userInfo)

	user, err := s.userRepo.GetByID(sd.BindToUserID)
	if err != nil {
		return nil, err
	}

	jwtToken, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return nil, err
	}
	return &SeatCallbackResult{Token: jwtToken, User: user, RedirectURL: sd.RedirectURL}, nil
}

// syncSeatCharacters 将 SeAT 角色列表同步到本系统
// 对于不在本系统中的角色，创建无 ESI Token 的 EveCharacter 记录
// 对于属于其他用户的角色，记录日志（需要用户手动迁移）
func (s *SeatSSOService) syncSeatCharacters(userID uint, userInfo *eve.SeatUserInfo) {
	for _, acct := range userInfo.Accounts {
		if !acct.Valid {
			continue
		}

		char, err := s.charRepo.GetByCharacterID(acct.ID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				global.Logger.Warn("SeAT 角色同步查询失败",
					zap.Int64("characterID", acct.ID),
					zap.Error(err))
				continue
			}

			// 角色不存在，创建仅有基本信息的记录（无 ESI Token）
			newChar := &model.EveCharacter{
				CharacterID:   acct.ID,
				CharacterName: acct.Name,
				PortraitURL:   eve.PortraitURL(acct.ID),
				UserID:        userID,
			}
			if err := s.charRepo.Create(newChar); err != nil {
				global.Logger.Warn("SeAT 角色同步创建失败",
					zap.Int64("characterID", acct.ID),
					zap.Error(err))
			} else {
				global.Logger.Info("SeAT 角色同步：新角色已绑定",
					zap.Int64("characterID", acct.ID),
					zap.String("name", acct.Name),
					zap.Uint("userID", userID))

				// 触发 affiliation 同步和权限检查（异步，避免多角色时阻塞回调重定向）
				if OnNewCharacterSyncFunc != nil {
					go OnNewCharacterSyncFunc(acct.ID, userID)
				}
				if OnNewCharacterFunc != nil {
					go OnNewCharacterFunc(acct.ID, userID)
				}
			}
			continue
		}

		// 角色已存在
		if char.UserID == userID {
			// 已属于当前用户，更新名字
			if char.CharacterName != acct.Name {
				char.CharacterName = acct.Name
				_ = s.charRepo.Update(char)
			}
			continue
		}

		// 角色属于其他用户，检查原用户是否已被删除
		_, userErr := s.userRepo.GetByID(char.UserID)
		if userErr != nil && errors.Is(userErr, gorm.ErrRecordNotFound) {
			// 原用户已删除，重新绑定
			char.UserID = userID
			char.CharacterName = acct.Name
			char.PortraitURL = eve.PortraitURL(acct.ID)
			if err := s.charRepo.Update(char); err != nil {
				global.Logger.Warn("SeAT 角色同步：孤儿角色重绑失败",
					zap.Int64("characterID", acct.ID),
					zap.Error(err))
			} else {
				global.Logger.Info("SeAT 角色同步：孤儿角色已重绑",
					zap.Int64("characterID", acct.ID),
					zap.Uint("oldUserID", char.UserID),
					zap.Uint("newUserID", userID))
			}
			continue
		}

		// 角色属于其他活跃用户，记录警告
		global.Logger.Warn("SeAT 角色同步：角色已绑定到其他用户，跳过",
			zap.Int64("characterID", acct.ID),
			zap.String("name", acct.Name),
			zap.Uint("ownerUserID", char.UserID),
			zap.Uint("seatUserID", userID))
	}

	// 确保用户有主角色
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return
	}
	if user.PrimaryCharacterID == 0 {
		for _, acct := range userInfo.Accounts {
			if acct.Valid {
				user.PrimaryCharacterID = acct.ID
				user.Avatar = eve.PortraitURL(acct.ID)
				user.Nickname = acct.Name
				_ = s.userRepo.Update(user)
				break
			}
		}
	}
}

// GetSeatUserByUserID 获取用户的 SeAT 绑定信息
func (s *SeatSSOService) GetSeatUserByUserID(userID uint) (*model.SeatUser, error) {
	return s.seatUserRepo.GetByUserID(userID)
}

// UnbindSeat 解除 SeAT 绑定
func (s *SeatSSOService) UnbindSeat(userID uint) error {
	su, err := s.seatUserRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("未绑定 SeAT 账号")
	}
	if su.UserID != userID {
		return errors.New("无权操作")
	}
	return s.seatUserRepo.Delete(su.ID)
}

// GetSeatConfig 获取 SeAT 配置（公开部分，不含 secret）
func (s *SeatSSOService) GetSeatConfig() map[string]interface{} {
	enabled := s.configRepo.GetBool(model.SysConfigSeatEnabled, false)
	baseURL, _ := s.configRepo.Get(model.SysConfigSeatBaseURL, "")
	clientID, _ := s.configRepo.Get(model.SysConfigSeatClientID, "")
	callbackURL, _ := s.configRepo.Get(model.SysConfigSeatCallbackURL, "")
	scopes, _ := s.configRepo.Get(model.SysConfigSeatScopes, model.SysConfigDefaultSeatScopes)

	return map[string]interface{}{
		"enabled":      enabled,
		"base_url":     baseURL,
		"client_id":    clientID,
		"callback_url": callbackURL,
		"scopes":       scopes,
	}
}

// GetSeatAdminConfig 获取 SeAT 完整配置（含 secret，管理员用）
func (s *SeatSSOService) GetSeatAdminConfig() map[string]interface{} {
	cfg := s.GetSeatConfig()
	secret, _ := s.configRepo.Get(model.SysConfigSeatClientSecret, "")
	cfg["client_secret"] = secret
	return cfg
}

// UpdateSeatConfig 更新 SeAT 配置
func (s *SeatSSOService) UpdateSeatConfig(updates map[string]string) error {
	keyMap := map[string]string{
		"enabled":       model.SysConfigSeatEnabled,
		"base_url":      model.SysConfigSeatBaseURL,
		"client_id":     model.SysConfigSeatClientID,
		"client_secret": model.SysConfigSeatClientSecret,
		"callback_url":  model.SysConfigSeatCallbackURL,
		"scopes":        model.SysConfigSeatScopes,
	}

	descMap := map[string]string{
		"enabled":       "SeAT 登录开关",
		"base_url":      "SeAT 基础 URL",
		"client_id":     "SeAT OAuth Client ID",
		"client_secret": "SeAT OAuth Client Secret",
		"callback_url":  "SeAT OAuth 回调 URL",
		"scopes":        "SeAT OAuth Scopes",
	}

	for k, v := range updates {
		sysKey, ok := keyMap[k]
		if !ok {
			continue
		}
		if err := s.configRepo.Set(sysKey, v, descMap[k]); err != nil {
			return err
		}
	}
	return nil
}

// GetESITokenForCharacter 为 SeAT-only 角色通过 passthrough 端点获取 ESI access_token
// 当角色没有 EVE SSO token 时（RefreshToken 为空），由 EveSSOService.GetValidToken 调用
func (s *SeatSSOService) GetESITokenForCharacter(ctx context.Context, characterID int64, userID uint) (string, error) {
	// 1. 找 SeAT 绑定
	seatUser, err := s.seatUserRepo.GetByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("用户 %d 未绑定 SeAT 账号", userID)
	}
	if seatUser.RefreshToken == "" {
		return "", fmt.Errorf("SeAT 绑定缺少 refresh_token，请重新登录 SeAT")
	}

	// 2. 构建 SeAT 客户端
	client, err := s.buildSeatClient()
	if err != nil {
		return "", fmt.Errorf("SeAT 未启用或配置不完整: %w", err)
	}

	// 3. 刷新 SeAT token（剩余不足 3 分钟时）
	seatAccessToken := seatUser.AccessToken
	if time.Until(seatUser.TokenExpiry) < 3*time.Minute {
		global.Logger.Info("[SeAT Passthrough] 刷新 SeAT token",
			zap.Uint("userID", userID),
			zap.Int64("characterID", characterID),
		)
		refreshed, refreshErr := client.RefreshAccessToken(ctx, seatUser.RefreshToken)
		if refreshErr != nil {
			return "", fmt.Errorf("刷新 SeAT token 失败: %w", refreshErr)
		}
		seatAccessToken = refreshed.AccessToken
		seatUser.AccessToken = refreshed.AccessToken
		seatUser.RefreshToken = refreshed.RefreshToken
		seatUser.TokenExpiry = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second)
		if saveErr := s.seatUserRepo.Update(seatUser); saveErr != nil {
			global.Logger.Warn("[SeAT Passthrough] 保存刷新后的 SeAT token 失败",
				zap.Uint("userID", userID),
				zap.Error(saveErr),
			)
		}
	}

	// 4. 调用 passthrough 端点换取 ESI access_token
	tokenResp, err := client.GetPassthroughToken(ctx, characterID, seatAccessToken, nil)
	if err != nil {
		return "", fmt.Errorf("SeAT passthrough 角色 %d 失败: %w", characterID, err)
	}

	// 5. 解析 passthrough token 中的 scopes，写回角色记录以供 ESI 队列调度
	if claims, parseErr := eve.ParsePassthroughToken(tokenResp.AccessToken); parseErr == nil && len(claims.Scp) > 0 {
		if char, charErr := s.charRepo.GetByCharacterID(characterID); charErr == nil {
			newScopes := strings.Join(claims.Scp, " ")
			if char.Scopes != newScopes {
				char.Scopes = newScopes
				if updateErr := s.charRepo.Update(char); updateErr != nil {
					global.Logger.Warn("[SeAT Passthrough] 更新角色 scopes 失败",
						zap.Int64("characterID", characterID),
						zap.Error(updateErr),
					)
				}
			}
		}
	}

	return tokenResp.AccessToken, nil
}

// RefreshAllSeatUserGroups 刷新所有 SeAT 用户的 token，解析 id_token 更新 groups
// 供定时任务调用
func (s *SeatSSOService) RefreshAllSeatUserGroups(ctx context.Context) {
	client, err := s.buildSeatClient()
	if err != nil {
		global.Logger.Warn("[SeAT Sync] SeAT 未启用，跳过分组同步", zap.Error(err))
		return
	}

	seatUsers, err := s.seatUserRepo.ListAll()
	if err != nil {
		global.Logger.Error("[SeAT Sync] 查询 SeAT 用户列表失败", zap.Error(err))
		return
	}
	if len(seatUsers) == 0 {
		return
	}

	global.Logger.Info("[SeAT Sync] 开始刷新 SeAT 用户分组", zap.Int("count", len(seatUsers)))

	updated := 0
	for i := range seatUsers {
		su := &seatUsers[i]
		if su.RefreshToken == "" {
			continue
		}

		// 刷新 SeAT token
		refreshed, refreshErr := client.RefreshAccessToken(ctx, su.RefreshToken)
		if refreshErr != nil {
			global.Logger.Warn("[SeAT Sync] 刷新 SeAT token 失败",
				zap.Uint("userID", su.UserID),
				zap.String("seatUser", su.SeatUsername),
				zap.Error(refreshErr))
			continue
		}

		su.AccessToken = refreshed.AccessToken
		su.RefreshToken = refreshed.RefreshToken
		su.TokenExpiry = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second)

		// 如果 refresh 返回了 id_token，直接解析更新 groups
		if refreshed.IDToken != "" {
			if userInfo, parseErr := eve.ParseIDToken(refreshed.IDToken); parseErr == nil {
				groupsJSON, _ := json.Marshal(userInfo.Groups)
				su.Groups = string(groupsJSON)

				// 更新主角色 ID
				if userInfo.UID != "" {
					if mainID, idErr := parseCharacterID(userInfo.UID); idErr == nil {
						su.MainCharID = mainID
					}
				}
			}
		} else {
			// refresh 没有 id_token，通过 userinfo 端点获取 groups
			userInfo, uiErr := client.GetUserInfo(ctx, refreshed.AccessToken)
			if uiErr != nil {
				global.Logger.Warn("[SeAT Sync] 获取 userinfo 失败",
					zap.Uint("userID", su.UserID),
					zap.Error(uiErr))
			} else {
				groupsJSON, _ := json.Marshal(userInfo.Groups)
				su.Groups = string(groupsJSON)

				if userInfo.UID != "" {
					if mainID, idErr := parseCharacterID(userInfo.UID); idErr == nil {
						su.MainCharID = mainID
					}
				}
			}
		}

		if saveErr := s.seatUserRepo.Update(su); saveErr != nil {
			global.Logger.Warn("[SeAT Sync] 保存 SeAT 用户失败",
				zap.Uint("userID", su.UserID),
				zap.Error(saveErr))
			continue
		}
		updated++
	}

	global.Logger.Info("[SeAT Sync] SeAT 用户分组刷新完成",
		zap.Int("total", len(seatUsers)),
		zap.Int("updated", updated))
}

// parseCharacterID 解析 character ID 字符串为 int64
func parseCharacterID(s string) (int64, error) {
	var id int64
	_, err := parseIntFromString(s, &id)
	return id, err
}

func parseIntFromString(s string, out *int64) (int, error) {
	return fmt.Sscanf(s, "%d", out)
}
