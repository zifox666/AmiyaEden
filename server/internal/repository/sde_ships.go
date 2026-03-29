package repository

import (
	"amiya-eden/global"
	"fmt"
	"strings"
)

// ---- 舰船 & 技能需求查询 ----

const sdeShipSkillRequirementMaxDepth = 10

type sdeSkillRequirementAttributePair struct {
	RequiredSkillAttributeID int
	RequiredLevelAttributeID int
	Name                     string
}

// These dogma attribute IDs are fixed identifiers from the EVE SDE schema.
// dgmTypeAttributes stores only numeric attribute IDs, so we centralize the
// known required-skill -> required-level pairs here instead of scattering
// bare numbers through the recursive SQL.
var sdeShipSkillRequirementAttributePairs = []sdeSkillRequirementAttributePair{
	{RequiredSkillAttributeID: 182, RequiredLevelAttributeID: 277, Name: "requiredSkill1"},
	{RequiredSkillAttributeID: 183, RequiredLevelAttributeID: 278, Name: "requiredSkill2"},
	{RequiredSkillAttributeID: 184, RequiredLevelAttributeID: 279, Name: "requiredSkill3"},
	{RequiredSkillAttributeID: 1285, RequiredLevelAttributeID: 1286, Name: "requiredSkill4"},
	{RequiredSkillAttributeID: 1289, RequiredLevelAttributeID: 1287, Name: "requiredSkill5"},
	{RequiredSkillAttributeID: 1290, RequiredLevelAttributeID: 1288, Name: "requiredSkill6"},
}

// ShipInfo 舰船基础信息（含翻译 + raceID + marketGroupID）
type ShipInfo struct {
	TypeID          int    `json:"type_id"           gorm:"column:type_id"`
	TypeName        string `json:"type_name"         gorm:"column:type_name"`
	GroupID         int    `json:"group_id"          gorm:"column:group_id"`
	GroupName       string `json:"group_name"        gorm:"column:group_name"`
	MarketGroupID   int    `json:"market_group_id"   gorm:"column:market_group_id"`
	MarketGroupName string `json:"market_group_name" gorm:"column:market_group_name"`
	RaceID          int    `json:"race_id"           gorm:"column:race_id"`
}

// ShipSkillReq 舰船技能需求行（含递归层深）
type ShipSkillReq struct {
	ShipTypeID    int `json:"ship_type_id"   gorm:"column:ship_type_id"`
	SkillTypeID   int `json:"skill_type_id"  gorm:"column:skill_type_id"`
	RequiredLevel int `json:"required_level" gorm:"column:required_level"`
	Depth         int `json:"depth"          gorm:"column:depth"`
}

// RaceInfo chrRaces 表行
type RaceInfo struct {
	RaceID   int    `json:"race_id"   gorm:"column:race_id"`
	RaceName string `json:"race_name" gorm:"column:race_name"`
}

// MarketGroupParent 返回 marketGroupID -> parentGroupID 映射
type MarketGroupNode struct {
	MarketGroupID int `gorm:"column:market_group_id"`
	ParentGroupID int `gorm:"column:parent_group_id"`
}

// GetShipsByCategoryID 获取 categoryID=6 的所有已发布舰船（含翻译、raceID、marketGroupID）
func (r *SdeRepository) GetShipsByCategoryID(languageID string) ([]ShipInfo, error) {
	result, err := r.getShipsByCategoryIDWithLayout(languageID, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getShipsByCategoryIDWithLayout(languageID, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getShipsByCategoryIDWithLayout(languageID string, camelCase bool) ([]ShipInfo, error) {
	const shipCategoryID = 6
	var result []ShipInfo
	naming := newSDENaming(camelCase)

	typeIDCol := naming.col("t", "typeID")
	groupIDCol := naming.col("g", "groupID")
	marketGroupIDCol := naming.col("t", "marketGroupID")
	raceIDCol := naming.col("t", "raceID")
	categoryIDCol := naming.col("c", "categoryID")
	publishedCol := naming.col("t", "published")
	typeNameBaseCol := naming.col("t", "typeName")
	groupNameBaseCol := naming.col("g", "groupName")
	marketGroupNameBaseCol := naming.col("mg", "marketGroupName")

	err := global.DB.Table(naming.table("invTypes", "t")).
		Select(fmt.Sprintf(`
			%s AS type_id,
			COALESCE(NULLIF(t_name.text, ''), %s) AS type_name,
			%s AS group_id,
			COALESCE(NULLIF(g_name.text, ''), %s) AS group_name,
			%s AS market_group_id,
			COALESCE(NULLIF(mg_name.text, ''), %s) AS market_group_name,
			%s AS race_id
		`, typeIDCol, typeNameBaseCol, groupIDCol, groupNameBaseCol, marketGroupIDCol, marketGroupNameBaseCol, raceIDCol)).
		Joins(fmt.Sprintf(`JOIN %s ON %s = %s`, naming.table("invGroups", "g"), groupIDCol, naming.col("t", "groupID"))).
		Joins(fmt.Sprintf(`JOIN %s ON %s = %s`, naming.table("invCategories", "c"), categoryIDCol, naming.col("g", "categoryID"))).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invMarketGroups", "mg"), naming.col("mg", "marketGroupID"), marketGroupIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "t_name"), naming.col("t_name", "tcID"), naming.col("t_name", "keyID"), typeIDCol, naming.col("t_name", "languageID")),
			sdeTranslationCategoryIDs["type"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "g_name"), naming.col("g_name", "tcID"), naming.col("g_name", "keyID"), groupIDCol, naming.col("g_name", "languageID")),
			sdeTranslationCategoryIDs["group"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "mg_name"), naming.col("mg_name", "tcID"), naming.col("mg_name", "keyID"), naming.col("mg", "marketGroupID"), naming.col("mg_name", "languageID")),
			sdeTranslationCategoryIDs["market_group"], languageID).
		Where(fmt.Sprintf(`%s = ? AND %s = ?`, categoryIDCol, publishedCol), shipCategoryID, 1).
		Scan(&result).Error
	return result, err
}

// GetShipSkillRequirements 批量获取舰船技能需求（含前置技能递归）
func (r *SdeRepository) GetShipSkillRequirements(shipTypeIDs []int) ([]ShipSkillReq, error) {
	result, err := r.getShipSkillRequirementsWithLayout(shipTypeIDs, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getShipSkillRequirementsWithLayout(shipTypeIDs, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getShipSkillRequirementsWithLayout(shipTypeIDs []int, camelCase bool) ([]ShipSkillReq, error) {
	if len(shipTypeIDs) == 0 {
		return nil, nil
	}

	var result []ShipSkillReq
	naming := newSDENaming(camelCase)
	requiredLevelCase := sdeRequiredSkillLevelCaseSQL(naming.col("sk", "attributeID"))
	requiredSkillFilter := sdeRequiredSkillFilterSQL()

	sql := fmt.Sprintf(`
WITH RECURSIVE skill_tree AS (
  SELECT
    %s AS ship_type_id,
    %s AS skill_type_id,
    %s AS required_level,
    1   AS depth
  FROM %s
  JOIN %s
    ON %s = %s
    AND %s = %s
  WHERE %s IN ?
    AND %s IN (%s)
    AND %s IS NOT NULL

  UNION

  SELECT
    st.ship_type_id,
    %s AS skill_type_id,
    %s AS required_level,
    st.depth + 1
  FROM skill_tree st
  JOIN %s
    ON %s = st.skill_type_id
    AND %s IN (%s)
  JOIN %s
    ON %s = %s
    AND %s = %s
  WHERE %s IS NOT NULL
    AND st.depth < %d
)
SELECT
  ship_type_id,
  skill_type_id,
  MAX(required_level) AS required_level,
  MIN(depth)          AS depth
FROM skill_tree
GROUP BY ship_type_id, skill_type_id`,
		naming.col("sk", "typeID"),
		naming.col("sk", "valueInt"),
		naming.col("lv", "valueInt"),
		naming.table("dgmTypeAttributes", "sk"),
		naming.table("dgmTypeAttributes", "lv"),
		naming.col("sk", "typeID"),
		naming.col("lv", "typeID"),
		naming.col("lv", "attributeID"),
		requiredLevelCase,
		naming.col("sk", "typeID"),
		naming.col("sk", "attributeID"),
		requiredSkillFilter,
		naming.col("sk", "valueInt"),
		naming.col("sk", "valueInt"),
		naming.col("lv", "valueInt"),
		naming.table("dgmTypeAttributes", "sk"),
		naming.col("sk", "typeID"),
		naming.col("sk", "attributeID"),
		requiredSkillFilter,
		naming.table("dgmTypeAttributes", "lv"),
		naming.col("sk", "typeID"),
		naming.col("lv", "typeID"),
		naming.col("lv", "attributeID"),
		requiredLevelCase,
		naming.col("sk", "valueInt"),
		sdeShipSkillRequirementMaxDepth,
	)

	err := global.DB.Raw(sql, shipTypeIDs).Scan(&result).Error
	return result, err
}

func sdeRequiredSkillLevelCaseSQL(attributeIDCol string) string {
	parts := make([]string, 0, len(sdeShipSkillRequirementAttributePairs))
	for _, pair := range sdeShipSkillRequirementAttributePairs {
		parts = append(parts, fmt.Sprintf("WHEN %d THEN %d", pair.RequiredSkillAttributeID, pair.RequiredLevelAttributeID))
	}
	return fmt.Sprintf("CASE %s %s END", attributeIDCol, strings.Join(parts, " "))
}

func sdeRequiredSkillFilterSQL() string {
	ids := make([]string, 0, len(sdeShipSkillRequirementAttributePairs))
	for _, pair := range sdeShipSkillRequirementAttributePairs {
		ids = append(ids, fmt.Sprintf("%d", pair.RequiredSkillAttributeID))
	}
	return strings.Join(ids, ", ")
}

// GetAllRaces 获取所有种族
func (r *SdeRepository) GetAllRaces() ([]RaceInfo, error) {
	result, err := r.getAllRacesWithLayout(true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getAllRacesWithLayout(false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getAllRacesWithLayout(camelCase bool) ([]RaceInfo, error) {
	var result []RaceInfo
	naming := newSDENaming(camelCase)

	err := global.DB.Table(naming.table("chrRaces")).
		Select(fmt.Sprintf(`%s AS race_id, %s AS race_name`, naming.bareCol("raceID"), naming.bareCol("raceName"))).
		Scan(&result).Error
	return result, err
}

// GetMarketGroupTree 获取所有市场分组的父子关系
func (r *SdeRepository) GetMarketGroupTree() ([]MarketGroupNode, error) {
	result, err := r.getMarketGroupTreeWithLayout(true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getMarketGroupTreeWithLayout(false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getMarketGroupTreeWithLayout(camelCase bool) ([]MarketGroupNode, error) {
	var result []MarketGroupNode
	naming := newSDENaming(camelCase)

	err := global.DB.Table(naming.table("invMarketGroups")).
		Select(fmt.Sprintf(`%s AS market_group_id, COALESCE(%s, 0) AS parent_group_id`, naming.bareCol("marketGroupID"), naming.bareCol("parentGroupID"))).
		Scan(&result).Error
	return result, err
}
