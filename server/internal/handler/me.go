package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type MeHandler struct {
	userSvc  *service.UserService
	roleSvc  *service.RoleService
	charRepo *repository.EveCharacterRepository
}

func NewMeHandler() *MeHandler {
	return &MeHandler{
		userSvc:  service.NewUserService(),
		roleSvc:  service.NewRoleService(),
		charRepo: repository.NewEveCharacterRepository(),
	}
}

// GetMe 获取当前登录用户信息
func (h *MeHandler) GetMe(c *gin.Context) {
	userID := c.GetUint("userID")

	user, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		response.Fail(c, response.CodeUnauthorized, "用户不存在")
		return
	}

	characters, _ := h.charRepo.ListByUserID(userID)
	if characters == nil {
		characters = []model.EveCharacter{}
	}

	roles := middleware.GetUserRoles(c)
	if roles == nil {
		roles = []string{}
	}
	permissions := middleware.GetUserPermissions(c)
	if permissions == nil {
		permissions = []string{}
	}

	response.OK(c, gin.H{
		"user":        user,
		"characters":  characters,
		"roles":       roles,
		"permissions": permissions,
	})
}
