package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/eve/esi"
	"context"
	"errors"
	"fmt"
	"strings"
)

const esiMailSendScope = "esi-mail.send_mail.v1"

type inGameMailRecipient struct {
	RecipientID   int64  `json:"recipient_id"`
	RecipientType string `json:"recipient_type"`
}

type inGameMailSendRequest struct {
	Subject    string                `json:"subject"`
	Body       string                `json:"body"`
	Recipients []inGameMailRecipient `json:"recipients"`
}

type inGameMailSupport struct {
	userRepo  *repository.UserRepository
	charRepo  *repository.EveCharacterRepository
	ssoSvc    *EveSSOService
	esiClient *esi.Client
}

func newInGameMailSupport(
	userRepo *repository.UserRepository,
	charRepo *repository.EveCharacterRepository,
	ssoSvc *EveSSOService,
	esiClient *esi.Client,
) inGameMailSupport {
	return inGameMailSupport{
		userRepo:  userRepo,
		charRepo:  charRepo,
		ssoSvc:    ssoSvc,
		esiClient: esiClient,
	}
}

func newConfiguredEveSSOService() *EveSSOService {
	if global.Config == nil {
		return nil
	}
	return NewEveSSOService()
}

func newConfiguredESIClient() *esi.Client {
	if global.Config == nil {
		return nil
	}
	return esi.NewClientWithConfig(global.Config.EveSSO.ESIBaseURL, global.Config.EveSSO.ESIAPIPrefix)
}

func (s inGameMailSupport) resolveUserPrimaryCharacterID(userID uint) (int64, error) {
	if s.userRepo == nil {
		return 0, errors.New("in-game mail user repository unavailable")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0, fmt.Errorf("用户不存在(user_id=%d): %w", userID, err)
	}
	if user.PrimaryCharacterID == 0 {
		return 0, fmt.Errorf("用户主角色未设置(user_id=%d)", userID)
	}
	return user.PrimaryCharacterID, nil
}

func (s inGameMailSupport) resolveSender(ctx context.Context, senderUserID uint) (int64, string, string, error) {
	if s.userRepo == nil || s.charRepo == nil || s.ssoSvc == nil {
		return 0, "", "", errors.New("in-game mail dependencies unavailable")
	}

	user, err := s.userRepo.GetByID(senderUserID)
	if err != nil {
		return 0, "", "", fmt.Errorf("发信用户不存在(user_id=%d): %w", senderUserID, err)
	}
	if user.PrimaryCharacterID == 0 {
		return 0, "", "", fmt.Errorf("发信用户主角色未设置(user_id=%d)", senderUserID)
	}

	senderChar, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
	if err != nil {
		return 0, "", "", fmt.Errorf("发信角色不存在: %w", err)
	}
	if !hasScope(senderChar.Scopes, esiMailSendScope) {
		return 0, "", "", fmt.Errorf("发信角色未授权 scope: %s", esiMailSendScope)
	}

	token, err := s.ssoSvc.GetValidToken(ctx, senderChar.CharacterID)
	if err != nil {
		return 0, "", "", fmt.Errorf("获取发信 token 失败: %w", err)
	}

	displayName := strings.TrimSpace(user.Nickname)
	if displayName == "" {
		displayName = strings.TrimSpace(senderChar.CharacterName)
	}
	if displayName == "" {
		displayName = fmt.Sprintf("Officer %d", senderUserID)
	}

	return senderChar.CharacterID, token, displayName, nil
}

func (s inGameMailSupport) send(
	ctx context.Context,
	senderCharacterID int64,
	accessToken string,
	recipientCharacterID int64,
	subject string,
	body string,
) error {
	if s.esiClient == nil {
		return errors.New("in-game mail client unavailable")
	}

	path := fmt.Sprintf("/characters/%d/mail/", senderCharacterID)
	return s.esiClient.PostNoContent(ctx, path, accessToken, inGameMailSendRequest{
		Subject: subject,
		Body:    body,
		Recipients: []inGameMailRecipient{
			{RecipientID: recipientCharacterID, RecipientType: "character"},
		},
	})
}

func hasScope(scopes, target string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == target {
			return true
		}
	}
	return false
}

func mailErrorDetail(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}
