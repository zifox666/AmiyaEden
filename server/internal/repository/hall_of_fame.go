package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HallOfFameRepository 名人堂数据访问层
type HallOfFameRepository struct{}

func NewHallOfFameRepository() *HallOfFameRepository {
	return &HallOfFameRepository{}
}

// ─── Config (singleton) ───

// GetConfig returns the singleton config row, or nil if not yet created.
func (r *HallOfFameRepository) GetConfig() (*model.HallOfFameConfig, error) {
	var cfg model.HallOfFameConfig
	if err := global.DB.First(&cfg, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// UpsertConfig creates or updates the singleton config row.
func (r *HallOfFameRepository) UpsertConfig(cfg *model.HallOfFameConfig) error {
	cfg.ID = 1
	return global.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"background_image", "canvas_width", "canvas_height", "updated_at"}),
	}).Create(cfg).Error
}

// ─── Cards ───

// ListCards returns all cards ordered by z_index ASC, id ASC.
// If visibleOnly is true, only visible cards are returned.
func (r *HallOfFameRepository) ListCards(visibleOnly bool) ([]model.HallOfFameCard, error) {
	var cards []model.HallOfFameCard
	db := global.DB.Model(&model.HallOfFameCard{})
	if visibleOnly {
		db = db.Where("visible = ?", true)
	}
	if err := db.Order("z_index ASC, id ASC").Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// GetCardByID returns a single card by primary key.
func (r *HallOfFameRepository) GetCardByID(id uint) (*model.HallOfFameCard, error) {
	var card model.HallOfFameCard
	if err := global.DB.First(&card, id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

// CreateCard inserts a new card.
func (r *HallOfFameRepository) CreateCard(card *model.HallOfFameCard) error {
	return global.DB.Create(card).Error
}

// UpdateCard saves all fields of an existing card.
func (r *HallOfFameRepository) UpdateCard(card *model.HallOfFameCard) error {
	return global.DB.Save(card).Error
}

// UpdateCardFields updates only the provided fields for a card.
func (r *HallOfFameRepository) UpdateCardFields(id uint, fields map[string]interface{}) error {
	result := global.DB.Model(&model.HallOfFameCard{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("卡片 %d 不存在", id)
	}
	return nil
}

// DeleteCard soft-deletes a card by id.
func (r *HallOfFameRepository) DeleteCard(id uint) error {
	return global.DB.Delete(&model.HallOfFameCard{}, id).Error
}

// BatchUpdateLayout updates pos_x, pos_y, z_index for multiple cards in a transaction.
func (r *HallOfFameRepository) BatchUpdateLayout(updates []model.CardLayoutUpdate) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			result := tx.Model(&model.HallOfFameCard{}).
				Where("id = ?", u.ID).
				Updates(map[string]interface{}{
					"pos_x":   u.PosX,
					"pos_y":   u.PosY,
					"width":   u.Width,
					"height":  u.Height,
					"z_index": u.ZIndex,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("卡片 %d 不存在", u.ID)
			}
		}
		return nil
	})
}
