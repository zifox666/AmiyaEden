package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// EveInfoHandler EVE 角色信息处理器
type EveInfoHandler struct {
	svc         *service.EveInfoService
	cloneSvc    *service.CloneService
	assetSvc    *service.AssetService
	contractSvc *service.ContractService
}

func NewEveInfoHandler() *EveInfoHandler {
	return &EveInfoHandler{
		svc:         service.NewEveInfoService(),
		cloneSvc:    service.NewCloneService(),
		assetSvc:    service.NewAssetService(),
		contractSvc: service.NewContractService(),
	}
}

// GetWalletJournal POST /info/wallet
// 获取指定角色的钱包余额和流水记录
func (h *EveInfoHandler) GetWalletJournal(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetWalletJournal(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCharacterSkills POST /info/skills
// 获取指定角色的技能列表和学习队列
func (h *EveInfoHandler) GetCharacterSkills(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetCharacterSkills(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCharacterShips POST /info/ships
// 获取指定角色的可用舰船列表
func (h *EveInfoHandler) GetCharacterShips(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoShipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetCharacterShips(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCharacterImplants POST /info/implants
// 获取指定角色的克隆体/植入体/跳跃疲劳信息
func (h *EveInfoHandler) GetCharacterImplants(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoImplantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.cloneSvc.GetCharacterImplants(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetAssets POST /info/assets
// 获取当前用户所有角色的资产汇总
func (h *EveInfoHandler) GetAssets(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoAssetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.assetSvc.GetUserAssets(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetContracts POST /info/contracts
// 分页获取当前用户所有角色的合同
func (h *EveInfoHandler) GetContracts(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoContractsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	list, total, err := h.contractSvc.GetUserContracts(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// GetContractDetail POST /info/contracts/detail
// 获取指定合同的物品与竞标详情
func (h *EveInfoHandler) GetContractDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoContractDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.contractSvc.GetContractDetail(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
