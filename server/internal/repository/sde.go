package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

var TC_ID = map[string]int{
	"type":          8,
	"group":         7,
	"category":      6,
	"description":   33,
	"tech":          34,
	"market_group":  36,
	"solar_system":  40,
	"constellation": 41,
	"region":        42,
}

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
func (r *SdeRepository) GetTranslation(tcID int, keyID int, languageID string) (*model.TrnTranslation, error) {
	var t model.TrnTranslation
	err := global.DB.Where("tcID = ? AND keyID = ? AND languageID = ?", tcID, keyID, languageID).First(&t).Error
	return &t, err
}

// ---- invTypes 查询 ----
type TypeInfo struct {
	TypeID          int    `json:"type_id"`
	TypeName        string `json:"type_name"`
	GroupID         int    `json:"group_id"`
	GroupName       string `json:"group_name"`
	MarketGroupID   int    `json:"market_group_id"`
	MarketGroupName string `json:"market_group_name"`
	CategoryID      int    `json:"category_id"`
	CategoryName    string `json:"category_name"`
}

// GetNames 批量查询 id -> name 映射，ids 为 tcID key -> id列表
func (r *SdeRepository) GetNames(ids map[string][]int, languageID string) (map[int]string, error) {
	result := make(map[int]string)

	type row struct {
		KeyID int    `gorm:"column:keyID"`
		Text  string `gorm:"column:text"`
	}

	for key, keyIDs := range ids {
		if len(keyIDs) == 0 {
			continue
		}
		tcID, ok := TC_ID[key]
		if !ok {
			continue // 忽略不存在的 key
		}
		var rows []row
		if err := global.DB.Table("trnTranslations").
			Select("keyID, text").
			Where("tcID = ? AND keyID IN ? AND languageID = ?", tcID, keyIDs, languageID).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			result[r.KeyID] = r.Text
		}
	}

	return result, nil
}

// GetTypes 查询物品(组)信息
func (r *SdeRepository) GetTypes(typeIDs []int, published *bool, languageID string) ([]TypeInfo, error) {
	var result []TypeInfo

	query := global.DB.Table("invTypes t").
		Select(`
            t.typeID        AS type_id,
            t_name.text     AS type_name,
            g.groupID       AS group_id,
            g_name.text     AS group_name,
            t.marketGroupID AS market_group_id,
            mg_name.text    AS market_group_name,
            c.categoryID    AS category_id,
            c_name.text     AS category_name
        `).
		// invTypes -> invGroups
		Joins("LEFT JOIN invGroups g ON g.groupID = t.groupID").
		// invGroups -> invCategories
		Joins("LEFT JOIN invCategories c ON c.categoryID = g.categoryID").
		// invGroups -> invMarketGroups
		Joins("LEFT JOIN invMarketGroups mg ON mg.marketGroupID = t.marketGroupID").
		// 物品名翻译
		Joins("LEFT JOIN trnTranslations t_name ON t_name.tcID = ? AND t_name.keyID = t.typeID AND t_name.languageID = ?",
			TC_ID["type"], languageID).
		// 组名翻译
		Joins("LEFT JOIN trnTranslations g_name ON g_name.tcID = ? AND g_name.keyID = g.groupID AND g_name.languageID = ?",
			TC_ID["group"], languageID).
		// 分类名翻译
		Joins("LEFT JOIN trnTranslations c_name ON c_name.tcID = ? AND c_name.keyID = c.categoryID AND c_name.languageID = ?",
			TC_ID["category"], languageID).
		// 市场组名翻译
		Joins("LEFT JOIN trnTranslations mg_name ON mg_name.tcID = ? AND mg_name.keyID = mg.marketGroupID AND mg_name.languageID = ?",
			TC_ID["market_group"], languageID)

	if len(typeIDs) > 0 {
		query = query.Where("t.typeID IN ?", typeIDs)
	}
	if published != nil && *published {
		query = query.Where("t.published = ?", 1)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
