package handler

import (
	"amiya-eden/global"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
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

// getAllowCorpFilter 根据调用者角色返回军团过滤列表
// super_admin 返回 nil（不过滤），admin 返回配置的 allow_corporations
func getAllowCorpFilter(c *gin.Context) []int64 {
	roles := middleware.GetUserRoles(c)
	if model.IsSuperAdmin(roles) {
		return nil
	}
	return global.Config.App.AllowCorporations
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
// 分页查询所有成员某月的联盟 PAP 汇总（管理员）
func (h *AlliancePAPHandler) GetAllAlliancePAP(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

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

	list, total, err := h.svc.GetAllPAPPaged(year, month, page, size, getAllowCorpFilter(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, page, size)
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

// ImportAlliancePAP  POST /system/pap/import
// 导入联盟 PAP 数据（管理员，可指定 year/month）
type importAlliancePAPRequest struct {
	Year          int  `json:"year"  binding:"required"`
	Month         int  `json:"month" binding:"required,min=1,max=12"`
	PAPImportInfo service.PAPImportInfo `json:"data" binding:"required"`
}

func (h *AlliancePAPHandler) ImportAlliancePAP(c *gin.Context) {
	var req importAlliancePAPRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PAPImportInfo.CalculatedAt == "" {
		if err != nil {
			response.Fail(c, response.CodeParamError, "请求参数错误: " + err.Error())
			return
		}
		response.Fail(c, response.CodeParamError, "请求参数错误: 缺少数据时间")
		return
	}

	char, err := h.charRepo.GetByCharacterName(req.PAPImportInfo.PrimaryCharacterName)
	if err != nil {
		response.Fail(c, response.CodeBizError, "主角色不存在")
		return
	}

	user, err := h.userRepo.GetByPrimaryCharacterID(char.CharacterID)
	if err != nil || user.PrimaryCharacterID == 0 {
		response.Fail(c, response.CodeBizError, "未设置主角色")
		return
	}

	err = h.svc.ImportAlliancePAP(req.Year, req.Month, &req.PAPImportInfo, char)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "导入成功"})
}

// GetExchangeConfig  GET /system/pap/config
// 查询 PAP 兑换系统钱包配置
func (h *AlliancePAPHandler) GetExchangeConfig(c *gin.Context) {
	cfg, err := h.svc.GetExchangeConfig()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cfg)
}

// SetExchangeConfig  PUT /system/pap/config
// 更新 PAP 兑换配置
func (h *AlliancePAPHandler) SetExchangeConfig(c *gin.Context) {
	var req service.SetExchangeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	cfg, err := h.svc.SetExchangeConfig(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cfg)
}

// settleMonthRequest 月度结算请求
type settleMonthRequest struct {
	Year          int  `json:"year"  binding:"required"`
	Month         int  `json:"month" binding:"required,min=1,max=12"`
	WalletConvert bool `json:"wallet_convert"` // 是否同时兑换系统钱包
}

// SettleMonth  POST /system/pap/settle
// 管理员触发月度归档（可选同时兑换系统钱包）
func (h *AlliancePAPHandler) SettleMonth(c *gin.Context) {
	var req settleMonthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	operatorID := middleware.GetUserID(c)
	result, err := h.svc.SettleMonth(req.Year, req.Month, req.WalletConvert, operatorID, getAllowCorpFilter(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
