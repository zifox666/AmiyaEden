package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type MenuRepository struct{}

func NewMenuRepository() *MenuRepository {
	return &MenuRepository{}
}

// ─── Menu CRUD ───

func (r *MenuRepository) Create(menu *model.Menu) error {
	return global.DB.Create(menu).Error
}

func (r *MenuRepository) GetByID(id uint) (*model.Menu, error) {
	var menu model.Menu
	err := global.DB.First(&menu, id).Error
	return &menu, err
}

func (r *MenuRepository) GetByName(name string) (*model.Menu, error) {
	var menu model.Menu
	err := global.DB.Where("name = ?", name).First(&menu).Error
	return &menu, err
}

func (r *MenuRepository) Update(menu *model.Menu) error {
	return global.DB.Save(menu).Error
}

func (r *MenuRepository) Delete(id uint) error {
	tx := global.DB.Begin()
	// 删除角色-菜单关联
	if err := tx.Where("menu_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 递归删除子菜单
	var childIDs []uint
	if err := tx.Model(&model.Menu{}).Where("parent_id = ?", id).Pluck("id", &childIDs).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, childID := range childIDs {
		if err := tx.Where("menu_id = ?", childID).Delete(&model.RoleMenu{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Delete(&model.Menu{}, childID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// 删除本菜单
	if err := tx.Delete(&model.Menu{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// ListAll 获取所有菜单（按 sort 排序）
func (r *MenuRepository) ListAll() ([]model.Menu, error) {
	var menus []model.Menu
	err := global.DB.Where("status = 1").Order("sort DESC, id ASC").Find(&menus).Error
	return menus, err
}

// ListAllIncludeDisabled 获取所有菜单（包含禁用）
func (r *MenuRepository) ListAllIncludeDisabled() ([]model.Menu, error) {
	var menus []model.Menu
	err := global.DB.Order("sort DESC, id ASC").Find(&menus).Error
	return menus, err
}

// ListByIDs 获取指定ID列表的菜单
func (r *MenuRepository) ListByIDs(ids []uint) ([]model.Menu, error) {
	var menus []model.Menu
	err := global.DB.Where("id IN ? AND status = 1", ids).Order("sort DESC, id ASC").Find(&menus).Error
	return menus, err
}

// UpsertByName 按名称查找并创建或更新
func (r *MenuRepository) UpsertByName(menu *model.Menu) error {
	var existing model.Menu
	err := global.DB.Where("name = ?", menu.Name).First(&existing).Error
	if err != nil {
		return global.DB.Create(menu).Error
	}
	menu.ID = existing.ID
	return global.DB.Model(&existing).Updates(map[string]interface{}{
		"parent_id": menu.ParentID, "type": menu.Type,
		"path": menu.Path, "component": menu.Component,
		"permission": menu.Permission, "title": menu.Title,
		"icon": menu.Icon, "sort": menu.Sort,
		"is_hide": menu.IsHide, "keep_alive": menu.KeepAlive,
		"is_hide_tab": menu.IsHideTab, "fixed_tab": menu.FixedTab,
		"status": menu.Status,
	}).Error
}

// ─── 树构建辅助 ───

// BuildTree 将平铺菜单列表构建为树形结构
func BuildTree(menus []model.Menu) []*model.Menu {
	menuMap := make(map[uint]*model.Menu, len(menus))
	roots := make([]*model.Menu, 0)

	// 先创建所有节点的指针
	for i := range menus {
		m := menus[i]
		menuMap[m.ID] = &m
	}

	// 构建父子关系
	for _, m := range menuMap {
		if m.ParentID == 0 {
			roots = append(roots, m)
		} else if parent, ok := menuMap[m.ParentID]; ok {
			parent.Children = append(parent.Children, m)
		}
	}

	// 排序
	sortMenus(roots)
	return roots
}

func sortMenus(menus []*model.Menu) {
	for i := 0; i < len(menus); i++ {
		for j := i + 1; j < len(menus); j++ {
			if menus[i].Sort < menus[j].Sort || (menus[i].Sort == menus[j].Sort && menus[i].ID > menus[j].ID) {
				menus[i], menus[j] = menus[j], menus[i]
			}
		}
		if len(menus[i].Children) > 0 {
			sortMenus(menus[i].Children)
		}
	}
}

// BuildMenuTree 构建前端路由菜单树（目录+页面，按钮转为 authList）
func BuildMenuTree(menus []model.Menu) []*model.MenuItem {
	menuMap := make(map[uint]*model.Menu, len(menus))
	buttonMap := make(map[uint][]*model.Menu) // parentID -> buttons

	// 分类：目录/页面 vs 按钮
	for i := range menus {
		m := menus[i]
		menuMap[m.ID] = &m
		if m.Type == model.MenuTypeButton {
			buttonMap[m.ParentID] = append(buttonMap[m.ParentID], &m)
		}
	}

	// 构建目录和页面的树
	roots := make([]*menuNode, 0)
	nodeMap := make(map[uint]*menuNode)

	for _, m := range menuMap {
		if m.Type == model.MenuTypeButton {
			continue
		}
		node := &menuNode{menu: m, buttons: buttonMap[m.ID]}
		nodeMap[m.ID] = node
	}

	for _, node := range nodeMap {
		if node.menu.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[node.menu.ParentID]; ok {
			parent.children = append(parent.children, node)
		}
	}

	// 排序
	sortNodes(roots)

	// 转换为 MenuItem
	return convertNodes(roots)
}

type menuNode struct {
	menu     *model.Menu
	buttons  []*model.Menu
	children []*menuNode
}

func sortNodes(nodes []*menuNode) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].menu.Sort < nodes[j].menu.Sort || (nodes[i].menu.Sort == nodes[j].menu.Sort && nodes[i].menu.ID > nodes[j].menu.ID) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
		if len(nodes[i].children) > 0 {
			sortNodes(nodes[i].children)
		}
	}
}

func convertNodes(nodes []*menuNode) []*model.MenuItem {
	result := make([]*model.MenuItem, 0, len(nodes))
	for _, node := range nodes {
		item := node.menu.ToMenuItem(node.buttons)
		if len(node.children) > 0 {
			item.Children = convertNodes(node.children)
		}
		result = append(result, item)
	}
	return result
}
