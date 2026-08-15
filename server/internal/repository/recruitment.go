package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RecruitmentRepository struct{}

func NewRecruitmentRepository() *RecruitmentRepository { return &RecruitmentRepository{} }

func (r *RecruitmentRepository) CreateQuestionnaire(questionnaire *model.RecruitmentQuestionnaire, fields []model.RecruitmentField) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(questionnaire).Error; err != nil {
			return err
		}
		for i := range fields {
			fields[i].QuestionnaireID = questionnaire.ID
		}
		if len(fields) > 0 {
			return tx.Create(&fields).Error
		}
		return nil
	})
}

func (r *RecruitmentRepository) UpdateQuestionnaire(questionnaire *model.RecruitmentQuestionnaire, fields []model.RecruitmentField) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.RecruitmentQuestionnaire{}).Where("id = ?", questionnaire.ID).Updates(map[string]any{
			"title": questionnaire.Title, "description": questionnaire.Description, "version": questionnaire.Version, "published": questionnaire.Published,
		}).Error; err != nil {
			return err
		}
		// Draft fields are replaced as one schema. A soft delete would retain
		// the unique (questionnaire_id, field_key) values and make publishing
		// an unchanged draft fail with SQLSTATE 23505.
		if err := tx.Unscoped().Where("questionnaire_id = ?", questionnaire.ID).Delete(&model.RecruitmentField{}).Error; err != nil {
			return err
		}
		for i := range fields {
			fields[i].QuestionnaireID = questionnaire.ID
		}
		if len(fields) > 0 {
			return tx.Create(&fields).Error
		}
		return nil
	})
}

func (r *RecruitmentRepository) GetQuestionnaire(id uint) (*model.RecruitmentQuestionnaire, []model.RecruitmentField, error) {
	var questionnaire model.RecruitmentQuestionnaire
	if err := global.DB.First(&questionnaire, id).Error; err != nil {
		return nil, nil, err
	}
	var fields []model.RecruitmentField
	err := global.DB.Where("questionnaire_id = ?", id).Order("sort ASC, id ASC").Find(&fields).Error
	return &questionnaire, fields, err
}

func (r *RecruitmentRepository) ListQuestionnaires(ownerUserID uint, all bool) ([]model.RecruitmentQuestionnaire, error) {
	var result []model.RecruitmentQuestionnaire
	query := global.DB.Order("id DESC")
	if !all {
		query = query.Where("owner_user_id = ?", ownerUserID)
	}
	return result, query.Find(&result).Error
}

func (r *RecruitmentRepository) CreateLink(link *model.RecruitmentLink) error {
	return global.DB.Create(link).Error
}

func (r *RecruitmentRepository) GetLinkByHash(hash string) (*model.RecruitmentLink, error) {
	var link model.RecruitmentLink
	err := global.DB.Where("token_hash = ?", hash).First(&link).Error
	return &link, err
}

func (r *RecruitmentRepository) GetLink(id uint) (*model.RecruitmentLink, error) {
	var link model.RecruitmentLink
	err := global.DB.First(&link, id).Error
	return &link, err
}

func (r *RecruitmentRepository) ListLinks(ownerUserID uint, all bool) ([]model.RecruitmentLink, error) {
	var result []model.RecruitmentLink
	query := global.DB.Order("id DESC")
	if !all {
		query = query.Where("owner_user_id = ?", ownerUserID)
	}
	return result, query.Find(&result).Error
}

func (r *RecruitmentRepository) SetLinkActive(id uint, active bool) error {
	return global.DB.Model(&model.RecruitmentLink{}).Where("id = ?", id).Update("active", active).Error
}

func (r *RecruitmentRepository) CreateApplication(application *model.RecruitmentApplication, answers []model.RecruitmentAnswer, skills []model.RecruitmentSkillSnapshot, event *model.RecruitmentReviewEvent) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var link model.RecruitmentLink
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&link, application.LinkID).Error; err != nil {
			return err
		}
		if !link.Active || (link.MaxSubmissions > 0 && link.SubmissionCount >= link.MaxSubmissions) {
			return gorm.ErrInvalidData
		}
		if err := tx.Create(application).Error; err != nil {
			return err
		}
		for i := range answers {
			answers[i].ApplicationID = application.ID
		}
		for i := range skills {
			skills[i].ApplicationID = application.ID
		}
		if len(answers) > 0 {
			if err := tx.Create(&answers).Error; err != nil {
				return err
			}
		}
		if len(skills) > 0 {
			if err := tx.Create(&skills).Error; err != nil {
				return err
			}
		}
		event.ApplicationID = application.ID
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return tx.Model(&model.RecruitmentLink{}).Where("id = ?", link.ID).UpdateColumn("submission_count", gorm.Expr("submission_count + 1")).Error
	})
}

func (r *RecruitmentRepository) GetApplicationByLinkAndUser(linkID, userID uint) (*model.RecruitmentApplication, error) {
	var application model.RecruitmentApplication
	return &application, global.DB.Where("link_id = ? AND user_id = ?", linkID, userID).First(&application).Error
}

func (r *RecruitmentRepository) ResubmitApplication(application *model.RecruitmentApplication, answers []model.RecruitmentAnswer, skills []model.RecruitmentSkillSnapshot, event *model.RecruitmentReviewEvent) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var current model.RecruitmentApplication
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, application.ID).Error; err != nil {
			return err
		}
		if current.Status != model.RecruitmentDraft {
			return gorm.ErrInvalidData
		}
		if err := tx.Unscoped().Where("application_id = ?", application.ID).Delete(&model.RecruitmentAnswer{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("application_id = ?", application.ID).Delete(&model.RecruitmentSkillSnapshot{}).Error; err != nil {
			return err
		}
		for i := range answers {
			answers[i].ApplicationID = application.ID
		}
		for i := range skills {
			skills[i].ApplicationID = application.ID
		}
		if len(answers) > 0 {
			if err := tx.Create(&answers).Error; err != nil {
				return err
			}
		}
		if len(skills) > 0 {
			if err := tx.Create(&skills).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"status": model.RecruitmentSubmitted, "main_character_id": application.MainCharacterID,
			"main_character_name": application.MainCharacterName, "portrait_url": application.PortraitURL,
			"corporation_id": application.CorporationID, "alliance_id": application.AllianceID,
			"faction_id": application.FactionID, "submitted_at": application.SubmittedAt,
		}).Error; err != nil {
			return err
		}
		event.ApplicationID = application.ID
		return tx.Create(event).Error
	})
}

type RecruitmentApplicationDetail struct {
	Application model.RecruitmentApplication     `json:"application"`
	Fields      []model.RecruitmentField         `json:"fields"`
	Answers     []model.RecruitmentAnswer        `json:"answers"`
	Skills      []model.RecruitmentSkillSnapshot `json:"skills"`
	Events      []model.RecruitmentReviewEvent   `json:"events"`
}

func (r *RecruitmentRepository) ListApplications(ownerUserID uint, all bool) ([]model.RecruitmentApplication, error) {
	var result []model.RecruitmentApplication
	query := global.DB.Table("recruitment_application a").Select("a.*").Order("a.id DESC")
	if !all {
		query = query.Joins("JOIN recruitment_link l ON l.id = a.link_id").Where("l.owner_user_id = ?", ownerUserID)
	}
	return result, query.Scan(&result).Error
}

func (r *RecruitmentRepository) ListUserApplications(userID uint) ([]model.RecruitmentApplication, error) {
	var result []model.RecruitmentApplication
	return result, global.DB.Where("user_id = ?", userID).Order("id DESC").Find(&result).Error
}

func (r *RecruitmentRepository) GetApplication(id uint) (*RecruitmentApplicationDetail, error) {
	var detail RecruitmentApplicationDetail
	if err := global.DB.First(&detail.Application, id).Error; err != nil {
		return nil, err
	}
	var link model.RecruitmentLink
	if err := global.DB.First(&link, detail.Application.LinkID).Error; err != nil {
		return nil, err
	}
	if err := json.Unmarshal(link.FieldsSnapshot, &detail.Fields); err != nil {
		return nil, err
	}
	if err := global.DB.Where("application_id = ?", id).Find(&detail.Answers).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("application_id = ?", id).Find(&detail.Skills).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("application_id = ?", id).Order("id ASC").Find(&detail.Events).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *RecruitmentRepository) UpdateApplicationStatus(id, actorUserID uint, status, action, reason string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.RecruitmentApplication{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		return tx.Create(&model.RecruitmentReviewEvent{ApplicationID: id, ActorUserID: actorUserID, Action: action, Reason: reason}).Error
	})
}

// ApproveApplication promotes a guest applicant to the ordinary member role
// in the same transaction as the review decision.
func (r *RecruitmentRepository) ApproveApplication(id, applicantUserID, actorUserID uint, reason string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var userRole model.Role
		if err := tx.Where("code = ? AND status = ?", model.RoleUser, 1).First(&userRole).Error; err != nil {
			return err
		}
		var guestRole model.Role
		if err := tx.Where("code = ?", model.RoleGuest).First(&guestRole).Error; err == nil {
			if err := tx.Where("user_id = ? AND role_id = ?", applicantUserID, guestRole.ID).Delete(&model.UserRole{}).Error; err != nil {
				return err
			}
		}
		membership := model.UserRole{UserID: applicantUserID, RoleID: userRole.ID, IsAuto: false}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&membership).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.User{}).
			Where("id = ? AND (role = ? OR role = '' OR role IS NULL)", applicantUserID, model.RoleGuest).
			Update("role", model.RoleUser).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.RecruitmentApplication{}).Where("id = ?", id).Update("status", model.RecruitmentApproved).Error; err != nil {
			return err
		}
		return tx.Create(&model.RecruitmentReviewEvent{ApplicationID: id, ActorUserID: actorUserID, Action: "approve_and_grant_member", Reason: reason}).Error
	})
}

func (r *RecruitmentRepository) CleanupPrivateData(actorUserID uint, reason string) (int64, error) {
	var applicationIDs []uint
	if err := global.DB.Model(&model.RecruitmentApplication{}).Pluck("id", &applicationIDs).Error; err != nil {
		return 0, err
	}
	if len(applicationIDs) == 0 {
		return 0, nil
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("application_id IN ?", applicationIDs).Delete(&model.RecruitmentAnswer{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("application_id IN ?", applicationIDs).Delete(&model.RecruitmentSkillSnapshot{}).Error; err != nil {
			return err
		}
		events := make([]model.RecruitmentReviewEvent, len(applicationIDs))
		for i, id := range applicationIDs {
			events[i] = model.RecruitmentReviewEvent{ApplicationID: id, ActorUserID: actorUserID, Action: "privacy_cleanup", Reason: reason}
		}
		return tx.Create(&events).Error
	})
	return int64(len(applicationIDs)), err
}
