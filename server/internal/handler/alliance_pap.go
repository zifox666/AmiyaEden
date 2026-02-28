package handler

import (
	"amiya-eden/global"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AlliancePAPHandler 联盟 PAP HTTP 处理器
type AlliancePAPHandler struct {
	svc      *service.AlliancePAPService
	charRepo *repository.EveCharacterRepository
	userRepo *repository.UserRepository
}

func NewAlliancePAPHandler() *AlliancePAPHandler {
	return &AlliancePAPHandler{
		svc:      service.NewAlliancePAPService(),
		charRepo: repository.NewEveCharacterRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

// GetMyAlliancePAP  GET /operation/pap/alliance
// 查询当前登录用户主角色的联盟 PAP 数据（默认当月）
func (h *AlliancePAPHandler) GetMyAlliancePAP(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			month = v
		}
	}

	userID := middleware.GetUserID(c)
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user.PrimaryCharacterID == 0 {
		response.Fail(c, response.CodeBizError, "未设置主角色")
		return
	}

	char, err := h.charRepo.GetByCharacterID(user.PrimaryCharacterID)
	if err != nil {
		response.Fail(c, response.CodeBizError, "主角色不存在")
		return
	}

	result, err := h.svc.GetMyPAP(char.CharacterName, year, month)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetAllAlliancePAP  GET /system/pap
// 查询所有成员某月的联盟 PAP 汇总（管理员）
func (h *AlliancePAPHandler) GetAllAlliancePAP(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			month = v
		}
	}

	list, err := h.svc.GetAllPAP(year, month)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"year":  year,
		"month": month,
		"list":  list,
	})
}

// TriggerFetch  POST /system/pap/fetch
// 手动触发拉取（管理员，可指定 year/month）
func (h *AlliancePAPHandler) TriggerFetch(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			month = v
		}
	}

	go func() {
		global.Logger.Info("手动触发联盟 PAP 拉取")
		h.svc.FetchAllUsers(year, month)
	}()

	response.OK(c, gin.H{"message": "已触发后台拉取任务"})
}
