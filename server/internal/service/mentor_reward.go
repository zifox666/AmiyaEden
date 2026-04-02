package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MentorRewardStageInput struct {
	StageOrder    int     `json:"stage_order" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	ConditionType string  `json:"condition_type" binding:"required"`
	Threshold     float64 `json:"threshold" binding:"required"`
	RewardAmount  float64 `json:"reward_amount" binding:"required"`
}

func validateMentorRewardStageInputs(inputs []MentorRewardStageInput) error {
	stageOrders := make(map[int]struct{}, len(inputs))
	for i, input := range inputs {
		if input.StageOrder <= 0 {
			return fmt.Errorf("阶段 %d: 序号必须大于 0", i+1)
		}
		if _, exists := stageOrders[input.StageOrder]; exists {
			return errors.New("阶段序号必须严格递增且不可重复")
		}
		stageOrders[input.StageOrder] = struct{}{}
		switch input.ConditionType {
		case model.MentorConditionSkillPoints, model.MentorConditionPapCount, model.MentorConditionDaysActive:
		default:
			return fmt.Errorf("阶段 %d: 无效的条件类型", i+1)
		}
		if input.Threshold != math.Trunc(input.Threshold) {
			return fmt.Errorf("阶段 %d: 阈值必须为整数", i+1)
		}
		if input.Threshold <= 0 {
			return fmt.Errorf("阶段 %d: 阈值必须大于 0", i+1)
		}
		if input.RewardAmount != math.Trunc(input.RewardAmount) {
			return fmt.Errorf("阶段 %d: 奖励金额必须为整数", i+1)
		}
		if input.RewardAmount <= 0 {
			return fmt.Errorf("阶段 %d: 奖励金额必须大于 0", i+1)
		}
	}
	return nil
}

func buildMentorRewardStages(inputs []MentorRewardStageInput) []model.MentorRewardStage {
	stages := make([]model.MentorRewardStage, 0, len(inputs))
	for _, input := range inputs {
		stages = append(stages, model.MentorRewardStage{
			StageOrder:    input.StageOrder,
			Name:          input.Name,
			ConditionType: input.ConditionType,
			Threshold:     input.Threshold,
			RewardAmount:  input.RewardAmount,
		})
	}
	sort.Slice(stages, func(i, j int) bool {
		return stages[i].StageOrder < stages[j].StageOrder
	})
	return stages
}

type mentorMetrics struct {
	TotalSP    int64
	TotalPap   float64
	DaysActive int
}

func isMentorConditionMet(stage model.MentorRewardStage, metrics *mentorMetrics) bool {
	switch stage.ConditionType {
	case model.MentorConditionSkillPoints:
		return float64(metrics.TotalSP) >= stage.Threshold
	case model.MentorConditionPapCount:
		return metrics.TotalPap >= stage.Threshold
	case model.MentorConditionDaysActive:
		return float64(metrics.DaysActive) >= stage.Threshold
	default:
		return false
	}
}

type MentorRewardProcessResult struct {
	ProcessedRelationships int     `json:"processed_relationships"`
	RewardsDistributed     int     `json:"rewards_distributed"`
	TotalCoinAwarded       float64 `json:"total_coin_awarded"`
	GraduatedCount         int     `json:"graduated_count"`
}

type MentorRewardDistributionView struct {
	ID                  uint      `json:"id"`
	RelationshipID      uint      `json:"relationship_id"`
	StageID             uint      `json:"stage_id"`
	StageOrder          int       `json:"stage_order"`
	MentorUserID        uint      `json:"mentor_user_id"`
	MentorCharacterName string    `json:"mentor_character_name"`
	MentorNickname      string    `json:"mentor_nickname"`
	MenteeUserID        uint      `json:"mentee_user_id"`
	MenteeCharacterName string    `json:"mentee_character_name"`
	MenteeNickname      string    `json:"mentee_nickname"`
	RewardAmount        float64   `json:"reward_amount"`
	DistributedAt       time.Time `json:"distributed_at"`
	WalletRefID         string    `json:"wallet_ref_id"`
}

type mentorRelationshipProcessOutcome struct {
	Processed          bool
	RewardsDistributed int
	TotalCoinAwarded   float64
	Graduated          bool
}

type mentorRewardDistributionSnapshot struct {
	mentorCharacterName string
	mentorNickname      string
	menteeCharacterName string
	menteeNickname      string
}

type MentorRewardService struct {
	stageRepo *repository.MentorRewardStageRepository
	distRepo  *repository.MentorRewardDistributionRepository
	relRepo   *repository.MentorRelationshipRepository
	userRepo  *repository.UserRepository
	charRepo  *repository.EveCharacterRepository
	skillRepo *repository.EveSkillRepository
	fleetRepo *repository.FleetRepository
	walletSvc *SysWalletService
}

func NewMentorRewardService() *MentorRewardService {
	return &MentorRewardService{
		stageRepo: repository.NewMentorRewardStageRepository(),
		distRepo:  repository.NewMentorRewardDistributionRepository(),
		relRepo:   repository.NewMentorRelationshipRepository(),
		userRepo:  repository.NewUserRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		skillRepo: repository.NewEveSkillRepository(),
		fleetRepo: repository.NewFleetRepository(),
		walletSvc: NewSysWalletService(),
	}
}

func (s *MentorRewardService) GetStages() ([]model.MentorRewardStage, error) {
	return s.stageRepo.ListAll()
}

func (s *MentorRewardService) UpdateStages(inputs []MentorRewardStageInput) ([]model.MentorRewardStage, error) {
	if err := validateMentorRewardStageInputs(inputs); err != nil {
		return nil, err
	}
	if err := s.stageRepo.ReplaceAll(buildMentorRewardStages(inputs)); err != nil {
		return nil, err
	}
	return s.stageRepo.ListAll()
}

func (s *MentorRewardService) ListAdminRewardDistributions(
	page,
	pageSize int,
	keyword string,
) ([]MentorRewardDistributionView, int64, error) {
	page = normalizePage(page)
	pageSize = normalizeLedgerPageSize(pageSize)
	if err := s.backfillMissingRewardDistributionSnapshots(); err != nil {
		return nil, 0, err
	}

	rows, total, err := s.distRepo.ListAdminPaged(repository.MentorRewardDistributionAdminFilter{
		Keyword: keyword,
	}, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []MentorRewardDistributionView{}, total, nil
	}

	result := make([]MentorRewardDistributionView, 0, len(rows))
	for _, row := range rows {
		result = append(result, MentorRewardDistributionView{
			ID:                  row.ID,
			RelationshipID:      row.RelationshipID,
			StageID:             row.StageID,
			StageOrder:          row.StageOrder,
			MentorUserID:        row.MentorUserID,
			MentorCharacterName: row.MentorCharacterName,
			MentorNickname:      row.MentorNickname,
			MenteeUserID:        row.MenteeUserID,
			MenteeCharacterName: row.MenteeCharacterName,
			MenteeNickname:      row.MenteeNickname,
			RewardAmount:        row.RewardAmount,
			DistributedAt:       row.DistributedAt,
			WalletRefID:         row.WalletRefID,
		})
	}

	return result, total, nil
}

func (s *MentorRewardService) backfillMissingRewardDistributionSnapshots() error {
	rows, err := s.distRepo.ListMissingSnapshots()
	if err != nil || len(rows) == 0 {
		return err
	}

	userIDs := make([]uint, 0, len(rows)*2)
	for _, row := range rows {
		userIDs = append(userIDs, row.MentorUserID, row.MenteeUserID)
	}
	users, err := s.userRepo.ListByIDs(userIDs)
	if err != nil {
		return err
	}

	userByID := make(map[uint]model.User, len(users))
	characterIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userByID[user.ID] = user
		if user.PrimaryCharacterID != 0 {
			characterIDs = append(characterIDs, user.PrimaryCharacterID)
		}
	}

	characters, err := s.charRepo.ListByCharacterIDs(characterIDs)
	if err != nil {
		return err
	}
	characterByID := make(map[int64]model.EveCharacter, len(characters))
	for _, character := range characters {
		characterByID[character.CharacterID] = character
	}

	for _, row := range rows {
		mentorUser := userByID[row.MentorUserID]
		menteeUser := userByID[row.MenteeUserID]

		mentorCharacterName := row.MentorCharacterName
		if mentorCharacterName == "" {
			mentorCharacterName = characterByID[mentorUser.PrimaryCharacterID].CharacterName
		}
		mentorNickname := row.MentorNickname
		if mentorNickname == "" {
			mentorNickname = mentorUser.Nickname
		}
		menteeCharacterName := row.MenteeCharacterName
		if menteeCharacterName == "" {
			menteeCharacterName = characterByID[menteeUser.PrimaryCharacterID].CharacterName
		}
		menteeNickname := row.MenteeNickname
		if menteeNickname == "" {
			menteeNickname = menteeUser.Nickname
		}

		if mentorCharacterName == row.MentorCharacterName &&
			mentorNickname == row.MentorNickname &&
			menteeCharacterName == row.MenteeCharacterName &&
			menteeNickname == row.MenteeNickname {
			continue
		}

		if err := s.distRepo.UpdateSnapshots(
			row.ID,
			mentorCharacterName,
			mentorNickname,
			menteeCharacterName,
			menteeNickname,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *MentorRewardService) ProcessRewards(now time.Time) (*MentorRewardProcessResult, error) {
	stages, err := s.stageRepo.ListAll()
	if err != nil {
		return nil, err
	}
	if len(stages) == 0 {
		return &MentorRewardProcessResult{}, nil
	}

	relationships, err := s.relRepo.ListActiveRelationships()
	if err != nil {
		return nil, err
	}

	result := &MentorRewardProcessResult{}
	for _, relationship := range relationships {
		metrics, err := s.getMenteeMetrics(relationship.MenteeUserID)
		if err != nil {
			global.Logger.Error("导师奖励处理：获取学员指标失败", zap.Uint("mentee_user_id", relationship.MenteeUserID), zap.Error(err))
			continue
		}

		outcome, err := s.processActiveRelationshipSnapshot(relationship, stages, metrics, now)
		if err != nil {
			global.Logger.Error("导师奖励处理：处理关系失败", zap.Uint("relationship_id", relationship.ID), zap.Error(err))
			continue
		}
		if !outcome.Processed {
			continue
		}
		result.ProcessedRelationships++
		result.RewardsDistributed += outcome.RewardsDistributed
		result.TotalCoinAwarded += outcome.TotalCoinAwarded
		if outcome.Graduated {
			result.GraduatedCount++
		}
	}

	return result, nil
}

func (s *MentorRewardService) processActiveRelationshipSnapshot(
	relationship model.MentorMenteeRelationship,
	stages []model.MentorRewardStage,
	metrics *mentorMetrics,
	now time.Time,
) (*mentorRelationshipProcessOutcome, error) {
	outcome := &mentorRelationshipProcessOutcome{}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		currentRelationship, err := s.relRepo.GetByIDForUpdateTx(tx, relationship.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if currentRelationship.Status != model.MentorRelationStatusActive {
			return nil
		}

		outcome.Processed = true
		var distributionSnapshot *mentorRewardDistributionSnapshot
		allDistributed := true
		for _, stage := range stages {
			exists, err := s.distRepo.ExistsByRelationshipAndStageOrderTx(tx, currentRelationship.ID, stage.StageOrder)
			if err != nil {
				return fmt.Errorf("check stage %d distribution: %w", stage.StageOrder, err)
			}
			if exists {
				continue
			}
			if !isMentorConditionMet(stage, metrics) {
				allDistributed = false
				break
			}
			if distributionSnapshot == nil {
				distributionSnapshot, err = s.buildRewardDistributionSnapshot(*currentRelationship)
				if err != nil {
					return fmt.Errorf("build reward distribution snapshot: %w", err)
				}
			}
			if err := s.distributeStageRewardTx(tx, *currentRelationship, stage, distributionSnapshot, now); err != nil {
				return fmt.Errorf("distribute stage %d reward: %w", stage.StageOrder, err)
			}
			outcome.RewardsDistributed++
			outcome.TotalCoinAwarded += stage.RewardAmount
		}

		if !allDistributed {
			return nil
		}
		if err := s.relRepo.UpdateStatusTx(tx, currentRelationship.ID, model.MentorRelationStatusGraduated, map[string]any{"graduated_at": now}); err != nil {
			return fmt.Errorf("mark relationship graduated: %w", err)
		}
		outcome.Graduated = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outcome, nil
}

func (s *MentorRewardService) getMenteeMetrics(userID uint) (*mentorMetrics, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	characterIDs := make([]int64, 0, len(characters))
	for _, character := range characters {
		characterIDs = append(characterIDs, character.CharacterID)
	}
	totalSP, err := s.skillRepo.SumTotalSPByCharacterIDs(characterIDs)
	if err != nil {
		return nil, err
	}
	totalPap, err := s.fleetRepo.SumPapByUserTotal(userID)
	if err != nil {
		return nil, err
	}
	return &mentorMetrics{
		TotalSP:    totalSP,
		TotalPap:   totalPap,
		DaysActive: calculateMentorDaysActive(user.CreatedAt, user.LastLoginAt),
	}, nil
}

func (s *MentorRewardService) buildRewardDistributionSnapshot(
	rel model.MentorMenteeRelationship,
) (*mentorRewardDistributionSnapshot, error) {
	mentorUser, err := s.userRepo.GetByID(rel.MentorUserID)
	if err != nil {
		return nil, err
	}
	menteeUser, err := s.userRepo.GetByID(rel.MenteeUserID)
	if err != nil {
		return nil, err
	}

	mentorCharacterName, err := s.getRewardDistributionCharacterName(mentorUser.PrimaryCharacterID)
	if err != nil {
		return nil, err
	}
	menteeCharacterName, err := s.getRewardDistributionCharacterName(menteeUser.PrimaryCharacterID)
	if err != nil {
		return nil, err
	}

	return &mentorRewardDistributionSnapshot{
		mentorCharacterName: mentorCharacterName,
		mentorNickname:      mentorUser.Nickname,
		menteeCharacterName: menteeCharacterName,
		menteeNickname:      menteeUser.Nickname,
	}, nil
}

func (s *MentorRewardService) getRewardDistributionCharacterName(characterID int64) (string, error) {
	if characterID == 0 {
		return "", nil
	}

	character, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return character.CharacterName, nil
}

func (s *MentorRewardService) distributeStageRewardTx(
	tx *gorm.DB,
	rel model.MentorMenteeRelationship,
	stage model.MentorRewardStage,
	snapshot *mentorRewardDistributionSnapshot,
	now time.Time,
) error {
	walletRefID := fmt.Sprintf("mentor_reward:%d:%d:%d", rel.ID, stage.StageOrder, now.Unix())
	reason := fmt.Sprintf("导师奖励 关系#%d 阶段#%d %s 学员#%d", rel.ID, stage.StageOrder, stage.Name, rel.MenteeUserID)

	dist := &model.MentorRewardDistribution{
		RelationshipID:      rel.ID,
		StageID:             stage.ID,
		StageOrder:          stage.StageOrder,
		MentorUserID:        rel.MentorUserID,
		MentorCharacterName: snapshot.mentorCharacterName,
		MentorNickname:      snapshot.mentorNickname,
		MenteeUserID:        rel.MenteeUserID,
		MenteeCharacterName: snapshot.menteeCharacterName,
		MenteeNickname:      snapshot.menteeNickname,
		RewardAmount:        stage.RewardAmount,
		DistributedAt:       now,
		WalletRefID:         walletRefID,
	}
	if err := s.distRepo.CreateTx(tx, dist); err != nil {
		return err
	}
	return s.walletSvc.ApplyWalletDeltaTx(tx, rel.MentorUserID, stage.RewardAmount, reason, model.WalletRefMentorReward, walletRefID)
}
