package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"
)

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

// GetNames 按 namespace 批量查询 id -> name 映射。
func (r *SdeRepository) GetNames(ids map[string][]int, languageID string) (SdeNameMap, error) {
	result, err := r.getNamesWithLayout(ids, languageID, true)
	if err == nil {
		return result, nil
	}

	fallbackResult, fallbackErr := r.getNamesWithLayout(ids, languageID, false)
	if fallbackErr == nil {
		return fallbackResult, nil
	}

	return nil, fmt.Errorf("%w; fallback query failed: %v", err, fallbackErr)
}

func (r *SdeRepository) getNamesWithLayout(ids map[string][]int, languageID string, camelCase bool) (SdeNameMap, error) {
	result := make(SdeNameMap)
	naming := newSDENaming(camelCase)

	type row struct {
		KeyID int    `gorm:"column:key_id"`
		Text  string `gorm:"column:text"`
	}

	for key, keyIDs := range ids {
		if len(keyIDs) == 0 {
			continue
		}

		tcID, ok := sdeTranslationCategoryIDs[key]
		if !ok {
			continue
		}

		var rows []row
		if err := global.DB.Table(naming.table("trnTranslations")).
			Select(fmt.Sprintf(`%s AS key_id, text`, naming.bareCol("keyID"))).
			Where(
				fmt.Sprintf(`%s = ? AND %s IN ? AND %s = ?`, naming.bareCol("tcID"), naming.bareCol("keyID"), naming.bareCol("languageID")),
				tcID,
				keyIDs,
				languageID,
			).
			Scan(&rows).Error; err != nil {
			return nil, err
		}

		namespaceResult := make(map[int]string, len(rows))
		for _, row := range rows {
			namespaceResult[row.KeyID] = row.Text
		}
		result[key] = namespaceResult
	}

	return result, nil
}
