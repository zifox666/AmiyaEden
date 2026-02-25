package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

// ─── Role CRUD ───

func (r *RoleRepository) Create(role *model.Role) error {
	return global.DB.Create(role).Error
}

func (r *RoleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := global.DB.First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := global.DB.Where("code = ?", code).First(&role).Error
	return &role, err
}

func (r *RoleRepository) Update(role *model.Role) error {
	return global.DB.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	tx := global.DB.Begin()
	if err := tx.Where("role_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.Role{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *RoleRepository) List(page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	db := global.DB.Model(&model.Role{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("sort DESC, id ASC").Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *RoleRepository) ListAll() ([]model.Role, error) {
	var roles []model.Role
	err := global.DB.Order("sort DESC, id ASC").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) UpsertSystemRole(role *model.Role) error {
	var existing model.Role
	err := global.DB.Where("code = ?", role.Code).First(&existing).Error
	if err != nil {
		return global.DB.Create(role).Error
	}
	return global.DB.Model(&existing).Updates(map[string]interface{}{
		"name": role.Name, "description": role.Description,
		"is_system": role.IsSystem, "sort": role.Sort, "status": role.Status,
	}).Error
}

// ─── RoleMenu ───

func (r *RoleRepository) GetRoleMenuIDs(roleID uint) ([]uint, error) {
	var ids []uint
	err := global.DB.Model(&model.RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}

func (r *RoleRepository) SetRoleMenus(roleID uint, menuIDs []uint) error {
	tx := global.DB.Begin()
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, mid := range menuIDs {
		if err := tx.Create(&model.RoleMenu{RoleID: roleID, MenuID: mid}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetMenuIDsByRoles 获取多个角色的所有菜单ID并集
func (r *RoleRepository) GetMenuIDsByRoles(roleIDs []uint) ([]uint, error) {
	var ids []uint
	err := global.DB.Model(&model.RoleMenu{}).Where("role_id IN ?", roleIDs).Distinct("menu_id").Pluck("menu_id", &ids).Error
	return ids, err
}

// ─── UserRole ───

func (r *RoleRepository) GetUserRoleIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := global.DB.Model(&model.UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error
	return ids, err
}

func (r *RoleRepository) GetUserRoleCodes(userID uint) ([]string, error) {
	var codes []string
	err := global.DB.Model(&model.UserRole{}).
		Joins("JOIN role ON role.id = user_role.role_id").
		Where("user_role.user_id = ? AND role.status = 1", userID).
		Pluck("role.code", &codes).Error
	return codes, err
}

func (r *RoleRepository) SetUserRoles(userID uint, roleIDs []uint) error {
	tx := global.DB.Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, rid := range roleIDs {
		if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *RoleRepository) AddUserRole(userID, roleID uint) error {
	return global.DB.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *RoleRepository) RemoveUserRole(userID, roleID uint) error {
	return global.DB.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error
}

// GetRoleUsers 获取拥有某角色的所有用户ID
func (r *RoleRepository) GetRoleUserIDs(roleID uint) ([]uint, error) {
	var ids []uint
	err := global.DB.Model(&model.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &ids).Error
	return ids, err
}
