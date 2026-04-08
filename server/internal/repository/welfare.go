package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WelfareRepository 福利数据访问层
type WelfareRepository struct{}

func NewWelfareRepository() *WelfareRepository {
	return &WelfareRepository{}
}

func buildPendingBadgeWelfareApplicationCountQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.WelfareApplication{}).
		Where("status = ?", model.WelfareAppStatusRequested)
}

func (r *WelfareRepository) CountPendingBadgeApplications() (int64, error) {
	var count int64
	err := buildPendingBadgeWelfareApplicationCountQuery(global.DB).Count(&count).Error
	return count, err
}

// ─────────────────────────────────────────────
//  福利定义
// ─────────────────────────────────────────────

// CreateWelfare 创建福利
func (r *WelfareRepository) CreateWelfare(w *model.Welfare) error {
	return global.DB.Create(w).Error
}

// UpdateWelfare 更新福利
func (r *WelfareRepository) UpdateWelfare(w *model.Welfare) error {
	return global.DB.Save(w).Error
}

// DeleteWelfare 删除福利（软删除）
func (r *WelfareRepository) DeleteWelfare(id uint) error {
	return global.DB.Delete(&model.Welfare{}, id).Error
}

// GetWelfareByID 根据 ID 获取福利
func (r *WelfareRepository) GetWelfareByID(id uint) (*model.Welfare, error) {
	var w model.Welfare
	if err := global.DB.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// GetWelfareByIDTx 根据 ID 在指定事务中获取福利
func (r *WelfareRepository) GetWelfareByIDTx(tx *gorm.DB, id uint) (*model.Welfare, error) {
	var w model.Welfare
	if err := tx.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WelfareRepository) ListWelfaresByIDs(ids []uint) ([]model.Welfare, error) {
	var list []model.Welfare
	if len(ids) == 0 {
		return list, nil
	}
	err := global.DB.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// WelfareFilter 福利查询筛选
type WelfareFilter struct {
	Status *int8
	Name   string
}

// ListWelfares 分页查询福利
func (r *WelfareRepository) ListWelfares(page, pageSize int, filter WelfareFilter) ([]model.Welfare, int64, error) {
	var list []model.Welfare
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.Welfare{})
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Name != "" {
		db = db.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("sort_order ASC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	// 填充 SkillPlanIDs
	if err := r.fillSkillPlanIDs(list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// WelfareSortUpdate 用于批量更新排序字段
type WelfareSortUpdate struct {
	ID        uint
	SortOrder int
}

// UpdateWelfareSortOrders 批量更新福利的排序
func (r *WelfareRepository) UpdateWelfareSortOrders(updates []WelfareSortUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	tx := global.DB.Begin()
	for _, u := range updates {
		if err := tx.Model(&model.Welfare{}).Where("id = ?", u.ID).Update("sort_order", u.SortOrder).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// ─────────────────────────────────────────────
//  福利-技能计划关联
// ─────────────────────────────────────────────

// ReplaceWelfareSkillPlans 替换福利的技能计划关联
func (r *WelfareRepository) ReplaceWelfareSkillPlans(welfareID uint, skillPlanIDs []uint) error {
	tx := global.DB.Begin()
	if err := tx.Where("welfare_id = ?", welfareID).Delete(&model.WelfareSkillPlan{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(skillPlanIDs) > 0 {
		rows := make([]model.WelfareSkillPlan, 0, len(skillPlanIDs))
		for _, spID := range skillPlanIDs {
			rows = append(rows, model.WelfareSkillPlan{WelfareID: welfareID, SkillPlanID: spID})
		}
		if err := tx.Create(&rows).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetSkillPlanIDsByWelfareID 获取福利关联的技能计划 ID 列表
func (r *WelfareRepository) GetSkillPlanIDsByWelfareID(welfareID uint) ([]uint, error) {
	var rows []model.WelfareSkillPlan
	if err := global.DB.Where("welfare_id = ?", welfareID).Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.SkillPlanID
	}
	return ids, nil
}

// fillSkillPlanIDs 批量填充 Welfare.SkillPlanIDs
func (r *WelfareRepository) fillSkillPlanIDs(list []model.Welfare) error {
	if len(list) == 0 {
		return nil
	}
	welfareIDs := make([]uint, len(list))
	for i, w := range list {
		welfareIDs[i] = w.ID
	}
	var rows []model.WelfareSkillPlan
	if err := global.DB.Where("welfare_id IN ?", welfareIDs).Find(&rows).Error; err != nil {
		return err
	}

	m := make(map[uint][]uint)
	for _, row := range rows {
		m[row.WelfareID] = append(m[row.WelfareID], row.SkillPlanID)
	}
	for i := range list {
		list[i].SkillPlanIDs = m[list[i].ID]
		if list[i].SkillPlanIDs == nil {
			list[i].SkillPlanIDs = []uint{}
		}
	}
	return nil
}

// ─────────────────────────────────────────────
//  福利申请记录
// ─────────────────────────────────────────────

// CountApplicationsByWelfareID 统计福利的申请记录数
func (r *WelfareRepository) CountApplicationsByWelfareID(welfareID uint) (int64, error) {
	var count int64
	err := global.DB.Model(&model.WelfareApplication{}).Where("welfare_id = ?", welfareID).Count(&count).Error
	return count, err
}

// CreateApplication 创建福利申请
func (r *WelfareRepository) CreateApplication(app *model.WelfareApplication) error {
	return global.DB.Create(app).Error
}

// CreateApplicationTx 在事务中创建福利申请
func (r *WelfareRepository) CreateApplicationTx(tx *gorm.DB, app *model.WelfareApplication) error {
	return tx.Create(app).Error
}

// BulkCreateApplications 批量创建福利申请记录
func (r *WelfareRepository) BulkCreateApplications(apps []model.WelfareApplication) error {
	if len(apps) == 0 {
		return nil
	}
	return global.DB.Create(&apps).Error
}

// ListApplicationsByUserID 查询用户的所有福利申请
func (r *WelfareRepository) ListApplicationsByUserID(userID uint, status string) ([]model.WelfareApplication, error) {
	var list []model.WelfareApplication
	db := global.DB.Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Order("id DESC").Find(&list).Error
	return list, err
}

// buildApplicationsByUserIDQuery 构建用户福利申请查询条件
func buildApplicationsByUserIDQuery(db *gorm.DB, userID uint, status string) *gorm.DB {
	query := db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	return query
}

// ListApplicationsByUserIDPaginated 分页查询用户的福利申请
func (r *WelfareRepository) ListApplicationsByUserIDPaginated(
	userID uint,
	page, pageSize int,
	status string,
) ([]model.WelfareApplication, int64, error) {
	var list []model.WelfareApplication
	var total int64
	offset := (page - 1) * pageSize

	db := buildApplicationsByUserIDQuery(global.DB.Model(&model.WelfareApplication{}), userID, status)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListApplicationsByWelfareIDs 查询多个福利的所有申请记录（用于批量判断资格）
func (r *WelfareRepository) ListApplicationsByWelfareIDs(welfareIDs []uint) ([]model.WelfareApplication, error) {
	var list []model.WelfareApplication
	if len(welfareIDs) == 0 {
		return list, nil
	}
	err := global.DB.Where("welfare_id IN ?", welfareIDs).Find(&list).Error
	return list, err
}

// WelfareApplicationFilter 福利申请查询筛选
type WelfareApplicationFilter struct {
	Status   string   // 单状态精确匹配
	StatusIn []string // 多状态匹配
	Keyword  string   // 匹配申请人昵称、人物名或 QQ（不区分大小写）
}

// ListApplicationsPaginated 分页查询所有福利申请（管理端）
func (r *WelfareRepository) ListApplicationsPaginated(page, pageSize int, filter WelfareApplicationFilter) ([]model.WelfareApplication, int64, error) {
	var list []model.WelfareApplication
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.WelfareApplication{})
	if filter.Status != "" {
		db = db.Where("welfare_application.status = ?", filter.Status)
	} else if len(filter.StatusIn) > 0 {
		db = db.Where("welfare_application.status IN ?", filter.StatusIn)
	}
	if filter.Keyword != "" {
		db = db.Joins(`LEFT JOIN "user" AS applicant_user ON applicant_user.id = welfare_application.user_id`)
		db = applyKeywordLikeFilter(
			db,
			filter.Keyword,
			`LOWER(applicant_user.nickname) LIKE ?`,
			`LOWER(welfare_application.character_name) LIKE ?`,
			`LOWER(welfare_application.qq) LIKE ?`)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("welfare_application.id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetApplicationByID 根据 ID 获取福利申请
func (r *WelfareRepository) GetApplicationByID(id uint) (*model.WelfareApplication, error) {
	var app model.WelfareApplication
	if err := global.DB.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func buildGetApplicationByIDForUpdateQuery(db *gorm.DB, id uint) *gorm.DB {
	return db.Model(&model.WelfareApplication{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("welfare_application.id = ?", id)
}

// GetApplicationByIDForUpdateTx 根据 ID 在事务中获取福利申请并加行锁
func (r *WelfareRepository) GetApplicationByIDForUpdateTx(tx *gorm.DB, id uint) (*model.WelfareApplication, error) {
	var app model.WelfareApplication
	if err := buildGetApplicationByIDForUpdateQuery(tx, id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApplication 更新福利申请
func (r *WelfareRepository) UpdateApplication(app *model.WelfareApplication) error {
	return global.DB.Save(app).Error
}

// UpdateApplicationTx 在事务中更新福利申请
func (r *WelfareRepository) UpdateApplicationTx(tx *gorm.DB, app *model.WelfareApplication) error {
	return tx.Save(app).Error
}

// DeleteApplication 删除福利申请记录
func (r *WelfareRepository) DeleteApplication(id uint) error {
	return global.DB.Delete(&model.WelfareApplication{}, id).Error
}

// ListActiveWelfares 查询所有启用的福利
func (r *WelfareRepository) ListActiveWelfares() ([]model.Welfare, error) {
	var list []model.Welfare
	if err := global.DB.Where("status = ?", model.WelfareStatusActive).Order("sort_order ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	if err := r.fillSkillPlanIDs(list); err != nil {
		return nil, err
	}
	return list, nil
}
