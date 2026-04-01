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
		result.ProcessedRelationships++
		metrics, err := s.getMenteeMetrics(relationship.MenteeUserID)
		if err != nil {
			global.Logger.Error("导师奖励处理：获取学员指标失败", zap.Uint("mentee_user_id", relationship.MenteeUserID), zap.Error(err))
			continue
		}

		allDistributed := true
		for _, stage := range stages {
			exists, err := s.distRepo.ExistsByRelationshipAndStageOrder(relationship.ID, stage.StageOrder)
			if err != nil {
				global.Logger.Error("导师奖励处理：检查奖励记录失败", zap.Uint("relationship_id", relationship.ID), zap.Int("stage_order", stage.StageOrder), zap.Error(err))
				allDistributed = false
				break
			}
			if exists {
				continue
			}
			if !isMentorConditionMet(stage, metrics) {
				allDistributed = false
				break
			}
			if err := s.distributeStageReward(relationship, stage, now); err != nil {
				global.Logger.Error("导师奖励处理：发放奖励失败", zap.Uint("relationship_id", relationship.ID), zap.Int("stage_order", stage.StageOrder), zap.Error(err))
				allDistributed = false
				break
			}
			result.RewardsDistributed++
			result.TotalCoinAwarded += stage.RewardAmount
		}

		if allDistributed {
			if err := s.relRepo.UpdateStatus(relationship.ID, model.MentorRelationStatusGraduated, map[string]any{"graduated_at": now}); err != nil {
				global.Logger.Error("导师奖励处理：标记毕业失败", zap.Uint("relationship_id", relationship.ID), zap.Error(err))
			} else {
				result.GraduatedCount++
			}
		}
	}

	return result, nil
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

func (s *MentorRewardService) distributeStageReward(rel model.MentorMenteeRelationship, stage model.MentorRewardStage, now time.Time) error {
	walletRefID := fmt.Sprintf("mentor_reward:%d:%d:%d", rel.ID, stage.StageOrder, now.Unix())
	reason := fmt.Sprintf("导师奖励 关系#%d 阶段#%d %s 学员#%d", rel.ID, stage.StageOrder, stage.Name, rel.MenteeUserID)

	return global.DB.Transaction(func(tx *gorm.DB) error {
		dist := &model.MentorRewardDistribution{
			RelationshipID: rel.ID,
			StageID:        stage.ID,
			StageOrder:     stage.StageOrder,
			MentorUserID:   rel.MentorUserID,
			MenteeUserID:   rel.MenteeUserID,
			RewardAmount:   stage.RewardAmount,
			DistributedAt:  now,
			WalletRefID:    walletRefID,
		}
		if err := s.distRepo.CreateTx(tx, dist); err != nil {
			return err
		}
		return s.walletSvc.ApplyWalletDeltaTx(tx, rel.MentorUserID, stage.RewardAmount, reason, model.WalletRefMentorReward, walletRefID)
	})
}
