package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单 HTTP 处理器
type MenuHandler struct{}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{}
}

// GetMenuList 获取当前登录用户的菜单路由列表
//
//	GET /api/v1/menu
//
// 根据用户角色在服务端过滤可访问的路由，返回给前端进行动态路由注册。
func (h *MenuHandler) GetMenuList(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role == "" {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	menus := model.GetMenuByRole(role)
	response.OK(c, menus)
}
