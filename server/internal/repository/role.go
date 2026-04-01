package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

// ─── UserRole (code-based) ───

// GetUserRoleCodes 获取用户的职权编码列表（按优先级降序）
func (r *RoleRepository) GetUserRoleCodes(userID uint) ([]string, error) {
	var codes []string
	err := global.DB.Model(&model.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_code", &codes).Error
	if err != nil {
		return nil, err
	}
	// Sort by role definition sort order (descending)
	return sortRoleCodesByPriority(codes), nil
}

// GetUserRoleCodesByUserIDs 批量获取多个用户的职权编码
func (r *RoleRepository) GetUserRoleCodesByUserIDs(userIDs []uint) (map[uint][]string, error) {
	roleCodesByUserID := make(map[uint][]string, len(userIDs))
	if len(userIDs) == 0 {
		return roleCodesByUserID, nil
	}

	type userRoleCodeRow struct {
		UserID   uint
		RoleCode string
	}

	var rows []userRoleCodeRow
	err := global.DB.Table("user_role").
		Select("user_id, role_code").
		Where("user_id IN ?", userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		roleCodesByUserID[row.UserID] = append(roleCodesByUserID[row.UserID], row.RoleCode)
	}
	// Sort each user's roles by priority
	for uid, codes := range roleCodesByUserID {
		roleCodesByUserID[uid] = sortRoleCodesByPriority(codes)
	}
	return roleCodesByUserID, nil
}

// SetUserRoles 设置用户的职权（替换所有）
func (r *RoleRepository) SetUserRoles(userID uint, roleCodes []string) error {
	tx := global.DB.Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, code := range roleCodes {
		if err := tx.Create(&model.UserRole{UserID: userID, RoleCode: code}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// AddUserRole 为用户添加一个职权
func (r *RoleRepository) AddUserRole(userID uint, roleCode string) error {
	return global.DB.Create(&model.UserRole{UserID: userID, RoleCode: roleCode}).Error
}

// RemoveUserRole 移除用户的一个职权
func (r *RoleRepository) RemoveUserRole(userID uint, roleCode string) error {
	return global.DB.Where("user_id = ? AND role_code = ?", userID, roleCode).Delete(&model.UserRole{}).Error
}

// GetRoleUserIDs 获取拥有某职权的所有用户ID
func (r *RoleRepository) GetRoleUserIDs(roleCode string) ([]uint, error) {
	var ids []uint
	err := global.DB.Table(`user_role AS ur`).
		Joins(`JOIN "user" AS u ON u.id = ur.user_id`).
		Where("ur.role_code = ? AND u.deleted_at IS NULL", roleCode).
		Pluck("ur.user_id", &ids).Error
	return ids, err
}

// ─── 内部辅助 ───

func sortRoleCodesByPriority(codes []string) []string {
	if len(codes) <= 1 {
		return codes
	}
	priorityOf := func(code string) int {
		if def, ok := model.GetRoleDefinition(code); ok {
			return def.Sort
		}
		return -1
	}
	// Simple insertion sort (small N)
	for i := 1; i < len(codes); i++ {
		for j := i; j > 0 && priorityOf(codes[j]) > priorityOf(codes[j-1]); j-- {
			codes[j], codes[j-1] = codes[j-1], codes[j]
		}
	}
	return codes
}
