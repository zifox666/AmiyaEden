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
		"user":             user,
		"characters":       characters,
		"roles":            roles,
		"permissions":      permissions,
		"profile_complete": user.ProfileComplete(),
	})
}

type updateMeRequest struct {
	Nickname  *string `json:"nickname"`
	QQ        *string `json:"qq"`
	DiscordID *string `json:"discord_id"`
}

// UpdateMe 更新当前登录用户的联系资料
func (h *MeHandler) UpdateMe(c *gin.Context) {
	userID := c.GetUint("userID")

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	user, err := h.userSvc.UpdateCurrentProfile(userID, service.UserPatch{
		Nickname:  req.Nickname,
		QQ:        req.QQ,
		DiscordID: req.DiscordID,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, gin.H{
		"user":             user,
		"profile_complete": user.ProfileComplete(),
	})
}
