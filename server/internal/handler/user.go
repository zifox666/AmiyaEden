package handler

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
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

	filter := repository.UserFilter{
		Nickname: c.Query("nickname"),
	}
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Status = &v
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
	if err != nil {
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
	Nickname string `json:"nickname"`
	Status   *int8  `json:"status"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	user := &model.User{}
	user.ID = uint(id)
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if err := h.svc.UpdateUser(user); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	if err := h.svc.DeleteUser(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
