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
	userSvc              *service.UserService
	roleSvc              *service.RoleService
	charRepo             *repository.EveCharacterRepository
	sysConfigRepo        *repository.SysConfigRepository
	eligibilitySvc       *service.NewbroEligibilityService
	mentorEligibilitySvc *service.MentorEligibilityService
}

func NewMeHandler() *MeHandler {
	return &MeHandler{
		userSvc:              service.NewUserService(),
		roleSvc:              service.NewRoleService(),
		charRepo:             repository.NewEveCharacterRepository(),
		sysConfigRepo:        repository.NewSysConfigRepository(),
		eligibilitySvc:       service.NewNewbroEligibilityService(),
		mentorEligibilitySvc: service.NewMentorEligibilityService(),
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

	characters, err := h.charRepo.ListByUserID(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, "加载人物信息失败")
		return
	}
	if characters == nil {
		characters = []model.EveCharacter{}
	}
	if err := h.userSvc.ValidateCurrentUserBootstrap(user, characters); err != nil {
		response.Fail(c, response.CodeUnauthorized, err.Error())
		return
	}

	roles := middleware.GetUserRoles(c)
	if roles == nil {
		roles = []string{}
	}
	var isCurrentlyNewbro *bool
	if state := h.eligibilitySvc.GetCachedState(userID); state != nil {
		value := state.IsCurrentlyNewbro
		isCurrentlyNewbro = &value
	}
	var isMentorMenteeEligible *bool
	if result, err := h.mentorEligibilitySvc.EvaluateEligibility(userID); err == nil && result != nil {
		value := result.IsEligible
		isMentorMenteeEligible = &value
	}
	enforceCharacterESIRestriction := h.sysConfigRepo.GetBool(
		model.SysConfigEnforceCharacterESIRestriction,
		model.SysConfigDefaultEnforceCharacterESIRestriction,
	)

	response.OK(c, gin.H{
		"user":                              user,
		"characters":                        characters,
		"roles":                             roles,
		"profile_complete":                  user.ProfileComplete(),
		"enforce_character_esi_restriction": enforceCharacterESIRestriction,
		"is_currently_newbro":               isCurrentlyNewbro,
		"is_mentor_mentee_eligible":         isMentorMenteeEligible,
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
