package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"
)

// KillmailRepository 击杀邮件数据访问层
type KillmailRepository struct{}

func NewKillmailRepository() *KillmailRepository {
	return &KillmailRepository{}
}

// GetCharacterKillmailLink 查询角色-KM 关联记录
func (r *KillmailRepository) GetCharacterKillmailLink(characterID, killmailID int64) (*model.EveCharacterKillmail, error) {
	var ckm model.EveCharacterKillmail
	err := global.DB.Where("character_id = ? AND killmail_id = ?", characterID, killmailID).First(&ckm).Error
	return &ckm, err
}

// GetKillmailByID 按 kill_mail_id 查询 KM 主记录
func (r *KillmailRepository) GetKillmailByID(killmailID int64) (*model.EveKillmailList, error) {
	var km model.EveKillmailList
	err := global.DB.Where("kill_mail_id = ?", killmailID).First(&km).Error
	return &km, err
}

// ListCharacterKillmailsByCharacterIDs 按角色 ID 列表查询关联记录
func (r *KillmailRepository) ListCharacterKillmailsByCharacterIDs(charIDs []int64) ([]model.EveCharacterKillmail, error) {
	var list []model.EveCharacterKillmail
	err := global.DB.Where("character_id IN ?", charIDs).Find(&list).Error
	return list, err
}

// ListVictimKillmailsByCharacterID 查询角色作为受害者的 KM 关联记录
func (r *KillmailRepository) ListVictimKillmailsByCharacterID(characterID int64) ([]model.EveCharacterKillmail, error) {
	var list []model.EveCharacterKillmail
	err := global.DB.Where("character_id = ? AND victim = ?", characterID, true).Find(&list).Error
	return list, err
}

// ListKillmailsByIDsSince 按 ID 列表查询 KM，过滤 since 之后的记录，按时间降序，限制条数
func (r *KillmailRepository) ListKillmailsByIDsSince(kmIDs []int64, since time.Time, limit int) ([]model.EveKillmailList, error) {
	var list []model.EveKillmailList
	err := global.DB.Where("kill_mail_id IN ? AND kill_mail_time >= ?", kmIDs, since).
		Order("kill_mail_time DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// ListKillmailsByIDsInTimeRange 按 ID 列表查询 KM，限定时间范围
func (r *KillmailRepository) ListKillmailsByIDsInTimeRange(kmIDs []int64, startAt, endAt time.Time) ([]model.EveKillmailList, error) {
	var list []model.EveKillmailList
	err := global.DB.Where("kill_mail_id IN ? AND kill_mail_time BETWEEN ? AND ?", kmIDs, startAt, endAt).
		Find(&list).Error
	return list, err
}

// ListKillmailItemsByKillmailID 查询 KM 的所有物品
func (r *KillmailRepository) ListKillmailItemsByKillmailID(killmailID int64) ([]model.EveKillmailItem, error) {
	var list []model.EveKillmailItem
	err := global.DB.Where("kill_mail_id = ?", killmailID).Find(&list).Error
	return list, err
}
