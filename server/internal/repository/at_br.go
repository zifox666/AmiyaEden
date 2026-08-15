package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm"
)

type ATBRRepository struct{}

func NewATBRRepository() *ATBRRepository { return &ATBRRepository{} }

type ATBRDetail struct {
	Version      model.ATBRVersion               `json:"version"`
	Teams        []model.ATBRTeamSummary         `json:"teams"`
	Participants []model.ATBRParticipantSnapshot `json:"participants"`
}
type ATBRListItem struct {
	ID        uint      `json:"id"`
	RoundID   uint      `json:"round_id"`
	MatchID   uint      `json:"match_id"`
	MatchUUID string    `json:"match_uuid"`
	RoundNo   int       `json:"round_no"`
	Version   int       `json:"version"`
	Status    string    `json:"status"`
	TeamAName string    `json:"team_a_name"`
	TeamBName string    `json:"team_b_name"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *ATBRRepository) Latest(roundID uint) (*ATBRDetail, error) {
	var version model.ATBRVersion
	if err := global.DB.Where("round_id = ?", roundID).Order("version DESC").First(&version).Error; err != nil {
		return nil, err
	}
	return r.Get(version.ID)
}
func (r *ATBRRepository) Get(id uint) (*ATBRDetail, error) {
	var detail ATBRDetail
	if err := global.DB.First(&detail.Version, id).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("br_version_id = ?", id).Order("side ASC").Find(&detail.Teams).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("br_version_id = ?", id).Order("side ASC, hostile_damage DESC").Find(&detail.Participants).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}
func (r *ATBRRepository) NextVersion(roundID uint) (int, error) {
	var max int
	err := global.DB.Model(&model.ATBRVersion{}).Where("round_id = ?", roundID).Select("COALESCE(MAX(version),0)").Scan(&max).Error
	return max + 1, err
}
func (r *ATBRRepository) Create(version *model.ATBRVersion, teams []model.ATBRTeamSummary, participants []model.ATBRParticipantSnapshot) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		for i := range teams {
			teams[i].BRVersionID = version.ID
		}
		for i := range participants {
			participants[i].BRVersionID = version.ID
		}
		if err := tx.Create(&teams).Error; err != nil {
			return err
		}
		return tx.Create(&participants).Error
	})
}
func (r *ATBRRepository) Exists(roundID uint) (bool, error) {
	var count int64
	err := global.DB.Model(&model.ATBRVersion{}).Where("round_id = ? AND status = ?", roundID, "confirmed").Count(&count).Error
	return count > 0, err
}
func (r *ATBRRepository) ListLatest() ([]ATBRListItem, error) {
	var result []ATBRListItem
	err := global.DB.Table("at_br_version br").
		Select("br.id,br.round_id,r.match_id,m.uuid AS match_uuid,r.round_no,br.version,br.status,m.team_a_name,m.team_b_name,br.created_at").
		Joins("JOIN at_match_round r ON r.id = br.round_id").
		Joins("JOIN at_match m ON m.id = r.match_id").
		Where("br.version = (SELECT MAX(v.version) FROM at_br_version v WHERE v.round_id = br.round_id AND v.deleted_at IS NULL)").
		Where("br.deleted_at IS NULL AND r.deleted_at IS NULL AND m.deleted_at IS NULL").
		Order("br.created_at DESC").Scan(&result).Error
	return result, err
}
