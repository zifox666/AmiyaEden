package repository

import (
	"amiya-eden/global"
	"fmt"
)

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

// GetTypes 查询物品(组)信息
func (r *SdeRepository) GetTypes(typeIDs []int, published *bool, languageID string) ([]TypeInfo, error) {
	result, err := r.getTypesWithLayout(typeIDs, published, languageID, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getTypesWithLayout(typeIDs, published, languageID, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getTypesWithLayout(typeIDs []int, published *bool, languageID string, camelCase bool) ([]TypeInfo, error) {
	var result []TypeInfo
	naming := newSDENaming(camelCase)

	typeIDCol := naming.col("t", "typeID")
	groupIDCol := naming.col("t", "groupID")
	marketGroupIDCol := naming.col("t", "marketGroupID")
	publishedCol := naming.col("t", "published")
	groupGroupIDCol := naming.col("g", "groupID")
	groupCategoryIDCol := naming.col("g", "categoryID")
	categoryIDCol := naming.col("c", "categoryID")
	marketGroupJoinIDCol := naming.col("mg", "marketGroupID")
	typeNameBaseCol := naming.col("t", "typeName")
	groupNameBaseCol := naming.col("g", "groupName")
	categoryNameBaseCol := naming.col("c", "categoryName")
	marketGroupNameBaseCol := naming.col("mg", "marketGroupName")
	trTcIDCol := naming.col("t_name", "tcID")
	trKeyIDCol := naming.col("t_name", "keyID")
	trLanguageIDCol := naming.col("t_name", "languageID")

	query := global.DB.Table(naming.table("invTypes", "t")).
		Select(fmt.Sprintf(`
			%s AS type_id,
			COALESCE(NULLIF(t_name.text, ''), %s) AS type_name,
			%s AS group_id,
			COALESCE(NULLIF(g_name.text, ''), %s) AS group_name,
			%s AS market_group_id,
			COALESCE(NULLIF(mg_name.text, ''), %s) AS market_group_name,
			%s AS category_id,
			COALESCE(NULLIF(c_name.text, ''), %s) AS category_name
		`, typeIDCol, typeNameBaseCol, groupGroupIDCol, groupNameBaseCol, marketGroupIDCol, marketGroupNameBaseCol, categoryIDCol, categoryNameBaseCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invGroups", "g"), groupGroupIDCol, groupIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invCategories", "c"), categoryIDCol, groupCategoryIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invMarketGroups", "mg"), marketGroupJoinIDCol, marketGroupIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "t_name"), trTcIDCol, trKeyIDCol, typeIDCol, trLanguageIDCol),
			sdeTranslationCategoryIDs["type"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "g_name"), naming.col("g_name", "tcID"), naming.col("g_name", "keyID"), groupGroupIDCol, naming.col("g_name", "languageID")),
			sdeTranslationCategoryIDs["group"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "c_name"), naming.col("c_name", "tcID"), naming.col("c_name", "keyID"), categoryIDCol, naming.col("c_name", "languageID")),
			sdeTranslationCategoryIDs["category"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "mg_name"), naming.col("mg_name", "tcID"), naming.col("mg_name", "keyID"), marketGroupJoinIDCol, naming.col("mg_name", "languageID")),
			sdeTranslationCategoryIDs["market_group"], languageID)

	if len(typeIDs) > 0 {
		query = query.Where(fmt.Sprintf(`%s IN ?`, typeIDCol), typeIDs)
	}
	if published != nil && *published {
		query = query.Where(fmt.Sprintf(`%s = ?`, publishedCol), 1)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// GetTypesByCategoryID 获取指定 categoryID 下所有已发布的物品（含翻译）
// 主要用于技能模块获取 categoryID=16 的全量技能列表
func (r *SdeRepository) GetTypesByCategoryID(categoryID int, languageID string) ([]TypeInfo, error) {
	result, err := r.getTypesByCategoryIDWithLayout(categoryID, languageID, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getTypesByCategoryIDWithLayout(categoryID, languageID, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getTypesByCategoryIDWithLayout(categoryID int, languageID string, camelCase bool) ([]TypeInfo, error) {
	var result []TypeInfo
	naming := newSDENaming(camelCase)

	typeIDCol := naming.col("t", "typeID")
	groupIDCol := naming.col("t", "groupID")
	marketGroupIDCol := naming.col("t", "marketGroupID")
	publishedCol := naming.col("t", "published")
	groupGroupIDCol := naming.col("g", "groupID")
	groupCategoryIDCol := naming.col("g", "categoryID")
	categoryIDCol := naming.col("c", "categoryID")
	marketGroupJoinIDCol := naming.col("mg", "marketGroupID")
	typeNameBaseCol := naming.col("t", "typeName")
	groupNameBaseCol := naming.col("g", "groupName")
	categoryNameBaseCol := naming.col("c", "categoryName")
	marketGroupNameBaseCol := naming.col("mg", "marketGroupName")

	query := global.DB.Table(naming.table("invTypes", "t")).
		Select(fmt.Sprintf(`
			%s AS type_id,
			COALESCE(NULLIF(t_name.text, ''), %s) AS type_name,
			%s AS group_id,
			COALESCE(NULLIF(g_name.text, ''), %s) AS group_name,
			%s AS market_group_id,
			COALESCE(NULLIF(mg_name.text, ''), %s) AS market_group_name,
			%s AS category_id,
			COALESCE(NULLIF(c_name.text, ''), %s) AS category_name
		`, typeIDCol, typeNameBaseCol, groupGroupIDCol, groupNameBaseCol, marketGroupIDCol, marketGroupNameBaseCol, categoryIDCol, categoryNameBaseCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invGroups", "g"), groupGroupIDCol, groupIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invCategories", "c"), categoryIDCol, groupCategoryIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = %s`, naming.table("invMarketGroups", "mg"), marketGroupJoinIDCol, marketGroupIDCol)).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "t_name"), naming.col("t_name", "tcID"), naming.col("t_name", "keyID"), typeIDCol, naming.col("t_name", "languageID")),
			sdeTranslationCategoryIDs["type"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "g_name"), naming.col("g_name", "tcID"), naming.col("g_name", "keyID"), groupGroupIDCol, naming.col("g_name", "languageID")),
			sdeTranslationCategoryIDs["group"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "c_name"), naming.col("c_name", "tcID"), naming.col("c_name", "keyID"), categoryIDCol, naming.col("c_name", "languageID")),
			sdeTranslationCategoryIDs["category"], languageID).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON %s = ? AND %s = %s AND %s = ?`, naming.table("trnTranslations", "mg_name"), naming.col("mg_name", "tcID"), naming.col("mg_name", "keyID"), marketGroupJoinIDCol, naming.col("mg_name", "languageID")),
			sdeTranslationCategoryIDs["market_group"], languageID).
		Where(fmt.Sprintf(`%s = ? AND %s = ?`, categoryIDCol, publishedCol), categoryID, 1)

	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// GetTypeIDsByNames 通过英文名称批量反查 typeID（用于 EFT 解析）
func (r *SdeRepository) GetTypeIDsByNames(names []string) (map[string]int64, error) {
	if len(names) == 0 {
		return map[string]int64{}, nil
	}
	result, err := r.getTypeIDsByNamesWithLayout(names, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getTypeIDsByNamesWithLayout(names, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getTypeIDsByNamesWithLayout(names []string, camelCase bool) (map[string]int64, error) {
	type row struct {
		TypeID   int64  `gorm:"column:type_id"`
		TypeName string `gorm:"column:type_name"`
	}
	var rows []row
	naming := newSDENaming(camelCase)

	err := global.DB.Table(naming.table("invTypes")).
		Select(fmt.Sprintf(`%s AS type_id, %s AS type_name`, naming.bareCol("typeID"), naming.bareCol("typeName"))).
		Where(fmt.Sprintf(`%s IN ?`, naming.bareCol("typeName")), names).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(rows))
	for _, row := range rows {
		result[row.TypeName] = row.TypeID
	}
	return result, nil
}
