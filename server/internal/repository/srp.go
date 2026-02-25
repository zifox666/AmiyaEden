package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// SrpRepository SRP 数据访问层
type SrpRepository struct{}

func NewSrpRepository() *SrpRepository {
	return &SrpRepository{}
}

// ─────────────────────────────────────────────
//  SrpShipPrice CRUD
// ─────────────────────────────────────────────

// GetShipPriceByTypeID 按 ship_type_id 查找
func (r *SrpRepository) GetShipPriceByTypeID(shipTypeID int64) (*model.SrpShipPrice, error) {
	var p model.SrpShipPrice
	err := global.DB.Where("ship_type_id = ?", shipTypeID).First(&p).Error
	return &p, err
}

// ListShipPrices 查询所有舰船价格（可按名称模糊搜索）
func (r *SrpRepository) ListShipPrices(keyword string) ([]model.SrpShipPrice, error) {
	var list []model.SrpShipPrice
	db := global.DB.Model(&model.SrpShipPrice{})
	if keyword != "" {
		db = db.Where("ship_name LIKE ?", "%"+keyword+"%")
	}
	err := db.Order("ship_type_id ASC").Find(&list).Error
	return list, err
}

// UpsertShipPrice 创建或更新舰船价格
func (r *SrpRepository) UpsertShipPrice(p *model.SrpShipPrice) error {
	return global.DB.Save(p).Error
}

// DeleteShipPrice 删除舰船价格
func (r *SrpRepository) DeleteShipPrice(id uint) error {
	return global.DB.Delete(&model.SrpShipPrice{}, id).Error
}

// ─────────────────────────────────────────────
//  SrpApplication CRUD
// ─────────────────────────────────────────────

// CreateApplication 创建补损申请
func (r *SrpRepository) CreateApplication(app *model.SrpApplication) error {
	return global.DB.Create(app).Error
}

// GetApplicationByID 按 ID 查询
func (r *SrpRepository) GetApplicationByID(id uint) (*model.SrpApplication, error) {
	var app model.SrpApplication
	err := global.DB.First(&app, id).Error
	return &app, err
}

// ExistsApplicationByKillmail 检查该 killmail 是否已被该角色提交过申请
func (r *SrpRepository) ExistsApplicationByKillmail(killmailID int64, characterID int64) bool {
	var count int64
	global.DB.Model(&model.SrpApplication{}).
		Where("killmail_id = ? AND character_id = ?", killmailID, characterID).
		Count(&count)
	return count > 0
}

// UpdateApplication 更新申请（审批 / 发放）
func (r *SrpRepository) UpdateApplication(app *model.SrpApplication) error {
	return global.DB.Save(app).Error
}

// SrpApplicationFilter 申请列表筛选条件
type SrpApplicationFilter struct {
	UserID       *uint
	CharacterID  *int64
	FleetID      *string
	ReviewStatus string
	PayoutStatus string
}

// ListApplications 分页查询申请列表
func (r *SrpRepository) ListApplications(page, pageSize int, filter SrpApplicationFilter) ([]model.SrpApplication, int64, error) {
	var list []model.SrpApplication
	var total int64

	db := global.DB.Model(&model.SrpApplication{})
	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.CharacterID != nil {
		db = db.Where("character_id = ?", *filter.CharacterID)
	}
	if filter.FleetID != nil {
		db = db.Where("fleet_id = ?", *filter.FleetID)
	}
	if filter.ReviewStatus != "" {
		db = db.Where("review_status = ?", filter.ReviewStatus)
	}
	if filter.PayoutStatus != "" {
		db = db.Where("payout_status = ?", filter.PayoutStatus)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListMyApplications 查询当前用户的申请（按用户 ID）
func (r *SrpRepository) ListMyApplications(userID uint, page, pageSize int) ([]model.SrpApplication, int64, error) {
	uid := &userID
	return r.ListApplications(page, pageSize, SrpApplicationFilter{UserID: uid})
}
