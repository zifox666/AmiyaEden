package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// SeatUserRepository SeAT 用户绑定数据访问层
type SeatUserRepository struct{}

func NewSeatUserRepository() *SeatUserRepository {
	return &SeatUserRepository{}
}

// Create 创建 SeAT 用户绑定记录
func (r *SeatUserRepository) Create(su *model.SeatUser) error {
	return global.DB.Create(su).Error
}

// Update 更新 SeAT 用户绑定记录
func (r *SeatUserRepository) Update(su *model.SeatUser) error {
	return global.DB.Save(su).Error
}

// GetBySeatUserID 根据 SeAT 用户 ID 查询
func (r *SeatUserRepository) GetBySeatUserID(seatUserID string) (*model.SeatUser, error) {
	var su model.SeatUser
	err := global.DB.Where("seat_user_id = ?", seatUserID).First(&su).Error
	return &su, err
}

// GetByUserID 根据本系统用户 ID 查询
func (r *SeatUserRepository) GetByUserID(userID uint) (*model.SeatUser, error) {
	var su model.SeatUser
	err := global.DB.Where("user_id = ?", userID).First(&su).Error
	return &su, err
}

// Delete 删除 SeAT 用户绑定记录
func (r *SeatUserRepository) Delete(id uint) error {
	return global.DB.Unscoped().Delete(&model.SeatUser{}, id).Error
}

// ListAll 获取所有 SeAT 用户绑定记录
func (r *SeatUserRepository) ListAll() ([]model.SeatUser, error) {
	var list []model.SeatUser
	err := global.DB.Find(&list).Error
	return list, err
}
