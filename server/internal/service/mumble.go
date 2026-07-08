package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
)

const mumblePasswordLength = 16
const voiceProviderMumble = "mumble"

type MumbleService struct {
	repo       *repository.MumbleRepository
	roleRepo   *repository.RoleRepository
	userRepo   *repository.UserRepository
	configRepo *repository.SysConfigRepository
}

func NewMumbleService() *MumbleService {
	return &MumbleService{
		repo:       repository.NewMumbleRepository(),
		roleRepo:   repository.NewRoleRepository(),
		userRepo:   repository.NewUserRepository(),
		configRepo: repository.NewSysConfigRepository(),
	}
}

type MumbleConfigDTO struct {
	Enabled       bool   `json:"enabled"`
	URL           string `json:"url"`
	Port          int    `json:"port"`
	ServerName    string `json:"server_name"`
	AuthSecretSet bool   `json:"auth_secret_set"`
}

type UpdateMumbleConfigRequest struct {
	Enabled    *bool   `json:"enabled"`
	URL        *string `json:"url"`
	Port       *int    `json:"port"`
	ServerName *string `json:"server_name"`
	AuthSecret *string `json:"auth_secret"`
}

type MumbleAccountDTO struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
	QuickURL    string   `json:"quick_url"`
}

type MumbleProfileDTO struct {
	Config  MumbleConfigDTO  `json:"config"`
	Account MumbleAccountDTO `json:"account"`
}

type MumbleICEAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MumbleICEAuthResponse struct {
	Allowed     bool     `json:"allowed"`
	UserID      uint     `json:"user_id"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
}

type MumbleRoleGroupDTO struct {
	RoleID    uint   `json:"role_id"`
	RoleCode  string `json:"role_code"`
	RoleName  string `json:"role_name"`
	GroupName string `json:"group_name"`
	Enabled   bool   `json:"enabled"`
}

type UpdateMumbleRoleGroupsRequest struct {
	Mappings []UpdateMumbleRoleGroupItem `json:"mappings"`
}

type UpdateMumbleRoleGroupItem struct {
	RoleCode  string `json:"role_code"`
	GroupName string `json:"group_name"`
	Enabled   bool   `json:"enabled"`
}

func (s *MumbleService) GetProfile(userID uint) (*MumbleProfileDTO, error) {
	account, err := s.ensureAccount(userID, false)
	if err != nil {
		return nil, err
	}
	config := s.GetConfig()
	groups, err := s.ResolveUserGroups(userID)
	if err != nil {
		return nil, err
	}
	return &MumbleProfileDTO{
		Config:  config,
		Account: s.toAccountDTO(account, config, groups),
	}, nil
}

func (s *MumbleService) ResetPassword(userID uint) (*MumbleAccountDTO, error) {
	account, err := s.ensureAccount(userID, true)
	if err != nil {
		return nil, err
	}
	config := s.GetConfig()
	groups, err := s.ResolveUserGroups(userID)
	if err != nil {
		return nil, err
	}
	dto := s.toAccountDTO(account, config, groups)
	return &dto, nil
}

func (s *MumbleService) AuthenticateICE(req MumbleICEAuthRequest) (*MumbleICEAuthResponse, error) {
	userID64, err := strconv.ParseUint(strings.TrimSpace(req.Username), 10, 64)
	if err != nil || userID64 == 0 {
		return &MumbleICEAuthResponse{Allowed: false}, nil
	}
	userID := uint(userID64)

	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		if s.repo.IsNotFound(err) {
			return &MumbleICEAuthResponse{Allowed: false}, nil
		}
		return nil, err
	}
	if account.Password != req.Password {
		return &MumbleICEAuthResponse{Allowed: false}, nil
	}

	groups, err := s.ResolveUserGroups(userID)
	if err != nil {
		return nil, err
	}
	return &MumbleICEAuthResponse{
		Allowed:     true,
		UserID:      account.UserID,
		DisplayName: account.DisplayName,
		Groups:      groups,
	}, nil
}

func (s *MumbleService) CheckICEAuthSecret(secret string) bool {
	expected, _ := s.configRepo.Get(model.SysConfigMumbleAuthSecret, model.SysConfigDefaultMumbleAuthSecret)
	return expected != "" && secret != "" && subtle.ConstantTimeCompare([]byte(secret), []byte(expected)) == 1
}

func (s *MumbleService) ListRoleGroups() ([]MumbleRoleGroupDTO, error) {
	roles, err := s.roleRepo.ListAll()
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaultRoleGroups(roles); err != nil {
		return nil, err
	}
	mappings, err := s.repo.ListRoleGroupMappings(voiceProviderMumble)
	if err != nil {
		return nil, err
	}
	mappingByRole := make(map[string]model.VoiceRoleGroupMapping, len(mappings))
	for _, mapping := range mappings {
		mappingByRole[mapping.RoleCode] = mapping
	}

	result := make([]MumbleRoleGroupDTO, 0, len(roles))
	for _, role := range roles {
		mapping := mappingByRole[role.Code]
		result = append(result, MumbleRoleGroupDTO{
			RoleID:    role.ID,
			RoleCode:  role.Code,
			RoleName:  role.Name,
			GroupName: mapping.GroupName,
			Enabled:   mapping.Enabled,
		})
	}
	return result, nil
}

func (s *MumbleService) UpdateRoleGroups(req UpdateMumbleRoleGroupsRequest) error {
	roles, err := s.roleRepo.ListAll()
	if err != nil {
		return err
	}
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		roleSet[role.Code] = struct{}{}
	}

	mappings := make([]model.VoiceRoleGroupMapping, 0, len(req.Mappings))
	for _, item := range req.Mappings {
		roleCode := strings.TrimSpace(item.RoleCode)
		if _, ok := roleSet[roleCode]; !ok {
			return fmt.Errorf("角色不存在: %s", roleCode)
		}
		groupName := strings.TrimSpace(item.GroupName)
		if item.Enabled && groupName == "" {
			return fmt.Errorf("角色 %s 的 Mumble group 不能为空", roleCode)
		}
		mappings = append(mappings, model.VoiceRoleGroupMapping{
			Provider:  voiceProviderMumble,
			RoleCode:  roleCode,
			GroupName: groupName,
			Enabled:   item.Enabled,
		})
	}
	return s.repo.UpsertRoleGroupMappings(mappings)
}

func (s *MumbleService) ResolveUserGroups(userID uint) ([]string, error) {
	roleCodes, err := s.roleRepo.GetUserRoleCodes(userID)
	if err != nil {
		return nil, err
	}
	roles, err := s.roleRepo.ListAll()
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaultRoleGroups(roles); err != nil {
		return nil, err
	}
	mappings, err := s.repo.ListRoleGroupMappings(voiceProviderMumble)
	if err != nil {
		return nil, err
	}
	mappingByRole := make(map[string]model.VoiceRoleGroupMapping, len(mappings))
	for _, mapping := range mappings {
		mappingByRole[mapping.RoleCode] = mapping
	}

	groups := make([]string, 0, len(roleCodes))
	seen := make(map[string]struct{}, len(roleCodes))
	for _, roleCode := range roleCodes {
		mapping, ok := mappingByRole[roleCode]
		if !ok || !mapping.Enabled || mapping.GroupName == "" {
			continue
		}
		if _, exists := seen[mapping.GroupName]; exists {
			continue
		}
		seen[mapping.GroupName] = struct{}{}
		groups = append(groups, mapping.GroupName)
	}
	return groups, nil
}

func (s *MumbleService) GetConfig() MumbleConfigDTO {
	portStr, _ := s.configRepo.Get(model.SysConfigMumblePort, model.SysConfigDefaultMumblePort)
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		port = 64738
	}

	urlValue, _ := s.configRepo.Get(model.SysConfigMumbleURL, model.SysConfigDefaultMumbleURL)
	serverName, _ := s.configRepo.Get(model.SysConfigMumbleServerName, model.SysConfigDefaultMumbleServerName)
	authSecret, _ := s.configRepo.Get(model.SysConfigMumbleAuthSecret, model.SysConfigDefaultMumbleAuthSecret)

	return MumbleConfigDTO{
		Enabled:       s.configRepo.GetBool(model.SysConfigMumbleEnabled, false),
		URL:           urlValue,
		Port:          port,
		ServerName:    serverName,
		AuthSecretSet: authSecret != "",
	}
}

func (s *MumbleService) UpdateConfig(req UpdateMumbleConfigRequest) error {
	if req.Port != nil && (*req.Port < 1 || *req.Port > 65535) {
		return errors.New("端口必须在 1 到 65535 之间")
	}

	if req.Enabled != nil {
		enabled := "false"
		if *req.Enabled {
			enabled = "true"
		}
		if err := s.configRepo.Set(model.SysConfigMumbleEnabled, enabled, "Mumble 是否启用"); err != nil {
			return err
		}
	}
	if req.URL != nil {
		normalizedURL := strings.TrimSpace(*req.URL)
		if err := s.configRepo.Set(model.SysConfigMumbleURL, normalizedURL, "Mumble 服务器地址"); err != nil {
			return err
		}
	}
	if req.Port != nil {
		if err := s.configRepo.Set(model.SysConfigMumblePort, strconv.Itoa(*req.Port), "Mumble 服务器端口"); err != nil {
			return err
		}
	}
	if req.ServerName != nil {
		serverName := strings.TrimSpace(*req.ServerName)
		if serverName == "" {
			serverName = model.SysConfigDefaultMumbleServerName
		}
		if err := s.configRepo.Set(model.SysConfigMumbleServerName, serverName, "Mumble 服务器名称"); err != nil {
			return err
		}
	}
	if req.AuthSecret != nil {
		authSecret := strings.TrimSpace(*req.AuthSecret)
		if err := s.configRepo.Set(model.SysConfigMumbleAuthSecret, authSecret, "Mumble ICE Authenticator 共享密钥"); err != nil {
			return err
		}
	}
	return nil
}

func (s *MumbleService) ensureAccount(userID uint, forceReset bool) (*model.MumbleAccount, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	displayName := strings.TrimSpace(user.Nickname)
	if displayName == "" {
		displayName = fmt.Sprintf("Capsuleer#%d", userID)
	}

	account, err := s.repo.GetByUserID(userID)
	if err == nil {
		if !forceReset {
			if account.DisplayName != displayName {
				_ = s.repo.UpdatePasswordAndDisplayName(userID, account.Password, displayName)
				account.DisplayName = displayName
			}
			return account, nil
		}
		password, err := generateMumblePassword()
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpdatePasswordAndDisplayName(userID, password, displayName); err != nil {
			return nil, err
		}
		account.Password = password
		account.DisplayName = displayName
		return account, nil
	}
	if !s.repo.IsNotFound(err) {
		return nil, err
	}

	password, err := generateMumblePassword()
	if err != nil {
		return nil, err
	}
	account = &model.MumbleAccount{
		UserID:      userID,
		Password:    password,
		DisplayName: displayName,
	}
	if err := s.repo.Create(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *MumbleService) toAccountDTO(account *model.MumbleAccount, config MumbleConfigDTO, groups []string) MumbleAccountDTO {
	username := strconv.FormatUint(uint64(account.UserID), 10)
	return MumbleAccountDTO{
		UserID:      account.UserID,
		Username:    username,
		Password:    account.Password,
		DisplayName: account.DisplayName,
		Groups:      groups,
		QuickURL:    buildMumbleURL(username, account.Password, config),
	}
}

func (s *MumbleService) ensureDefaultRoleGroups(roles []model.Role) error {
	existing, err := s.repo.ListRoleGroupMappings(voiceProviderMumble)
	if err != nil {
		return err
	}
	existingByRole := make(map[string]struct{}, len(existing))
	for _, mapping := range existing {
		existingByRole[mapping.RoleCode] = struct{}{}
	}

	defaults := make([]model.VoiceRoleGroupMapping, 0)
	for _, role := range roles {
		if _, ok := existingByRole[role.Code]; ok {
			continue
		}
		defaults = append(defaults, model.VoiceRoleGroupMapping{
			Provider:  voiceProviderMumble,
			RoleCode:  role.Code,
			GroupName: defaultMumbleGroupName(role.Code),
			Enabled:   role.Code != model.RoleGuest,
		})
	}
	return s.repo.UpsertRoleGroupMappings(defaults)
}

func defaultMumbleGroupName(roleCode string) string {
	return "amiya_" + strings.ReplaceAll(strings.ToLower(roleCode), "-", "_")
}

func buildMumbleURL(username, password string, config MumbleConfigDTO) string {
	host := strings.TrimSpace(config.URL)
	if host == "" {
		return ""
	}
	host = strings.TrimPrefix(host, "mumble://")
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimRight(host, "/")
	title := url.QueryEscape(config.ServerName)
	return fmt.Sprintf("mumble://%s:%s@%s:%d/?version=1.2.0&title=%s", username, password, host, config.Port, title)
}

func generateMumblePassword() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	result := make([]byte, mumblePasswordLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}
