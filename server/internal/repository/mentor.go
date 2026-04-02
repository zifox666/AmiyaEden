package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MentorRelationshipRepository struct{}

func NewMentorRelationshipRepository() *MentorRelationshipRepository {
	return &MentorRelationshipRepository{}
}

func (r *MentorRelationshipRepository) Create(row *model.MentorMenteeRelationship) error {
	return global.DB.Create(row).Error
}

func (r *MentorRelationshipRepository) CreateTx(tx *gorm.DB, row *model.MentorMenteeRelationship) error {
	return tx.Create(row).Error
}

func (r *MentorRelationshipRepository) GetByID(id uint) (*model.MentorMenteeRelationship, error) {
	var row model.MentorMenteeRelationship
	err := global.DB.First(&row, id).Error
	return &row, err
}

func (r *MentorRelationshipRepository) GetByIDForUpdateTx(tx *gorm.DB, id uint) (*model.MentorMenteeRelationship, error) {
	var row model.MentorMenteeRelationship
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error
	return &row, err
}

func (r *MentorRelationshipRepository) GetActiveOrPendingByMenteeUserID(menteeUserID uint) (*model.MentorMenteeRelationship, error) {
	var row model.MentorMenteeRelationship
	err := global.DB.
		Where("mentee_user_id = ? AND status IN ?", menteeUserID, []string{model.MentorRelationStatusPending, model.MentorRelationStatusActive}).
		Order("applied_at DESC, id DESC").
		First(&row).Error
	return &row, err
}

func (r *MentorRelationshipRepository) UpdateStatus(id uint, status string, updates map[string]any) error {
	if updates == nil {
		updates = map[string]any{}
	}
	updates["status"] = status
	return global.DB.Model(&model.MentorMenteeRelationship{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MentorRelationshipRepository) UpdateStatusTx(tx *gorm.DB, id uint, status string, updates map[string]any) error {
	if updates == nil {
		updates = map[string]any{}
	}
	updates["status"] = status
	return tx.Model(&model.MentorMenteeRelationship{}).Where("id = ?", id).Updates(updates).Error
}

func buildMentorRelationshipListQuery(db *gorm.DB, mentorUserID uint, statuses []string) *gorm.DB {
	query := db.Model(&model.MentorMenteeRelationship{}).
		Where("mentor_mentee_relationship.mentor_user_id = ?", mentorUserID)
	if len(statuses) > 0 {
		query = query.Where("mentor_mentee_relationship.status IN ?", statuses)
	}
	return query
}

func applyMenteeLastLoginOrdering(query *gorm.DB) *gorm.DB {
	return query.
		Joins(`LEFT JOIN "user" AS mentee_user ON mentee_user.id = mentor_mentee_relationship.mentee_user_id`).
		Joins(`LEFT JOIN eve_character AS mentee_primary_character ON mentee_primary_character.character_id = mentee_user.primary_character_id`).
		Order(`mentee_user.last_login_at DESC NULLS LAST`).
		Order(`mentee_primary_character.character_name ASC`).
		Order(`mentor_mentee_relationship.id ASC`)
}

func buildMentorRelationshipListSelectQuery(query *gorm.DB, page, pageSize int) *gorm.DB {
	offset := (page - 1) * pageSize
	return applyMenteeLastLoginOrdering(query).Offset(offset).Limit(pageSize)
}

func buildPendingMentorRelationshipListSelectQuery(query *gorm.DB) *gorm.DB {
	return applyMenteeLastLoginOrdering(query)
}

func (r *MentorRelationshipRepository) ListByMentorUserID(mentorUserID uint, statuses []string, page, pageSize int) ([]model.MentorMenteeRelationship, int64, error) {
	var rows []model.MentorMenteeRelationship
	var total int64
	query := buildMentorRelationshipListQuery(global.DB, mentorUserID, statuses)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := buildMentorRelationshipListSelectQuery(query, page, pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *MentorRelationshipRepository) ListPendingByMentorUserID(mentorUserID uint) ([]model.MentorMenteeRelationship, error) {
	var rows []model.MentorMenteeRelationship
	err := buildPendingMentorRelationshipListSelectQuery(
		buildMentorRelationshipListQuery(global.DB, mentorUserID, []string{model.MentorRelationStatusPending}),
	).Find(&rows).Error
	return rows, err
}

type MentorRelationshipAdminFilter struct {
	Status  string
	Keyword string
}

func buildMentorRelationshipAdminListQuery(db *gorm.DB, filter MentorRelationshipAdminFilter) *gorm.DB {
	query := db.Model(&model.MentorMenteeRelationship{})
	if filter.Status != "" {
		query = query.Where("mentor_mentee_relationship.status = ?", filter.Status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.
			Joins(`LEFT JOIN "user" AS mentor_user ON mentor_user.id = mentor_mentee_relationship.mentor_user_id`).
			Joins(`LEFT JOIN eve_character AS mentor_character ON mentor_character.character_id = mentor_user.primary_character_id`).
			Joins(`LEFT JOIN "user" AS mentee_user ON mentee_user.id = mentor_mentee_relationship.mentee_user_id`).
			Joins(`LEFT JOIN eve_character AS mentee_character ON mentee_character.character_id = mentee_user.primary_character_id`).
			Where(`mentor_user.nickname ILIKE ? OR mentor_character.character_name ILIKE ? OR mentee_user.nickname ILIKE ? OR mentee_character.character_name ILIKE ?`, like, like, like, like)
	}
	return query
}

func (r *MentorRelationshipRepository) ListAllPaged(filter MentorRelationshipAdminFilter, page, pageSize int) ([]model.MentorMenteeRelationship, int64, error) {
	var rows []model.MentorMenteeRelationship
	var total int64
	offset := (page - 1) * pageSize
	query := buildMentorRelationshipAdminListQuery(global.DB, filter)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("mentor_mentee_relationship.applied_at DESC").Order("mentor_mentee_relationship.id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *MentorRelationshipRepository) ListActiveRelationships() ([]model.MentorMenteeRelationship, error) {
	var rows []model.MentorMenteeRelationship
	err := global.DB.Where("status = ?", model.MentorRelationStatusActive).Find(&rows).Error
	return rows, err
}

func (r *MentorRelationshipRepository) CountActiveByMentorUserID(mentorUserID uint) (int64, error) {
	var count int64
	err := buildMentorRelationshipListQuery(global.DB, mentorUserID, []string{model.MentorRelationStatusActive}).Count(&count).Error
	return count, err
}

func (r *MentorRelationshipRepository) CountPendingByMentorUserID(mentorUserID uint) (int64, error) {
	var count int64
	err := global.DB.Model(&model.MentorMenteeRelationship{}).
		Where("mentor_mentee_relationship.mentor_user_id = ?", mentorUserID).
		Where("mentor_mentee_relationship.status = ?", model.MentorRelationStatusPending).
		Count(&count).Error
	return count, err
}

func (r *MentorRelationshipRepository) CountActiveByMentorUserIDs(mentorUserIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(mentorUserIDs))
	if len(mentorUserIDs) == 0 {
		return result, nil
	}

	type countRow struct {
		MentorUserID uint
		Count        int64
	}

	var rows []countRow
	err := global.DB.Model(&model.MentorMenteeRelationship{}).
		Select("mentor_user_id, COUNT(*) AS count").
		Where("mentor_user_id IN ? AND status = ?", mentorUserIDs, model.MentorRelationStatusActive).
		Group("mentor_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.MentorUserID] = row.Count
	}
	return result, nil
}

type MentorRewardStageRepository struct{}

func NewMentorRewardStageRepository() *MentorRewardStageRepository {
	return &MentorRewardStageRepository{}
}

func (r *MentorRewardStageRepository) ListAll() ([]model.MentorRewardStage, error) {
	var rows []model.MentorRewardStage
	err := global.DB.Order("stage_order ASC").Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *MentorRewardStageRepository) ReplaceAll(stages []model.MentorRewardStage) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&model.MentorRewardStage{}).Error; err != nil {
			return err
		}
		if len(stages) == 0 {
			return nil
		}
		return tx.Create(&stages).Error
	})
}

type MentorRewardDistributionRepository struct{}

func NewMentorRewardDistributionRepository() *MentorRewardDistributionRepository {
	return &MentorRewardDistributionRepository{}
}

type MentorRewardDistributionAdminFilter struct {
	Keyword string
}

func (r *MentorRewardDistributionRepository) Create(row *model.MentorRewardDistribution) error {
	return global.DB.Create(row).Error
}

func (r *MentorRewardDistributionRepository) CreateTx(tx *gorm.DB, row *model.MentorRewardDistribution) error {
	return tx.Create(row).Error
}

func (r *MentorRewardDistributionRepository) ExistsByRelationshipAndStageOrder(relationshipID uint, stageOrder int) (bool, error) {
	var count int64
	err := global.DB.Model(&model.MentorRewardDistribution{}).
		Where("relationship_id = ? AND stage_order = ?", relationshipID, stageOrder).
		Count(&count).Error
	return count > 0, err
}

func buildMentorRewardDistributionAdminListQuery(db *gorm.DB, filter MentorRewardDistributionAdminFilter) *gorm.DB {
	query := db.Model(&model.MentorRewardDistribution{})
	return applyKeywordLikeFilter(
		query,
		filter.Keyword,
		`LOWER(mentor_reward_distribution.mentor_nickname) LIKE ?`,
		`LOWER(mentor_reward_distribution.mentor_character_name) LIKE ?`,
	)
}

func (r *MentorRewardDistributionRepository) ListAdminPaged(
	filter MentorRewardDistributionAdminFilter,
	page int,
	pageSize int,
) ([]model.MentorRewardDistribution, int64, error) {
	var rows []model.MentorRewardDistribution
	var total int64
	offset := (page - 1) * pageSize
	query := buildMentorRewardDistributionAdminListQuery(global.DB, filter)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Order("mentor_reward_distribution.distributed_at DESC").
		Order("mentor_reward_distribution.id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *MentorRewardDistributionRepository) ListMissingSnapshots() ([]model.MentorRewardDistribution, error) {
	var rows []model.MentorRewardDistribution
	err := global.DB.
		Where(`mentor_character_name = '' OR mentor_nickname = '' OR mentee_character_name = '' OR mentee_nickname = ''`).
		Find(&rows).Error
	return rows, err
}

func (r *MentorRewardDistributionRepository) UpdateSnapshots(
	id uint,
	mentorCharacterName string,
	mentorNickname string,
	menteeCharacterName string,
	menteeNickname string,
) error {
	return global.DB.Model(&model.MentorRewardDistribution{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"mentor_character_name": mentorCharacterName,
			"mentor_nickname":       mentorNickname,
			"mentee_character_name": menteeCharacterName,
			"mentee_nickname":       menteeNickname,
		}).Error
}

func (r *MentorRewardDistributionRepository) ExistsByRelationshipAndStageOrderTx(tx *gorm.DB, relationshipID uint, stageOrder int) (bool, error) {
	var count int64
	err := tx.Model(&model.MentorRewardDistribution{}).
		Where("relationship_id = ? AND stage_order = ?", relationshipID, stageOrder).
		Count(&count).Error
	return count > 0, err
}

func (r *MentorRewardDistributionRepository) ListDistributedStageOrdersByRelationshipIDs(relationshipIDs []uint) (map[uint][]int, error) {
	result := make(map[uint][]int, len(relationshipIDs))
	if len(relationshipIDs) == 0 {
		return result, nil
	}

	type row struct {
		RelationshipID uint
		StageOrder     int
	}

	var rows []row
	err := global.DB.Model(&model.MentorRewardDistribution{}).
		Select("relationship_id, stage_order").
		Where("relationship_id IN ?", relationshipIDs).
		Order("stage_order ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RelationshipID] = append(result[row.RelationshipID], row.StageOrder)
	}
	return result, nil
}

func buildMentorRewardAmountSummaryQuery(db *gorm.DB, relationshipIDs []uint) *gorm.DB {
	return db.Model(&model.MentorRewardDistribution{}).
		Select("relationship_id, SUM(reward_amount) AS total_reward_amount").
		Where("relationship_id IN ?", relationshipIDs).
		Group("relationship_id")
}

func (r *MentorRewardDistributionRepository) SumRewardAmountsByRelationshipIDs(relationshipIDs []uint) (map[uint]float64, error) {
	result := make(map[uint]float64, len(relationshipIDs))
	if len(relationshipIDs) == 0 {
		return result, nil
	}

	type row struct {
		RelationshipID    uint
		TotalRewardAmount float64
	}

	var rows []row
	err := buildMentorRewardAmountSummaryQuery(global.DB, relationshipIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RelationshipID] = row.TotalRewardAmount
	}
	return result, nil
}

func IsActiveMentorRelationConflictError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "idx_mentor_rel_active_mentee") || strings.Contains(msg, "duplicate key")
}
