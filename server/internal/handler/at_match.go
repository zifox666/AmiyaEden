package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ATMatchHandler struct{ svc *service.ATMatchService }

func NewATMatchHandler() *ATMatchHandler { return &ATMatchHandler{svc: service.NewATMatchService()} }
func atAdmin(c *gin.Context) bool {
	return model.ContainsAnyRole(middleware.GetUserRoles(c), model.RoleAdmin, model.RoleSuperAdmin)
}
func queryUint(c *gin.Context, name string) (uint, error) {
	value, err := strconv.ParseUint(c.Query(name), 10, 64)
	return uint(value), err
}
func (h *ATMatchHandler) List(c *gin.Context) {
	result, err := h.svc.List()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) Get(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	match, rounds, registrations, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, gin.H{"match": match, "rounds": rounds, "registrations": registrations})
}
func (h *ATMatchHandler) GetByUUID(c *gin.Context) {
	match, rounds, registrations, err := h.svc.GetByUUID(c.Param("uuid"))
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, gin.H{"match": match, "rounds": rounds, "registrations": registrations})
}
func (h *ATMatchHandler) Create(c *gin.Context) {
	var input service.ATMatchCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.Create(middleware.GetUserID(c), input)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) Transition(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.Transition(middleware.GetUserID(c), id, input.Status, input.Reason, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) Register(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Register(middleware.GetUserID(c), id); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) SavePosition(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input service.ATPositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SavePosition(middleware.GetUserID(c), id, input, atAdmin(c), true); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) AssignPosition(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input service.ATPositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SavePosition(middleware.GetUserID(c), id, input, atAdmin(c), false); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) PositionEFT(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	slot, err := strconv.Atoi(c.Param("slot"))
	if err != nil || slot <= 0 {
		response.Fail(c, response.CodeParamError, "阵容位置无效")
		return
	}
	result, err := h.svc.PositionEFT(middleware.GetUserID(c), id, slot, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) DecideRegistration(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		UserID uint   `json:"user_id"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.DecideRegistration(middleware.GetUserID(c), id, input.UserID, input.Status, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) CreateRound(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input service.ATCreateRoundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.CreateRound(middleware.GetUserID(c), id, input, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) SetRoster(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Roster []model.ATMatchRoundRoster `json:"roster"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SetRoundRoster(middleware.GetUserID(c), id, input.Roster, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) GetRoster(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.GetRoundRoster(middleware.GetUserID(c), id, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) FinishRound(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.FinishRound(middleware.GetUserID(c), id, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) ReopenRound(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.VoidAndReopen(middleware.GetUserID(c), id, input.Reason, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) ConfirmRound(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.ConfirmRound(middleware.GetUserID(c), id, input.Result, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) Reset(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.Reset(middleware.GetUserID(c), id, input.Reason); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) SaveSeason(c *gin.Context) {
	var input model.ATSeason
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if id := c.Param("id"); id != "" {
		value, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			response.Fail(c, response.CodeParamError, "无效 ID")
			return
		}
		input.ID = uint(value)
	}
	if err := h.svc.SaveSeason(&input); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, input)
}
func (h *ATMatchHandler) ListSeasons(c *gin.Context) {
	result, err := h.svc.ListSeasons()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) SaveMatchType(c *gin.Context) {
	var input model.ATMatchType
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if id := c.Param("id"); id != "" {
		value, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			response.Fail(c, response.CodeParamError, "无效 ID")
			return
		}
		input.ID = uint(value)
	}
	if err := h.svc.SaveMatchType(&input); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, input)
}
func (h *ATMatchHandler) ListMatchTypes(c *gin.Context) {
	result, err := h.svc.ListMatchTypes()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) SetCleanup(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		TypeIDs []int64 `json:"type_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SetCleanupShips(id, input.TypeIDs); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) ListCleanup(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.ListCleanupShips(id, c.DefaultQuery("language", "en"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) SearchContributionUsers(c *gin.Context) {
	result, err := h.svc.SearchContributionUsers(c.Query("keyword"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATMatchHandler) ManualContribution(c *gin.Context) {
	var input struct {
		SeasonID uint    `json:"season_id"`
		UserID   uint    `json:"user_id"`
		Amount   float64 `json:"amount"`
		Reason   string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.ManualContribution(middleware.GetUserID(c), input.SeasonID, input.UserID, input.Amount, input.Reason); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATMatchHandler) Leaderboard(c *gin.Context) {
	var seasonID *uint
	if raw := c.Query("season_id"); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, response.CodeParamError, "无效赛季 ID")
			return
		}
		converted := uint(value)
		seasonID = &converted
	}
	result, err := h.svc.Leaderboard(seasonID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATMatchHandler) MyContributionLedger(c *gin.Context) {
	var seasonID *uint
	if value := c.Query("season_id"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			response.Fail(c, response.CodeParamError, "赛季 ID 无效")
			return
		}
		id := uint(parsed)
		seasonID = &id
	}
	result, err := h.svc.UserLedger(middleware.GetUserID(c), seasonID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
