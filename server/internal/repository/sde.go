package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// SdeRepository SDE 数据访问层
type SdeRepository struct{}

func NewSdeRepository() *SdeRepository { return &SdeRepository{} }

// ---- SDE 版本管理 ----

// GetLatestVersion 获取最新已导入的 SDE 版本
func (r *SdeRepository) GetLatestVersion() (*model.SdeVersion, error) {
	var v model.SdeVersion
	err := global.DB.Order("id DESC").First(&v).Error
	return &v, err
}

// CreateVersion 记录新版本
func (r *SdeRepository) CreateVersion(v *model.SdeVersion) error {
	return global.DB.Create(v).Error
}

// VersionExists 检查某个版本是否已存在
func (r *SdeRepository) VersionExists(version string) (bool, error) {
	var count int64
	err := global.DB.Model(&model.SdeVersion{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

// ---- trnTranslations 翻译查询 ----

// GetTranslation 精确查询翻译
func (r *SdeRepository) GetTranslation(tcID, keyID int, languageID string) (*model.TrnTranslation, error) {
	var t model.TrnTranslation
	err := global.DB.Where("tcID = ? AND keyID = ? AND languageID = ?", tcID, keyID, languageID).First(&t).Error
	return &t, err
}

// GetTranslationsByKey 查询某个 key 全部语言的翻译
func (r *SdeRepository) GetTranslationsByKey(tcID, keyID int) ([]model.TrnTranslation, error) {
	var list []model.TrnTranslation
	err := global.DB.Where("tcID = ? AND keyID = ?", tcID, keyID).Find(&list).Error
	return list, err
}

// ---- invTypes 查询 ----

// GetTypeByID 根据 typeID 查询物品信息
func (r *SdeRepository) GetTypeByID(typeID int) (*model.InvType, error) {
	var t model.InvType
	err := global.DB.Where("typeID = ?", typeID).First(&t).Error
	return &t, err
}

// FuzzySearchTypesByName 模糊搜索物品名称（英文名称，忽略大小写）
func (r *SdeRepository) FuzzySearchTypesByName(keyword string, limit int) ([]model.InvType, error) {
	var list []model.InvType
	err := global.DB.Where("typeName LIKE ?", "%"+keyword+"%").
		Limit(limit).Find(&list).Error
	return list, err
}

// ---- invGroups 查询 ----

// GetGroupByID 根据 groupID 查询组信息
func (r *SdeRepository) GetGroupByID(groupID int) (*model.InvGroup, error) {
	var g model.InvGroup
	err := global.DB.Where("groupID = ?", groupID).First(&g).Error
	return &g, err
}

// ---- invCategories 查询 ----

// GetCategoryByID 根据 categoryID 查询分类信息
func (r *SdeRepository) GetCategoryByID(categoryID int) (*model.InvCategory, error) {
	var c model.InvCategory
	err := global.DB.Where("categoryID = ?", categoryID).First(&c).Error
	return &c, err
}
