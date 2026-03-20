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
	err := global.DB.Where(`"tcID" = ? AND "keyID" = ? AND "languageID" = ?`, tcID, keyID, languageID).First(&t).Error
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
		if err := global.DB.Table(`"trnTranslations"`).
			Select(`"keyID", text`).
			Where(`"tcID" = ? AND "keyID" IN ? AND "languageID" = ?`, tcID, keyIDs, languageID).
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

	query := global.DB.Table(`"invTypes" t`).
		Select(`
            t."typeID"        AS type_id,
            t_name.text       AS type_name,
            g."groupID"       AS group_id,
            g_name.text       AS group_name,
            t."marketGroupID" AS market_group_id,
            mg_name.text      AS market_group_name,
            c."categoryID"    AS category_id,
            c_name.text       AS category_name
        `).
		// invTypes -> invGroups
		Joins(`LEFT JOIN "invGroups" g ON g."groupID" = t."groupID"`).
		// invGroups -> invCategories
		Joins(`LEFT JOIN "invCategories" c ON c."categoryID" = g."categoryID"`).
		// invGroups -> invMarketGroups
		Joins(`LEFT JOIN "invMarketGroups" mg ON mg."marketGroupID" = t."marketGroupID"`).
		// 物品名翻译
		Joins(`LEFT JOIN "trnTranslations" t_name ON t_name."tcID" = ? AND t_name."keyID" = t."typeID" AND t_name."languageID" = ?`,
			TC_ID["type"], languageID).
		// 组名翻译
		Joins(`LEFT JOIN "trnTranslations" g_name ON g_name."tcID" = ? AND g_name."keyID" = g."groupID" AND g_name."languageID" = ?`,
			TC_ID["group"], languageID).
		// 分类名翻译
		Joins(`LEFT JOIN "trnTranslations" c_name ON c_name."tcID" = ? AND c_name."keyID" = c."categoryID" AND c_name."languageID" = ?`,
			TC_ID["category"], languageID).
		// 市场组名翻译
		Joins(`LEFT JOIN "trnTranslations" mg_name ON mg_name."tcID" = ? AND mg_name."keyID" = mg."marketGroupID" AND mg_name."languageID" = ?`,
			TC_ID["market_group"], languageID)

	if len(typeIDs) > 0 {
		query = query.Where(`t."typeID" IN ?`, typeIDs)
	}
	if published != nil && *published {
		query = query.Where(`t.published = ?`, 1)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// GetTypesByCategoryID 获取指定 categoryID 下所有已发布的物品（含翻译）
// 主要用于技能模块获取 categoryID=16 的全量技能列表
func (r *SdeRepository) GetTypesByCategoryID(categoryID int, languageID string) ([]TypeInfo, error) {
	var result []TypeInfo

	query := global.DB.Table(`"invTypes" t`).
		Select(`
            t."typeID"        AS type_id,
            t_name.text       AS type_name,
            g."groupID"       AS group_id,
            g_name.text       AS group_name,
            t."marketGroupID" AS market_group_id,
            mg_name.text      AS market_group_name,
            c."categoryID"    AS category_id,
            c_name.text       AS category_name
        `).
		Joins(`LEFT JOIN "invGroups" g ON g."groupID" = t."groupID"`).
		Joins(`LEFT JOIN "invCategories" c ON c."categoryID" = g."categoryID"`).
		Joins(`LEFT JOIN "invMarketGroups" mg ON mg."marketGroupID" = t."marketGroupID"`).
		Joins(`LEFT JOIN "trnTranslations" t_name ON t_name."tcID" = ? AND t_name."keyID" = t."typeID" AND t_name."languageID" = ?`,
			TC_ID["type"], languageID).
		Joins(`LEFT JOIN "trnTranslations" g_name ON g_name."tcID" = ? AND g_name."keyID" = g."groupID" AND g_name."languageID" = ?`,
			TC_ID["group"], languageID).
		Joins(`LEFT JOIN "trnTranslations" c_name ON c_name."tcID" = ? AND c_name."keyID" = c."categoryID" AND c_name."languageID" = ?`,
			TC_ID["category"], languageID).
		Joins(`LEFT JOIN "trnTranslations" mg_name ON mg_name."tcID" = ? AND mg_name."keyID" = mg."marketGroupID" AND mg_name."languageID" = ?`,
			TC_ID["market_group"], languageID).
		Where(`c."categoryID" = ? AND t.published = 1`, categoryID)

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// FlagInfo invFlags 表行
type FlagInfo struct {
	FlagID   int    `json:"flag_id"   gorm:"column:flagID"`
	FlagName string `json:"flag_name" gorm:"column:flagName"`
	FlagText string `json:"flag_text" gorm:"column:flagText"`
	OrderID  int    `json:"order_id"  gorm:"column:orderID"`
}

// FuzzySearchItem 模糊搜索结果条目
type FuzzySearchItem struct {
	ID        int    `json:"id"         gorm:"column:id"`
	Name      string `json:"name"       gorm:"column:name"`
	GroupID   int    `json:"group_id"   gorm:"column:group_id"`
	GroupName string `json:"group_name" gorm:"column:group_name"`
	Category  string `json:"category"   gorm:"column:category"` // "type" | "character"
}

// GetTypeIDsByNames 通过英文名称批量反查 typeID（用于 EFT 解析）
func (r *SdeRepository) GetTypeIDsByNames(names []string) (map[string]int64, error) {
	if len(names) == 0 {
		return map[string]int64{}, nil
	}
	type row struct {
		TypeID   int64  `gorm:"column:typeID"`
		TypeName string `gorm:"column:typeName"`
	}
	var rows []row
	err := global.DB.Table(`"invTypes"`).
		Select(`"typeID", "typeName"`).
		Where(`"typeName" IN ?`, names).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.TypeName] = r.TypeID
	}
	return result, nil
}

// GetTypeIDsByTranslatedNames 通过翻译名称批量反查 typeID（用于技能规划解析）
func (r *SdeRepository) GetTypeIDsByTranslatedNames(names []string, languageID string) (map[string]int64, error) {
	if len(names) == 0 {
		return map[string]int64{}, nil
	}
	type row struct {
		TypeID int64  `gorm:"column:keyID"`
		Text   string `gorm:"column:text"`
	}
	var rows []row
	err := global.DB.Table(`"trnTranslations"`).
		Select(`"keyID", text`).
		Where(`"tcID" = ? AND "languageID" = ? AND text IN ?`, TC_ID["type"], languageID, names).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.Text] = r.TypeID
	}
	return result, nil
}

// FuzzySearch 模糊搜索 trnTranslations（物品名称）及成员名称
func (r *SdeRepository) FuzzySearch(keyword string, languageID string, categoryIDs []int, excludeCategoryIDs []int, limit int, searchMember bool) ([]FuzzySearchItem, error) {
	if keyword == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if languageID == "" {
		languageID = "en"
	}

	var result []FuzzySearchItem
	pattern := "%" + keyword + "%"

	// 1. 搜索 SDE 物品名称（通过 trnTranslations 的 type 翻译）
	query := global.DB.Table(`"trnTranslations" tr`).
		Select(`
			t."typeID"   AS id,
			tr.text      AS name,
			g."groupID"  AS group_id,
			g_name.text  AS group_name,
			'type'       AS category
		`).
		Joins(`JOIN "invTypes" t ON t."typeID" = tr."keyID" AND t.published = 1`).
		Joins(`JOIN "invGroups" g ON g."groupID" = t."groupID"`).
		Joins(`JOIN "invCategories" c ON c."categoryID" = g."categoryID"`).
		Joins(`LEFT JOIN "trnTranslations" g_name ON g_name."tcID" = ? AND g_name."keyID" = g."groupID" AND g_name."languageID" = ?`,
			TC_ID["group"], languageID).
		Where(`tr."tcID" = ? AND tr."languageID" = ? AND tr.text ILIKE ?`, TC_ID["type"], languageID, pattern)

	if len(categoryIDs) > 0 {
		query = query.Where(`c."categoryID" IN ?`, categoryIDs)
	}
	if len(excludeCategoryIDs) > 0 {
		query = query.Where(`c."categoryID" NOT IN ?`, excludeCategoryIDs)
	}

	query = query.Limit(limit)
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

	// 2. 搜索成员名称（eve_character 表）
	if searchMember && limit > len(result) {
		remaining := limit - len(result)
		var members []FuzzySearchItem
		if err := global.DB.Table("eve_character").
			Select(`character_id AS id, character_name AS name, 0 AS group_id, '' AS group_name, 'character' AS category`).
			Where("character_name ILIKE ?", pattern).
			Limit(remaining).
			Scan(&members).Error; err != nil {
			return nil, err
		}
		result = append(result, members...)
	}

	return result, nil
}

// GetFlags 批量查询 invFlags
func (r *SdeRepository) GetFlags(flagIDs []int) ([]FlagInfo, error) {
	var result []FlagInfo
	if len(flagIDs) == 0 {
		return result, nil
	}
	err := global.DB.Table(`"invFlags"`).
		Where(`"flagID" IN ?`, flagIDs).
		Order(`"orderID" ASC`).
		Scan(&result).Error
	return result, err
}

// ---- 舰船 & 技能需求查询 ----

// ShipInfo 舰船基础信息（含翻译 + raceID + marketGroupID）
type ShipInfo struct {
	TypeID          int    `json:"type_id"           gorm:"column:type_id"`
	TypeName        string `json:"type_name"         gorm:"column:type_name"`
	GroupID         int    `json:"group_id"           gorm:"column:group_id"`
	GroupName       string `json:"group_name"         gorm:"column:group_name"`
	MarketGroupID   int    `json:"market_group_id"    gorm:"column:market_group_id"`
	MarketGroupName string `json:"market_group_name"  gorm:"column:market_group_name"`
	RaceID          int    `json:"race_id"            gorm:"column:race_id"`
}

// GetShipsByCategoryID 获取 categoryID=6 的所有已发布舰船（含翻译、raceID、marketGroupID）
func (r *SdeRepository) GetShipsByCategoryID(languageID string) ([]ShipInfo, error) {
	const shipCategoryID = 6
	var result []ShipInfo

	err := global.DB.Table(`"invTypes" t`).
		Select(`
			t."typeID"        AS type_id,
			t_name.text       AS type_name,
			g."groupID"       AS group_id,
			g_name.text       AS group_name,
			t."marketGroupID" AS market_group_id,
			mg_name.text      AS market_group_name,
			t."raceID"        AS race_id
		`).
		Joins(`JOIN "invGroups" g ON g."groupID" = t."groupID"`).
		Joins(`JOIN "invCategories" c ON c."categoryID" = g."categoryID"`).
		Joins(`LEFT JOIN "invMarketGroups" mg ON mg."marketGroupID" = t."marketGroupID"`).
		Joins(`LEFT JOIN "trnTranslations" t_name ON t_name."tcID" = ? AND t_name."keyID" = t."typeID" AND t_name."languageID" = ?`,
			TC_ID["type"], languageID).
		Joins(`LEFT JOIN "trnTranslations" g_name ON g_name."tcID" = ? AND g_name."keyID" = g."groupID" AND g_name."languageID" = ?`,
			TC_ID["group"], languageID).
		Joins(`LEFT JOIN "trnTranslations" mg_name ON mg_name."tcID" = ? AND mg_name."keyID" = mg."marketGroupID" AND mg_name."languageID" = ?`,
			TC_ID["market_group"], languageID).
		Where(`c."categoryID" = ? AND t.published = 1`, shipCategoryID).
		Scan(&result).Error
	return result, err
}

// ShipSkillReq 舰船技能需求行（含递归层深）
type ShipSkillReq struct {
	ShipTypeID    int `json:"ship_type_id"    gorm:"column:ship_type_id"`
	SkillTypeID   int `json:"skill_type_id"   gorm:"column:skill_type_id"`
	RequiredLevel int `json:"required_level"  gorm:"column:required_level"`
	Depth         int `json:"depth"           gorm:"column:depth"`
}

// GetShipSkillRequirements 批量获取舰船技能需求（含前置技能递归）
func (r *SdeRepository) GetShipSkillRequirements(shipTypeIDs []int) ([]ShipSkillReq, error) {
	if len(shipTypeIDs) == 0 {
		return nil, nil
	}
	var result []ShipSkillReq

	sql := `
WITH RECURSIVE skill_tree AS (
  SELECT
    sk."typeID"   AS ship_type_id,
    sk."valueInt" AS skill_type_id,
    lv."valueInt" AS required_level,
    1             AS depth
  FROM "dgmTypeAttributes" sk
  JOIN "dgmTypeAttributes" lv
    ON sk."typeID" = lv."typeID"
    AND lv."attributeID" = CASE sk."attributeID"
      WHEN 182  THEN 277  WHEN 183 THEN 278  WHEN 184  THEN 279
      WHEN 1285 THEN 1286 WHEN 1289 THEN 1287 WHEN 1290 THEN 1288
    END
  WHERE sk."typeID" IN ?
    AND sk."attributeID" IN (182, 183, 184, 1285, 1289, 1290)
    AND sk."valueInt" IS NOT NULL

  UNION

  SELECT
    st.ship_type_id,
    sk."valueInt" AS skill_type_id,
    lv."valueInt" AS required_level,
    st.depth + 1
  FROM skill_tree st
  JOIN "dgmTypeAttributes" sk
    ON sk."typeID" = st.skill_type_id
    AND sk."attributeID" IN (182, 183, 184, 1285, 1289, 1290)
  JOIN "dgmTypeAttributes" lv
    ON sk."typeID" = lv."typeID"
    AND lv."attributeID" = CASE sk."attributeID"
      WHEN 182  THEN 277  WHEN 183 THEN 278  WHEN 184  THEN 279
      WHEN 1285 THEN 1286 WHEN 1289 THEN 1287 WHEN 1290 THEN 1288
    END
  WHERE sk."valueInt" IS NOT NULL
    AND st.depth < 10
)
SELECT
  ship_type_id,
  skill_type_id,
  MAX(required_level) AS required_level,
  MIN(depth)          AS depth
FROM skill_tree
GROUP BY ship_type_id, skill_type_id`

	err := global.DB.Raw(sql, shipTypeIDs).Scan(&result).Error
	return result, err
}

// RaceInfo chrRaces 表行
type RaceInfo struct {
	RaceID   int    `json:"race_id"   gorm:"column:raceID"`
	RaceName string `json:"race_name" gorm:"column:raceName"`
}

// GetAllRaces 获取所有种族
func (r *SdeRepository) GetAllRaces() ([]RaceInfo, error) {
	var result []RaceInfo
	err := global.DB.Table(`"chrRaces"`).
		Select(`"raceID", "raceName"`).
		Scan(&result).Error
	return result, err
}

// MarketGroupParent 返回 marketGroupID -> parentGroupID 映射
type MarketGroupNode struct {
	MarketGroupID int `gorm:"column:marketGroupID"`
	ParentGroupID int `gorm:"column:parentGroupID"`
}

// GetMarketGroupTree 获取所有市场分组的父子关系
func (r *SdeRepository) GetMarketGroupTree() ([]MarketGroupNode, error) {
	var result []MarketGroupNode
	err := global.DB.Table(`"invMarketGroups"`).
		Select(`"marketGroupID", COALESCE("parentGroupID", 0) AS "parentGroupID"`).
		Scan(&result).Error
	return result, err
}
