package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
)

const (
	BadgeCountWelfareEligible           = "welfare_eligible"
	BadgeCountSrpPending                = "srp_pending"
	BadgeCountWelfarePending            = "welfare_pending"
	BadgeCountOrderPending              = "order_pending"
	BadgeCountMentorPendingApplications = "mentor_pending_applications"
)

type BadgeCounts map[string]int64

type BadgeService struct {
	welfareSvc  *WelfareService
	srpRepo     *repository.SrpRepository
	welfareRepo *repository.WelfareRepository
	shopRepo    *repository.ShopRepository
	mentorRepo  *repository.MentorRelationshipRepository
}

func NewBadgeService() *BadgeService {
	return &BadgeService{
		welfareSvc:  NewWelfareService(),
		srpRepo:     repository.NewSrpRepository(),
		welfareRepo: repository.NewWelfareRepository(),
		shopRepo:    repository.NewShopRepository(),
		mentorRepo:  repository.NewMentorRelationshipRepository(),
	}
}

func (s *BadgeService) GetBadgeCounts(userID uint, userRoles []string) (BadgeCounts, error) {
	counts := BadgeCounts{}

	welfareEligible, err := s.countEligibleWelfares(userID)
	if err != nil {
		return nil, errors.New("获取可申请福利数量失败")
	}
	if welfareEligible > 0 {
		counts[BadgeCountWelfareEligible] = welfareEligible
	}

	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleSRP, model.RoleFC) {
		pending, err := s.srpRepo.CountPendingBadgeApplications()
		if err != nil {
			return nil, errors.New("获取补损待审批数量失败")
		}
		if pending > 0 {
			counts[BadgeCountSrpPending] = pending
		}
	}

	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleWelfare) {
		pending, err := s.welfareRepo.CountPendingBadgeApplications()
		if err != nil {
			return nil, errors.New("获取福利待审批数量失败")
		}
		if pending > 0 {
			counts[BadgeCountWelfarePending] = pending
		}
	}

	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleWelfare) {
		pending, err := s.shopRepo.CountPendingOrders()
		if err != nil {
			return nil, errors.New("获取商店订单待处理数量失败")
		}
		if pending > 0 {
			counts[BadgeCountOrderPending] = pending
		}
	}

	if model.ContainsAnyRole(userRoles, model.RoleMentor) {
		pending, err := s.mentorRepo.CountPendingByMentorUserID(userID)
		if err != nil {
			return nil, errors.New("获取导师待处理申请数量失败")
		}
		if pending > 0 {
			counts[BadgeCountMentorPendingApplications] = pending
		}
	}

	return counts, nil
}

func (s *BadgeService) countEligibleWelfares(userID uint) (int64, error) {
	eligibleWelfares, err := s.welfareSvc.GetEligibleWelfares(userID)
	if err != nil {
		return 0, err
	}

	var count int64
	for _, welfare := range eligibleWelfares {
		if welfare.DistMode == model.WelfareDistModePerCharacter {
			for _, character := range welfare.EligibleCharacters {
				if character.CanApplyNow {
					count++
					break
				}
			}
			continue
		}

		if welfare.CanApplyNow {
			count++
		}
	}

	return count, nil
}
