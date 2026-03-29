package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/internal/utils"
	"amiya-eden/pkg/response"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{svc: service.NewUserService()}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	page, size = normalizeLedgerPagination(page, size)

	filter := repository.UserFilter{
		Keyword: c.Query("keyword"),
	}
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Status = &v
	}

	// admin 只能看到 allow_corporations 下有角色的用户，super_admin 看全部
	roles := middleware.GetUserRoles(c)
	if !model.IsSuperAdmin(roles) {
		filter.AllowCorporations = utils.GetAllowCorporations()
	}

	users, total, err := h.svc.ListUsers(page, size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, users, total, page, size)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	user, err := h.svc.GetUserByID(uint(id))
	if err != nil {
		response.Fail(c, response.CodeNotFound, "用户不存在")
		return
	}
	response.OK(c, user)
}

type updateUserRequest struct {
	Nickname  *string `json:"nickname"`
	QQ        *string `json:"qq"`
	DiscordID *string `json:"discord_id"`
	Status    *int8   `json:"status"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	operatorRoles := middleware.GetUserRoles(c)
	if err := h.svc.UpdateUserByAdmin(uint(id), operatorRoles, service.UserPatch{
		Nickname:  req.Nickname,
		QQ:        req.QQ,
		DiscordID: req.DiscordID,
		Status:    req.Status,
	}); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	operatorRoles := middleware.GetUserRoles(c)
	if err := h.svc.DeleteUser(uint(id), operatorRoles); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ImpersonateUser 以指定用户身份签发 JWT（仅超级管理员可用）
func (h *UserHandler) ImpersonateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	token, user, err := h.svc.ImpersonateUser(uint(id))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"token": token,
		"user":  user,
	})
}
