package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type RecruitmentService struct {
	repo      *repository.RecruitmentRepository
	userSvc   *UserService
	charRepo  *repository.EveCharacterRepository
	skillRepo *repository.EveSkillRepository
	roleSvc   *RoleService
}

func NewRecruitmentService() *RecruitmentService {
	return &RecruitmentService{
		repo: repository.NewRecruitmentRepository(), userSvc: NewUserService(),
		charRepo: repository.NewEveCharacterRepository(), skillRepo: repository.NewEveSkillRepository(), roleSvc: NewRoleService(),
	}
}

type RecruitmentQuestionnaireInput struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Fields      []model.RecruitmentField `json:"fields"`
}

type RecruitmentQuestionnaireDTO struct {
	Questionnaire model.RecruitmentQuestionnaire `json:"questionnaire"`
	Fields        []model.RecruitmentField       `json:"fields"`
}

func (s *RecruitmentService) CreateQuestionnaire(userID uint, input RecruitmentQuestionnaireInput) (*RecruitmentQuestionnaireDTO, error) {
	if err := validateQuestionnaire(input); err != nil {
		return nil, err
	}
	questionnaire := &model.RecruitmentQuestionnaire{OwnerUserID: userID, Title: strings.TrimSpace(input.Title), Description: input.Description, Version: 1}
	if err := s.repo.CreateQuestionnaire(questionnaire, input.Fields); err != nil {
		return nil, err
	}
	return &RecruitmentQuestionnaireDTO{Questionnaire: *questionnaire, Fields: input.Fields}, nil
}

func (s *RecruitmentService) UpdateQuestionnaire(userID, id uint, input RecruitmentQuestionnaireInput, publish, admin bool) (*RecruitmentQuestionnaireDTO, error) {
	questionnaire, _, err := s.repo.GetQuestionnaire(id)
	if err != nil {
		return nil, err
	}
	if !admin && questionnaire.OwnerUserID != userID {
		return nil, errors.New("无权修改该问卷")
	}
	if questionnaire.Published {
		return nil, errors.New("已发布问卷不可原地修改，请创建新版本")
	}
	if err := validateQuestionnaire(input); err != nil {
		return nil, err
	}
	questionnaire.Title, questionnaire.Description = strings.TrimSpace(input.Title), input.Description
	questionnaire.Version++
	if publish {
		questionnaire.Published = true
	}
	if err := s.repo.UpdateQuestionnaire(questionnaire, input.Fields); err != nil {
		return nil, err
	}
	return &RecruitmentQuestionnaireDTO{Questionnaire: *questionnaire, Fields: input.Fields}, nil
}

func (s *RecruitmentService) GetQuestionnaire(userID, id uint, admin bool) (*RecruitmentQuestionnaireDTO, error) {
	questionnaire, fields, err := s.repo.GetQuestionnaire(id)
	if err != nil {
		return nil, err
	}
	if !admin && questionnaire.OwnerUserID != userID {
		return nil, errors.New("无权查看该问卷")
	}
	return &RecruitmentQuestionnaireDTO{Questionnaire: *questionnaire, Fields: fields}, nil
}

func (s *RecruitmentService) ListQuestionnaires(userID uint, admin bool) ([]model.RecruitmentQuestionnaire, error) {
	return s.repo.ListQuestionnaires(userID, admin)
}

var validRecruitmentFieldTypes = map[string]bool{"text": true, "textarea": true, "radio": true, "checkbox": true, "select": true, "number": true, "date": true, "image": true}

func validateQuestionnaire(input RecruitmentQuestionnaireInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("问卷标题不能为空")
	}
	keys := map[string]bool{}
	for i := range input.Fields {
		key := strings.TrimSpace(input.Fields[i].FieldKey)
		if key == "" {
			continue
		}
		if keys[key] {
			return fmt.Errorf("字段键 %s 重复", key)
		}
		keys[key] = true
	}
	nextKey := 1
	for i := range input.Fields {
		field := &input.Fields[i]
		field.FieldKey = strings.TrimSpace(field.FieldKey)
		field.Label = strings.TrimSpace(field.Label)
		field.Description = strings.TrimSpace(field.Description)
		if field.FieldKey == "" {
			for {
				candidate := fmt.Sprintf("question_%02d", nextKey)
				nextKey++
				if !keys[candidate] {
					field.FieldKey = candidate
					keys[candidate] = true
					break
				}
			}
		}
		if field.Label == "" || len([]rune(field.Label)) > 200 || len([]rune(field.Description)) > 500 || !validRecruitmentFieldTypes[field.Type] {
			return fmt.Errorf("第 %d 个字段无效", i+1)
		}
		if field.Type == "radio" || field.Type == "checkbox" || field.Type == "select" {
			var options []string
			if json.Unmarshal(field.Options, &options) != nil {
				return fmt.Errorf("字段 %s 必须配置选项", field.FieldKey)
			}
			cleaned := make([]string, 0, len(options))
			seenOptions := map[string]bool{}
			for _, option := range options {
				option = strings.TrimSpace(option)
				if option == "" || seenOptions[option] {
					continue
				}
				seenOptions[option] = true
				cleaned = append(cleaned, option)
			}
			if len(cleaned) < 2 {
				return fmt.Errorf("字段 %s 至少需要两个不同选项", field.FieldKey)
			}
			field.Options, _ = json.Marshal(cleaned)
		}
	}
	return nil
}

type CreateRecruitmentLinkInput struct {
	QuestionnaireID uint       `json:"questionnaire_id"`
	ExpiresAt       *time.Time `json:"expires_at"`
	MaxSubmissions  int        `json:"max_submissions"`
}

type CreatedRecruitmentLink struct {
	Link  model.RecruitmentLink `json:"link"`
	Token string                `json:"token"`
}

func (s *RecruitmentService) CreateLink(userID uint, input CreateRecruitmentLinkInput, admin bool) (*CreatedRecruitmentLink, error) {
	questionnaire, fields, err := s.repo.GetQuestionnaire(input.QuestionnaireID)
	if err != nil {
		return nil, err
	}
	if !admin && questionnaire.OwnerUserID != userID {
		return nil, errors.New("无权为该问卷创建链接")
	}
	if !questionnaire.Published {
		return nil, errors.New("问卷发布后才能创建报名链接")
	}
	if input.ExpiresAt != nil && input.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("有效期必须晚于当前时间")
	}
	if input.MaxSubmissions < 0 {
		return nil, errors.New("提交上限不能为负数")
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	snapshot, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	link := &model.RecruitmentLink{QuestionnaireID: questionnaire.ID, OwnerUserID: userID, TokenHash: hashRecruitmentToken(token), ExpiresAt: input.ExpiresAt, MaxSubmissions: input.MaxSubmissions, Active: true, QuestionnaireVersion: questionnaire.Version, FieldsSnapshot: snapshot}
	if err := s.repo.CreateLink(link); err != nil {
		return nil, err
	}
	return &CreatedRecruitmentLink{Link: *link, Token: token}, nil
}

type PublicRecruitmentContext struct {
	LinkID               uint                     `json:"link_id"`
	Title                string                   `json:"title"`
	Description          string                   `json:"description"`
	QuestionnaireVersion int                      `json:"questionnaire_version"`
	Fields               []model.RecruitmentField `json:"fields"`
	ExpiresAt            *time.Time               `json:"expires_at"`
}

func (s *RecruitmentService) GetPublicContext(token string) (*PublicRecruitmentContext, error) {
	link, err := s.activeLink(token)
	if err != nil {
		return nil, err
	}
	questionnaire, _, err := s.repo.GetQuestionnaire(link.QuestionnaireID)
	if err != nil {
		return nil, err
	}
	var fields []model.RecruitmentField
	if err := json.Unmarshal(link.FieldsSnapshot, &fields); err != nil {
		return nil, err
	}
	return &PublicRecruitmentContext{LinkID: link.ID, Title: questionnaire.Title, Description: questionnaire.Description, QuestionnaireVersion: link.QuestionnaireVersion, Fields: fields, ExpiresAt: link.ExpiresAt}, nil
}

func (s *RecruitmentService) activeLink(token string) (*model.RecruitmentLink, error) {
	link, err := s.repo.GetLinkByHash(hashRecruitmentToken(token))
	if err != nil {
		return nil, errors.New("报名链接不存在")
	}
	if !link.Active || (link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt)) || (link.MaxSubmissions > 0 && link.SubmissionCount >= link.MaxSubmissions) {
		return nil, errors.New("报名链接已失效")
	}
	return link, nil
}

func hashRecruitmentToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

type SubmitRecruitmentInput struct {
	Answers map[string]json.RawMessage `json:"answers"`
}

func (s *RecruitmentService) Submit(userID uint, token string, input SubmitRecruitmentInput) (*model.RecruitmentApplication, error) {
	link, err := s.repo.GetLinkByHash(hashRecruitmentToken(token))
	if err != nil {
		return nil, errors.New("报名链接不存在")
	}
	if !link.Active || (link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt)) {
		return nil, errors.New("报名链接已失效")
	}
	existing, existingErr := s.repo.GetApplicationByLinkAndUser(link.ID, userID)
	resubmitting := existingErr == nil && existing.Status == model.RecruitmentDraft
	if existingErr == nil && !resubmitting {
		return nil, errors.New("你已提交过该报名申请")
	}
	if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
		return nil, existingErr
	}
	if !resubmitting && link.MaxSubmissions > 0 && link.SubmissionCount >= link.MaxSubmissions {
		return nil, errors.New("报名链接已达到提交上限")
	}
	profile, err := s.userSvc.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	if profile.QQ == "" || profile.Nickname == "" || profile.TournamentCharacterName == "" || profile.TournamentCharacterID == nil {
		return nil, errors.New("请先完善 QQ、昵称和比赛服角色信息")
	}
	if profile.MainCharacterID <= 0 {
		return nil, errors.New("请先绑定正式服主角色")
	}
	character, err := s.charRepo.GetByCharacterID(profile.MainCharacterID)
	if err != nil {
		return nil, errors.New("正式服主角色不存在")
	}
	if _, err := s.skillRepo.GetSkill(int(profile.MainCharacterID)); err != nil {
		return nil, errors.New("正式服主角色技能尚未成功拉取")
	}
	skills, err := s.skillRepo.GetSkillList(int(profile.MainCharacterID))
	if err != nil {
		return nil, err
	}
	var fields []model.RecruitmentField
	if err := json.Unmarshal(link.FieldsSnapshot, &fields); err != nil {
		return nil, err
	}
	allowed := map[string]model.RecruitmentField{}
	for _, field := range fields {
		allowed[field.FieldKey] = field
		if field.Required && emptyJSON(input.Answers[field.FieldKey]) {
			return nil, fmt.Errorf("必填项 %s 未填写", field.Label)
		}
	}
	answers := make([]model.RecruitmentAnswer, 0, len(input.Answers))
	for key, value := range input.Answers {
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("未知问卷字段 %s", key)
		}
		if len(value) > 0 && !json.Valid(value) {
			return nil, fmt.Errorf("字段 %s 不是有效 JSON", key)
		}
		answers = append(answers, model.RecruitmentAnswer{FieldKey: key, Value: value})
	}
	snapshots := make([]model.RecruitmentSkillSnapshot, len(skills))
	for i, skill := range skills {
		snapshots[i] = model.RecruitmentSkillSnapshot{SkillID: skill.SkillID, ActiveLevel: skill.ActiveLevel, TrainedLevel: skill.TrainedLevel, SkillpointsInSkill: skill.SkillpointsInSkill}
	}
	now := time.Now()
	application := &model.RecruitmentApplication{LinkID: link.ID, QuestionnaireID: link.QuestionnaireID, UserID: userID, Status: model.RecruitmentSubmitted, MainCharacterID: character.CharacterID, MainCharacterName: character.CharacterName, PortraitURL: character.PortraitURL, CorporationID: character.CorporationID, AllianceID: character.AllianceID, FactionID: character.FactionID, SubmittedAt: &now}
	event := &model.RecruitmentReviewEvent{ActorUserID: userID, Action: "submit"}
	if resubmitting {
		application.ID = existing.ID
		event.Action = "resubmit"
		if err := s.repo.ResubmitApplication(application, answers, snapshots, event); err != nil {
			return nil, err
		}
		return application, nil
	}
	if err := s.repo.CreateApplication(application, answers, snapshots, event); err != nil {
		return nil, err
	}
	return application, nil
}

func emptyJSON(value json.RawMessage) bool {
	if len(value) == 0 {
		return true
	}
	trimmed := strings.TrimSpace(string(value))
	return trimmed == "" || trimmed == "null" || trimmed == `""` || trimmed == "[]"
}

func (s *RecruitmentService) ListLinks(userID uint, admin bool) ([]model.RecruitmentLink, error) {
	return s.repo.ListLinks(userID, admin)
}
func (s *RecruitmentService) SetLinkActive(userID, id uint, active, admin bool) error {
	link, err := s.repo.GetLink(id)
	if err != nil {
		return err
	}
	if !admin && link.OwnerUserID != userID {
		return errors.New("无权管理该链接")
	}
	return s.repo.SetLinkActive(id, active)
}
func (s *RecruitmentService) ListApplications(userID uint, admin bool) ([]model.RecruitmentApplication, error) {
	return s.repo.ListApplications(userID, admin)
}

func (s *RecruitmentService) ListMyApplications(userID uint) ([]model.RecruitmentApplication, error) {
	return s.repo.ListUserApplications(userID)
}

func (s *RecruitmentService) Withdraw(userID, id uint) error {
	detail, err := s.repo.GetApplication(id)
	if err != nil {
		return err
	}
	if detail.Application.UserID != userID {
		return errors.New("无权撤回该申请")
	}
	if detail.Application.Status != model.RecruitmentSubmitted && detail.Application.Status != model.RecruitmentReviewing {
		return errors.New("当前状态不能撤回")
	}
	return s.repo.UpdateApplicationStatus(id, userID, model.RecruitmentWithdrawn, "withdraw", "")
}

func (s *RecruitmentService) GetApplication(userID, id uint, admin bool) (*repository.RecruitmentApplicationDetail, error) {
	detail, err := s.repo.GetApplication(id)
	if err != nil {
		return nil, err
	}
	link, err := s.repo.GetLink(detail.Application.LinkID)
	if err != nil {
		return nil, err
	}
	if !admin && link.OwnerUserID != userID {
		return nil, errors.New("无权查看该申请")
	}
	return detail, nil
}

var reviewTransitions = map[string]map[string]bool{
	model.RecruitmentSubmitted: {model.RecruitmentReviewing: true, model.RecruitmentDraft: true},
	model.RecruitmentReviewing: {model.RecruitmentApproved: true, model.RecruitmentRejected: true, model.RecruitmentDraft: true},
}

func (s *RecruitmentService) Review(userID, id uint, status, reason string, admin bool) error {
	detail, err := s.GetApplication(userID, id, admin)
	if err != nil {
		return err
	}
	if !reviewTransitions[detail.Application.Status][status] {
		return errors.New("不允许的申请状态转换")
	}
	if (status == model.RecruitmentRejected || status == model.RecruitmentDraft) && strings.TrimSpace(reason) == "" {
		return errors.New("拒绝或退回必须填写原因")
	}
	action := "review"
	if status == model.RecruitmentDraft {
		action = "return"
	}
	if status == model.RecruitmentApproved {
		if err := s.repo.ApproveApplication(id, detail.Application.UserID, userID, reason); err != nil {
			return err
		}
		s.roleSvc.InvalidateUserCache(context.Background(), detail.Application.UserID)
		return nil
	}
	return s.repo.UpdateApplicationStatus(id, userID, status, action, reason)
}

func (s *RecruitmentService) Cleanup(userID uint, reason string) (int64, error) {
	if strings.TrimSpace(reason) == "" {
		return 0, errors.New("隐私清理必须填写原因")
	}
	return s.repo.CleanupPrivateData(userID, reason)
}
