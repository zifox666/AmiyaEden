package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// LotteryService 抽奖业务逻辑层
type LotteryService struct {
	repo      *repository.LotteryRepository
	walletSvc *SysWalletService
}

func NewLotteryService() *LotteryService {
	return &LotteryService{
		repo:      repository.NewLotteryRepository(),
		walletSvc: NewSysWalletService(),
	}
}

// ─────────────────────────────────────────────
//  用户端
// ─────────────────────────────────────────────

// ListActiveActivities 获取用户可见的抽奖活动列表（含奖品）
func (s *LotteryService) ListActiveActivities(page, pageSize int) ([]model.ShopLotteryActivity, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.repo.ListActivities(page, pageSize, false)
}

// DrawResult 抽奖结果
type DrawResult struct {
	Prize model.ShopLotteryPrize `json:"prize"`
}

// Draw 用户抽奖
func (s *LotteryService) Draw(userID uint, activityID uint) (*DrawResult, error) {
	// 1. 获取活动（含奖品）
	activity, err := s.repo.GetActivityByID(activityID)
	if err != nil {
		return nil, errors.New("抽奖活动不存在")
	}

	// 2. 检查活动状态和时间
	if activity.Status != model.LotteryStatusActive {
		return nil, errors.New("抽奖活动已关闭")
	}
	now := time.Now()
	if activity.StartAt != nil && now.Before(*activity.StartAt) {
		return nil, errors.New("抽奖活动尚未开始")
	}
	if activity.EndAt != nil && now.After(*activity.EndAt) {
		return nil, errors.New("抽奖活动已结束")
	}

	// 3. 过滤可用奖品（排除库存耗尽的奖品）
	var availablePrizes []model.ShopLotteryPrize
	for _, p := range activity.Prizes {
		if p.TotalStock <= 0 || p.DrawnCount < p.TotalStock {
			availablePrizes = append(availablePrizes, p)
		}
	}
	if len(availablePrizes) == 0 {
		return nil, errors.New("该活动奖品已全部抽完")
	}

	// 4. 检查并扣除费用
	if activity.CostPerDraw > 0 {
		wallet, err := s.walletSvc.GetMyWallet(userID)
		if err != nil {
			return nil, fmt.Errorf("获取钱包失败: %w", err)
		}
		if wallet.Balance < activity.CostPerDraw {
			return nil, errors.New("余额不足")
		}
		refID := fmt.Sprintf("lottery:%d:user:%d", activityID, userID)
		reason := fmt.Sprintf("抽奖: %s", activity.Name)
		if err := s.walletSvc.DebitUser(userID, activity.CostPerDraw, reason, model.WalletRefLotteryDraw, refID); err != nil {
			return nil, fmt.Errorf("扣款失败: %w", err)
		}
	}

	// 5. 加权随机抽奖（仅从可用奖品中选择）
	prize := weightedRandom(availablePrizes)

	// 6. 递增奖品已抽出数量
	_ = s.repo.IncrementPrizeDrawnCount(prize.ID)

	// 7. 记录抽奖结果
	record := &model.ShopLotteryRecord{
		UserID:       userID,
		ActivityID:   activityID,
		ActivityName: activity.Name,
		PrizeID:      prize.ID,
		PrizeName:    prize.Name,
		PrizeTier:    prize.Tier,
		PrizeImage:   prize.Image,
		Cost:         activity.CostPerDraw,
	}
	_ = s.repo.CreateRecord(record)

	return &DrawResult{
		Prize: prize,
	}, nil
}

// GetMyLotteryRecords 获取我的抽奖记录
func (s *LotteryService) GetMyLotteryRecords(userID uint, page, pageSize int) ([]model.ShopLotteryRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	uid := userID
	return s.repo.ListRecords(page, pageSize, &uid, nil)
}

// ─────────────────────────────────────────────
//  管理员端
// ─────────────────────────────────────────────

// AdminListActivities 管理员查询所有活动
func (s *LotteryService) AdminListActivities(page, pageSize int) ([]model.ShopLotteryActivity, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListActivities(page, pageSize, true)
}

// AdminCreateActivity 创建抽奖活动
func (s *LotteryService) AdminCreateActivity(a *model.ShopLotteryActivity) error {
	return s.repo.CreateActivity(a)
}

// AdminUpdateActivity 更新抽奖活动
func (s *LotteryService) AdminUpdateActivity(id uint, req *AdminLotteryActivityUpdateRequest) (*model.ShopLotteryActivity, error) {
	a, err := s.repo.GetActivityByID(id)
	if err != nil {
		return nil, errors.New("活动不存在")
	}
	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	if req.Image != nil {
		a.Image = *req.Image
	}
	if req.CostPerDraw != nil {
		a.CostPerDraw = *req.CostPerDraw
	}
	if req.Status != nil {
		a.Status = *req.Status
	}
	if req.StartAt != nil {
		a.StartAt = req.StartAt
	}
	if req.EndAt != nil {
		a.EndAt = req.EndAt
	}
	if req.SortOrder != nil {
		a.SortOrder = *req.SortOrder
	}
	if err := s.repo.UpdateActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}

// AdminLotteryActivityUpdateRequest 活动更新字段
type AdminLotteryActivityUpdateRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Image       *string    `json:"image"`
	CostPerDraw *float64   `json:"cost_per_draw"`
	Status      *int8      `json:"status"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	SortOrder   *int       `json:"sort_order"`
}

// AdminDeleteActivity 删除抽奖活动
func (s *LotteryService) AdminDeleteActivity(id uint) error {
	return s.repo.DeleteActivity(id)
}

// AdminCreatePrize 添加奖品到活动
func (s *LotteryService) AdminCreatePrize(p *model.ShopLotteryPrize) error {
	if p.ProbabilityWeight < 1 {
		p.ProbabilityWeight = 1
	}
	if p.TotalStock < 0 {
		p.TotalStock = 0
	}
	return s.repo.CreatePrize(p)
}

// AdminUpdatePrize 更新奖品
func (s *LotteryService) AdminUpdatePrize(id uint, req *AdminLotteryPrizeUpdateRequest) (*model.ShopLotteryPrize, error) {
	p, err := s.repo.GetPrizeByID(id)
	if err != nil {
		return nil, errors.New("奖品不存在")
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Image != nil {
		p.Image = *req.Image
	}
	if req.Tier != nil {
		p.Tier = *req.Tier
	}
	if req.ProbabilityWeight != nil {
		p.ProbabilityWeight = *req.ProbabilityWeight
	}
	if req.TotalStock != nil {
		p.TotalStock = *req.TotalStock
	}
	if err := s.repo.UpdatePrize(p); err != nil {
		return nil, err
	}
	return p, nil
}

// AdminLotteryPrizeUpdateRequest 奖品更新字段
type AdminLotteryPrizeUpdateRequest struct {
	ActivityID        uint    `json:"activity_id"`
	Name              *string `json:"name"`
	Image             *string `json:"image"`
	Tier              *string `json:"tier"`
	ProbabilityWeight *int    `json:"probability_weight"`
	TotalStock        *int    `json:"total_stock"`
}

// AdminDeletePrize 删除奖品
func (s *LotteryService) AdminDeletePrize(id uint) error {
	return s.repo.DeletePrize(id)
}

// AdminUpdateRecordDelivery 更新抽奖记录发放状态
func (s *LotteryService) AdminUpdateRecordDelivery(id uint, status string) error {
	if status != model.LotteryDeliveryPending && status != model.LotteryDeliveryDelivered {
		return errors.New("无效的发放状态")
	}
	return s.repo.UpdateRecordDeliveryStatus(id, status)
}

// AdminListRecords 管理员查询抽奖记录
func (s *LotteryService) AdminListRecords(page, pageSize int, activityID *uint) ([]model.ShopLotteryRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListRecords(page, pageSize, nil, activityID)
}

// ─────────────────────────────────────────────
//  内部工具
// ─────────────────────────────────────────────

// weightedRandom 根据 ProbabilityWeight 进行加权随机选择
func weightedRandom(prizes []model.ShopLotteryPrize) model.ShopLotteryPrize {
	total := 0
	for _, p := range prizes {
		w := p.ProbabilityWeight
		if w < 1 {
			w = 1
		}
		total += w
	}
	r := rand.Intn(total)
	cumulative := 0
	for _, p := range prizes {
		w := p.ProbabilityWeight
		if w < 1 {
			w = 1
		}
		cumulative += w
		if r < cumulative {
			return p
		}
	}
	return prizes[len(prizes)-1]
}
