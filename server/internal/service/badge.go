package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"sync"
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
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	// 1. 可申请福利
	wg.Add(1)
	go func() {
		defer wg.Done()
		welfareEligible, err := s.countEligibleWelfares(userID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			if firstErr == nil {
				firstErr = errors.New("获取可申请福利数量失败")
			}
			return
		}
		if welfareEligible > 0 {
			counts[BadgeCountWelfareEligible] = welfareEligible
		}
	}()

	// 2. SRP 待审批
	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleSRP, model.RoleFC) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pending, err := s.srpRepo.CountPendingBadgeApplications()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = errors.New("获取补损待审批数量失败")
				}
				return
			}
			if pending > 0 {
				counts[BadgeCountSrpPending] = pending
			}
		}()
	}

	// 3. 福利待审批
	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleWelfare) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pending, err := s.welfareRepo.CountPendingBadgeApplications()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = errors.New("获取福利待审批数量失败")
				}
				return
			}
			if pending > 0 {
				counts[BadgeCountWelfarePending] = pending
			}
		}()
	}

	// 4. 商店订单
	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleWelfare) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pending, err := s.shopRepo.CountPendingOrders()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = errors.New("获取商店订单待处理数量失败")
				}
				return
			}
			if pending > 0 {
				counts[BadgeCountOrderPending] = pending
			}
		}()
	}

	// 5. 导师待处理
	if model.ContainsAnyRole(userRoles, model.RoleMentor) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pending, err := s.mentorRepo.CountPendingByMentorUserID(userID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = errors.New("获取导师待处理申请数量失败")
				}
				return
			}
			if pending > 0 {
				counts[BadgeCountMentorPendingApplications] = pending
			}
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
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
