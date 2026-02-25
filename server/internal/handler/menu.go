package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	menuSvc *service.MenuService
	roleSvc *service.RoleService
}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{
		menuSvc: service.NewMenuService(),
		roleSvc: service.NewRoleService(),
	}
}

// GetMenuTree 管理后台获取完整菜单树（含按钮）
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	tree, err := h.menuSvc.GetMenuTree()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, tree)
}

// GetMenuList 当前用户可用菜单（前端路由格式）
func (h *MenuHandler) GetMenuList(c *gin.Context) {
	userID := c.GetUint("userID")
	roleCodes := middleware.GetUserRoles(c)
	tree, err := h.menuSvc.GetUserMenuTree(userID, roleCodes)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, tree)
}

type createMenuReq struct {
	ParentID   uint   `json:"parent_id"`
	Type       string `json:"type" binding:"required,oneof=dir menu button"`
	Name       string `json:"name" binding:"required"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Permission string `json:"permission"`
	Title      string `json:"title" binding:"required"`
	Icon       string `json:"icon"`
	Sort       int    `json:"sort"`
	IsHide     bool   `json:"is_hide"`
	KeepAlive  bool   `json:"keep_alive"`
	IsHideTab  bool   `json:"is_hide_tab"`
	FixedTab   bool   `json:"fixed_tab"`
}

func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req createMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	menu := &model.Menu{
		ParentID:   req.ParentID,
		Type:       req.Type,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Permission: req.Permission,
		Title:      req.Title,
		Icon:       req.Icon,
		Sort:       req.Sort,
		IsHide:     req.IsHide,
		KeepAlive:  req.KeepAlive,
		IsHideTab:  req.IsHideTab,
		FixedTab:   req.FixedTab,
	}
	if err := h.menuSvc.CreateMenu(menu); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, menu)
}

type updateMenuReq struct {
	ParentID   *uint  `json:"parent_id"`
	Type       string `json:"type" binding:"omitempty,oneof=dir menu button"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Permission string `json:"permission"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	Sort       *int   `json:"sort"`
	IsHide     *bool  `json:"is_hide"`
	KeepAlive  *bool  `json:"keep_alive"`
	IsHideTab  *bool  `json:"is_hide_tab"`
	FixedTab   *bool  `json:"fixed_tab"`
}

func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的菜单ID")
		return
	}
	var req updateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	menu := &model.Menu{
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Permission: req.Permission,
		Title:      req.Title,
		Icon:       req.Icon,
	}
	menu.ID = uint(id)
	if req.Type != "" {
		menu.Type = req.Type
	}
	if req.ParentID != nil {
		menu.ParentID = *req.ParentID
	}
	if req.Sort != nil {
		menu.Sort = *req.Sort
	}
	if req.IsHide != nil {
		menu.IsHide = *req.IsHide
	}
	if req.KeepAlive != nil {
		menu.KeepAlive = *req.KeepAlive
	}
	if req.IsHideTab != nil {
		menu.IsHideTab = *req.IsHideTab
	}
	if req.FixedTab != nil {
		menu.FixedTab = *req.FixedTab
	}
	if err := h.menuSvc.UpdateMenu(menu); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的菜单ID")
		return
	}
	if err := h.menuSvc.DeleteMenu(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
