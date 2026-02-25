package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

type MenuService struct {
	repo     *repository.MenuRepository
	roleRepo *repository.RoleRepository
}

func NewMenuService() *MenuService {
	return &MenuService{
		repo:     repository.NewMenuRepository(),
		roleRepo: repository.NewRoleRepository(),
	}
}

// ─── 菜单 CRUD（管理后台用）───

func (s *MenuService) GetMenuTree() ([]*model.Menu, error) {
	menus, err := s.repo.ListAllIncludeDisabled()
	if err != nil {
		return nil, err
	}
	return repository.BuildTree(menus), nil
}

func (s *MenuService) CreateMenu(menu *model.Menu) error {
	if menu.Name == "" {
		return ErrMenuNameEmpty
	}
	if menu.Status == 0 {
		menu.Status = 1
	}
	return s.repo.Create(menu)
}

func (s *MenuService) UpdateMenu(menu *model.Menu) error {
	if _, err := s.repo.GetByID(menu.ID); err != nil {
		return ErrMenuNotFound
	}
	return s.repo.Update(menu)
}

func (s *MenuService) DeleteMenu(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return ErrMenuNotFound
	}
	return s.repo.Delete(id)
}

// ─── 用户菜单（前端路由用）───

// GetUserMenuTree 获取用户可访问的菜单树（前端路由格式）
func (s *MenuService) GetUserMenuTree(userID uint, roleCodes []string) ([]*model.MenuItem, error) {
	// super_admin 返回全部菜单
	if model.IsSuperAdmin(roleCodes) {
		allMenus, err := s.repo.ListAll()
		if err != nil {
			return nil, err
		}
		return repository.BuildMenuTree(allMenus), nil
	}

	// 获取用户角色对应的角色ID
	roleIDs, err := s.roleRepo.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []*model.MenuItem{}, nil
	}

	// 获取所有角色的菜单ID并集
	menuIDs, err := s.roleRepo.GetMenuIDsByRoles(roleIDs)
	if err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return []*model.MenuItem{}, nil
	}

	// 获取菜单详情
	menus, err := s.repo.ListByIDs(menuIDs)
	if err != nil {
		return nil, err
	}

	// 补全父菜单（确保目录菜单被包含）
	menus = s.ensureParentMenus(menus, menuIDs)

	return repository.BuildMenuTree(menus), nil
}

// ensureParentMenus 确保所有菜单的父目录菜单也被包含
func (s *MenuService) ensureParentMenus(menus []model.Menu, existingIDs []uint) []model.Menu {
	idSet := make(map[uint]bool, len(existingIDs))
	for _, id := range existingIDs {
		idSet[id] = true
	}
	for _, m := range menus {
		idSet[m.ID] = true
	}

	// 收集缺失的父菜单ID
	var missingIDs []uint
	for _, m := range menus {
		if m.ParentID != 0 && !idSet[m.ParentID] {
			missingIDs = append(missingIDs, m.ParentID)
			idSet[m.ParentID] = true
		}
	}

	if len(missingIDs) == 0 {
		return menus
	}

	// 查询缺失的父菜单
	parentMenus, err := s.repo.ListByIDs(missingIDs)
	if err != nil {
		return menus
	}

	menus = append(menus, parentMenus...)

	// 递归检查更上层的父菜单
	return s.ensureParentMenus(menus, extractIDs(menus))
}

func extractIDs(menus []model.Menu) []uint {
	ids := make([]uint, len(menus))
	for i, m := range menus {
		ids[i] = m.ID
	}
	return ids
}

// ─── 错误定义 ───

var (
	ErrMenuNameEmpty = errStr("菜单名称不能为空")
	ErrMenuNotFound  = errStr("菜单不存在")
)

type errStr string

func (e errStr) Error() string { return string(e) }
