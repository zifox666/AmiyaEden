package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"regexp"
	"strings"
)

type AutoRoleRepository struct{}

func NewAutoRoleRepository() *AutoRoleRepository {
	return &AutoRoleRepository{}
}

// ─── ESI Role Mapping ───

// ListEsiRoleMappings 获取所有 ESI 角色映射
func (r *AutoRoleRepository) ListEsiRoleMappings() ([]model.EsiRoleMapping, error) {
	var mappings []model.EsiRoleMapping
	err := global.DB.Order("esi_role ASC, role_id ASC").Find(&mappings).Error
	return mappings, err
}

// GetEsiRoleMappingsByEsiRole 根据 ESI 角色名获取映射
func (r *AutoRoleRepository) GetEsiRoleMappingsByEsiRole(esiRole string) ([]model.EsiRoleMapping, error) {
	var mappings []model.EsiRoleMapping
	err := global.DB.Where("esi_role = ?", esiRole).Find(&mappings).Error
	return mappings, err
}

// GetEsiRoleMappingsByEsiRoles 根据多个 ESI 角色名获取映射
func (r *AutoRoleRepository) GetEsiRoleMappingsByEsiRoles(esiRoles []string) ([]model.EsiRoleMapping, error) {
	var mappings []model.EsiRoleMapping
	if len(esiRoles) == 0 {
		return mappings, nil
	}
	err := global.DB.Where("esi_role IN ?", esiRoles).Find(&mappings).Error
	return mappings, err
}

// CreateEsiRoleMapping 创建 ESI 角色映射
func (r *AutoRoleRepository) CreateEsiRoleMapping(mapping *model.EsiRoleMapping) error {
	return global.DB.Create(mapping).Error
}

// DeleteEsiRoleMapping 删除 ESI 角色映射
func (r *AutoRoleRepository) DeleteEsiRoleMapping(id uint) error {
	return global.DB.Delete(&model.EsiRoleMapping{}, id).Error
}

// DeleteEsiRoleMappingsByEsiRole 删除指定 ESI 角色的所有映射
func (r *AutoRoleRepository) DeleteEsiRoleMappingsByEsiRole(esiRole string) error {
	return global.DB.Where("esi_role = ?", esiRole).Delete(&model.EsiRoleMapping{}).Error
}

// ─── ESI Title Mapping ───

// ListEsiTitleMappings 获取所有 ESI 头衔映射
func (r *AutoRoleRepository) ListEsiTitleMappings() ([]model.EsiTitleMapping, error) {
	var mappings []model.EsiTitleMapping
	err := global.DB.Order("corporation_id ASC, title_id ASC, role_id ASC").Find(&mappings).Error
	return mappings, err
}

// GetEsiTitleMappingsByCorpAndTitles 根据军团 ID 和头衔 ID 列表获取映射
func (r *AutoRoleRepository) GetEsiTitleMappingsByCorpAndTitles(corpID int64, titleIDs []int) ([]model.EsiTitleMapping, error) {
	var mappings []model.EsiTitleMapping
	if len(titleIDs) == 0 {
		return mappings, nil
	}
	err := global.DB.Where("corporation_id = ? AND title_id IN ?", corpID, titleIDs).Find(&mappings).Error
	return mappings, err
}

// CreateEsiTitleMapping 创建 ESI 头衔映射
func (r *AutoRoleRepository) CreateEsiTitleMapping(mapping *model.EsiTitleMapping) error {
	return global.DB.Create(mapping).Error
}

// DeleteEsiTitleMapping 删除 ESI 头衔映射
func (r *AutoRoleRepository) DeleteEsiTitleMapping(id uint) error {
	return global.DB.Delete(&model.EsiTitleMapping{}, id).Error
}

// ─── Character Corp Roles ───

// ListCharacterCorpRoles 获取角色的所有 ESI 军团角色
func (r *AutoRoleRepository) ListCharacterCorpRoles(characterID int64) ([]string, error) {
	var roles []string
	err := global.DB.Model(&model.EveCharacterCorpRole{}).
		Where("character_id = ?", characterID).
		Pluck("corp_role", &roles).Error
	return roles, err
}

// SyncCharacterCorpRoles 同步角色的 ESI 军团角色（先删后插）
func (r *AutoRoleRepository) SyncCharacterCorpRoles(characterID int64, roles []string) error {
	tx := global.DB.Begin()
	if err := tx.Where("character_id = ?", characterID).Delete(&model.EveCharacterCorpRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(roles) > 0 {
		records := make([]model.EveCharacterCorpRole, 0, len(roles))
		for _, role := range roles {
			records = append(records, model.EveCharacterCorpRole{
				CharacterID: characterID,
				CorpRole:    role,
			})
		}
		if err := tx.Create(&records).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// ListAllCharacterCorpRoles 获取所有角色的 ESI 军团角色（用于批量刷新）
func (r *AutoRoleRepository) ListAllCharacterCorpRoles() ([]model.EveCharacterCorpRole, error) {
	var roles []model.EveCharacterCorpRole
	err := global.DB.Find(&roles).Error
	return roles, err
}

// ─── Auto Role Log ───

// CreateAutoRoleLog 写入一条自动权限操作日志
func (r *AutoRoleRepository) CreateAutoRoleLog(log *model.AutoRoleLog) error {
	return global.DB.Create(log).Error
}

// ListAutoRoleLogs 分页查询自动权限操作日志（按时间倒序）
func (r *AutoRoleRepository) ListAutoRoleLogs(page, pageSize int) ([]model.AutoRoleLog, int64, error) {
	var logs []model.AutoRoleLog
	var total int64
	db := global.DB.Model(&model.AutoRoleLog{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// CorpTitleInfo 军团头衔去重信息（用于前端下拉选择）
type CorpTitleInfo struct {
	CorporationID int64  `json:"corporation_id"`
	TitleID       int    `json:"title_id"`
	TitleName     string `json:"title_name"`
}

var eveTagRe = regexp.MustCompile(`<[^>]+>`)

func stripEveTags(s string) string {
	return strings.TrimSpace(eveTagRe.ReplaceAllString(s, ""))
}

// ListDistinctCorpTitles 获取所有去重的(军团+头衔)组合，来源于 ESI 头衔快照
// allowCorps 非空时只返回指定军团的头衔
func (r *AutoRoleRepository) ListDistinctCorpTitles(allowCorps []int64) ([]CorpTitleInfo, error) {
	var results []CorpTitleInfo
	db := global.DB.
		Table("eve_character_title ect").
		Select("ec.corporation_id, ect.title_id, MIN(ect.name) AS title_name").
		Joins("JOIN eve_character ec ON ec.character_id = ect.character_id").
		Where("ec.corporation_id > 0")
	if len(allowCorps) > 0 {
		db = db.Where("ec.corporation_id IN ?", allowCorps)
	}
	err := db.Group("ec.corporation_id, ect.title_id").
		Order("ec.corporation_id, ect.title_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	for i := range results {
		results[i].TitleName = stripEveTags(results[i].TitleName)
	}
	return results, nil
}
