package repository

import (
	"amiya-eden/global"
	"fmt"
)

// FlagInfo invFlags 表行
type FlagInfo struct {
	FlagID   int    `json:"flag_id"   gorm:"column:flag_id"`
	FlagName string `json:"flag_name" gorm:"column:flag_name"`
	FlagText string `json:"flag_text" gorm:"column:flag_text"`
	OrderID  int    `json:"order_id"  gorm:"column:order_id"`
}

// FuzzySearchItem 模糊搜索结果条目
type FuzzySearchItem struct {
	ID        int    `json:"id"         gorm:"column:id"`
	Name      string `json:"name"       gorm:"column:name"`
	GroupID   int    `json:"group_id"   gorm:"column:group_id"`
	GroupName string `json:"group_name" gorm:"column:group_name"`
	Category  string `json:"category"   gorm:"column:category"` // "type" | "character"
}

// FuzzySearch 模糊搜索 trnTranslations（物品名称）及成员名称
func (r *SdeRepository) FuzzySearch(keyword string, languageID string, categoryIDs []int, excludeCategoryIDs []int, limit int, searchMember bool) ([]FuzzySearchItem, error) {
	result, err := r.fuzzySearchWithLayout(keyword, languageID, categoryIDs, excludeCategoryIDs, limit, searchMember, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.fuzzySearchWithLayout(keyword, languageID, categoryIDs, excludeCategoryIDs, limit, searchMember, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) fuzzySearchWithLayout(keyword string, languageID string, categoryIDs []int, excludeCategoryIDs []int, limit int, searchMember bool, camelCase bool) ([]FuzzySearchItem, error) {
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
	naming := newSDENaming(camelCase)

	typeIDCol := naming.col("t", "typeID")
	groupIDCol := naming.col("t", "groupID")
	publishedCol := naming.col("t", "published")
	groupGroupIDCol := naming.col("g", "groupID")
	groupCategoryIDCol := naming.col("g", "categoryID")
	categoryIDCol := naming.col("c", "categoryID")
	typeNameBaseCol := naming.col("t", "typeName")
	groupNameBaseCol := naming.col("g", "groupName")

	query := global.DB.Table(naming.table("invTypes", "t")).
		Select(fmt.Sprintf(`
			%s AS id,
			COALESCE(NULLIF(tr.text, ''), %s) AS name,
			%s AS group_id,
			COALESCE(NULLIF(g_name.text, ''), %s) AS group_name,
			'type' AS category
		`, typeIDCol, typeNameBaseCol, groupGroupIDCol, groupNameBaseCol)).
		Joins(fmt.Sprintf(`JOIN %s ON %s = %s`, naming.table("invGroups", "g"), groupGroupIDCol, groupIDCol)).
		Joins(fmt.Sprintf(`JOIN %s ON %s = %s`, naming.table("invCategories", "c"), categoryIDCol, groupCategoryIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "tr"), naming.col("tr", "tcID"), naming.col("tr", "keyID"), typeIDCol, naming.col("tr", "languageID")),
			sdeTranslationCategoryIDs["type"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "g_name"), naming.col("g_name", "tcID"), naming.col("g_name", "keyID"), groupGroupIDCol, naming.col("g_name", "languageID")),
			sdeTranslationCategoryIDs["group"], languageID).
		Where(fmt.Sprintf(`%s = 1`, publishedCol)).
		Where(fmt.Sprintf(`(tr.text ILIKE ? OR %s ILIKE ?)`, typeNameBaseCol), pattern, pattern)

	if len(categoryIDs) > 0 {
		query = query.Where(fmt.Sprintf(`%s IN ?`, categoryIDCol), categoryIDs)
	}
	if len(excludeCategoryIDs) > 0 {
		query = query.Where(fmt.Sprintf(`%s NOT IN ?`, categoryIDCol), excludeCategoryIDs)
	}

	query = query.Limit(limit)
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

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
	result, err := r.getFlagsWithLayout(flagIDs, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getFlagsWithLayout(flagIDs, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getFlagsWithLayout(flagIDs []int, camelCase bool) ([]FlagInfo, error) {
	var result []FlagInfo
	if len(flagIDs) == 0 {
		return result, nil
	}

	naming := newSDENaming(camelCase)
	err := global.DB.Table(naming.table("invFlags")).
		Select(fmt.Sprintf(`%s AS flag_id, %s AS flag_name, %s AS flag_text, %s AS order_id`, naming.bareCol("flagID"), naming.bareCol("flagName"), naming.bareCol("flagText"), naming.bareCol("orderID"))).
		Where(fmt.Sprintf(`%s IN ?`, naming.bareCol("flagID")), flagIDs).
		Order(fmt.Sprintf(`%s ASC`, naming.bareCol("orderID"))).
		Scan(&result).Error
	return result, err
}
