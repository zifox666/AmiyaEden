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

type MailAttemptSummary struct {
	MailError                  string `json:"mail_error,omitempty"`
	MailID                     int64  `json:"mail_id,omitempty"`
	MailSenderCharacterID      int64  `json:"mail_sender_character_id,omitempty"`
	MailSenderCharacterName    string `json:"mail_sender_character_name,omitempty"`
	MailRecipientCharacterID   int64  `json:"mail_recipient_character_id,omitempty"`
	MailRecipientCharacterName string `json:"mail_recipient_character_name,omitempty"`
}

type resolvedMailSender struct {
	CharacterID   int64
	CharacterName string
	AccessToken   string
	DisplayName   string
}

type resolvedMailRecipient struct {
	CharacterID   int64
	CharacterName string
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

func (s inGameMailSupport) resolveUserPrimaryCharacter(userID uint) (resolvedMailRecipient, error) {
	if s.userRepo == nil {
		return resolvedMailRecipient{}, errors.New("in-game mail user repository unavailable")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return resolvedMailRecipient{}, fmt.Errorf("用户不存在(user_id=%d): %w", userID, err)
	}
	recipient := resolvedMailRecipient{CharacterID: user.PrimaryCharacterID}
	if user.PrimaryCharacterID == 0 {
		return recipient, fmt.Errorf("用户主角色未设置(user_id=%d)", userID)
	}
	if s.charRepo == nil {
		return recipient, errors.New("in-game mail character repository unavailable")
	}

	char, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
	if err != nil {
		return recipient, fmt.Errorf("收信角色不存在: %w", err)
	}
	recipient.CharacterName = strings.TrimSpace(char.CharacterName)
	return recipient, nil
}

func (s inGameMailSupport) resolveSender(ctx context.Context, senderUserID uint) (resolvedMailSender, error) {
	if s.userRepo == nil || s.charRepo == nil || s.ssoSvc == nil {
		return resolvedMailSender{}, errors.New("in-game mail dependencies unavailable")
	}

	user, err := s.userRepo.GetByID(senderUserID)
	if err != nil {
		return resolvedMailSender{}, fmt.Errorf("发信用户不存在(user_id=%d): %w", senderUserID, err)
	}
	sender := resolvedMailSender{CharacterID: user.PrimaryCharacterID}
	if user.PrimaryCharacterID == 0 {
		return sender, fmt.Errorf("发信用户主角色未设置(user_id=%d)", senderUserID)
	}

	senderChar, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
	if err != nil {
		return sender, fmt.Errorf("发信角色不存在: %w", err)
	}
	sender.CharacterName = strings.TrimSpace(senderChar.CharacterName)
	if !hasScope(senderChar.Scopes, esiMailSendScope) {
		return sender, fmt.Errorf("发信角色未授权 scope: %s", esiMailSendScope)
	}

	token, err := s.ssoSvc.GetValidToken(ctx, senderChar.CharacterID)
	if err != nil {
		return sender, fmt.Errorf("获取发信 token 失败: %w", err)
	}
	sender.AccessToken = token

	displayName := strings.TrimSpace(user.Nickname)
	if displayName == "" {
		displayName = strings.TrimSpace(senderChar.CharacterName)
	}
	if displayName == "" {
		displayName = fmt.Sprintf("Officer %d", senderUserID)
	}
	sender.DisplayName = displayName

	return sender, nil
}

func (s inGameMailSupport) send(
	ctx context.Context,
	senderCharacterID int64,
	accessToken string,
	recipientCharacterID int64,
	subject string,
	body string,
) (int64, error) {
	if s.esiClient == nil {
		return 0, errors.New("in-game mail client unavailable")
	}

	path := fmt.Sprintf("/characters/%d/mail/", senderCharacterID)
	var mailID int64
	err := s.esiClient.PostCreatedJSON(ctx, path, accessToken, inGameMailSendRequest{
		Subject: subject,
		Body:    body,
		Recipients: []inGameMailRecipient{
			{RecipientID: recipientCharacterID, RecipientType: "character"},
		},
	}, &mailID)
	return mailID, err
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

func (s MailAttemptSummary) withError(err error) MailAttemptSummary {
	if s.MailError == "" {
		s.MailError = mailErrorDetail(err)
	}
	return s
}
