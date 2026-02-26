package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Contracts 角色合同
//  GET /characters/{character_id}/contracts
//  GET /characters/{character_id}/contracts/{contract_id}/bids
//  GET /characters/{character_id}/contracts/{contract_id}/items
//  默认刷新间隔: 1 Day / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&ContractsTask{})
}

// ContractsTask 角色合同刷新任务
type ContractsTask struct{}

func (t *ContractsTask) Name() string        { return "character_contracts" }
func (t *ContractsTask) Description() string { return "角色合同（含竞标/物品）" }
func (t *ContractsTask) Priority() Priority  { return PriorityNormal }

func (t *ContractsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *ContractsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-contracts.read_character_contracts.v1", Description: "读取角色合同"},
	}
}

// Contract 合同
type Contract struct {
	AcceptorID          int64      `json:"acceptor_id"`
	AssigneeID          int64      `json:"assignee_id"`
	Availability        string     `json:"availability"`
	Buyout              *float64   `json:"buyout,omitempty"`
	Collateral          *float64   `json:"collateral,omitempty"`
	ContractID          int64      `json:"contract_id"`
	DateAccepted        *time.Time `json:"date_accepted,omitempty"`
	DateCompleted       *time.Time `json:"date_completed,omitempty"`
	DateExpired         time.Time  `json:"date_expired"`
	DateIssued          time.Time  `json:"date_issued"`
	DaysToComplete      *int       `json:"days_to_complete,omitempty"`
	EndLocationID       *int64     `json:"end_location_id,omitempty"`
	ForCorporation      bool       `json:"for_corporation"`
	IssuerCorporationID int64      `json:"issuer_corporation_id"`
	IssuerID            int64      `json:"issuer_id"`
	Price               *float64   `json:"price,omitempty"`
	Reward              *float64   `json:"reward,omitempty"`
	StartLocationID     *int64     `json:"start_location_id,omitempty"`
	Status              string     `json:"status"`
	Title               *string    `json:"title,omitempty"`
	Type                string     `json:"type"`
	Volume              *float64   `json:"volume,omitempty"`
}

// ContractBid 合同竞标
type ContractBid struct {
	Amount   float64   `json:"amount"`
	BidID    int64     `json:"bid_id"`
	BidderID int64     `json:"bidder_id"`
	DateBid  time.Time `json:"date_bid"`
}

// ContractItem 合同物品
type ContractItem struct {
	IsIncluded  bool  `json:"is_included"`
	IsSingleton bool  `json:"is_singleton"`
	Quantity    int   `json:"quantity"`
	RawQuantity *int  `json:"raw_quantity,omitempty"`
	RecordID    int64 `json:"record_id"`
	TypeID      int   `json:"type_id"`
}

func (t *ContractsTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	// 1. 获取合同列表（自动分页）
	contractPath := fmt.Sprintf("/characters/%d/contracts/", ctx.CharacterID)
	var contracts []Contract
	if _, err := ctx.Client.GetPaginated(bgCtx, contractPath, ctx.AccessToken, &contracts); err != nil {
		return fmt.Errorf("fetch contracts: %w", err)
	}

	global.Logger.Debug("[ESI] 角色合同刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(contracts)),
	)

	// 2. 获取未完成合同的竞标和物品详情，并入库
	for _, c := range contracts {
		record := model.EveCharacterContract{
			CharacterID:         ctx.CharacterID,
			ContractID:          c.ContractID,
			AcceptorID:          c.AcceptorID,
			AssigneeID:          c.AssigneeID,
			Availability:        c.Availability,
			Buyout:              c.Buyout,
			Collateral:          c.Collateral,
			DateAccepted:        c.DateAccepted,
			DateCompleted:       c.DateCompleted,
			DateExpired:         c.DateExpired,
			DateIssued:          c.DateIssued,
			DaysToComplete:      c.DaysToComplete,
			EndLocationID:       c.EndLocationID,
			ForCorporation:      c.ForCorporation,
			IssuerCorporationID: c.IssuerCorporationID,
			IssuerID:            c.IssuerID,
			Price:               c.Price,
			Reward:              c.Reward,
			StartLocationID:     c.StartLocationID,
			Status:              c.Status,
			Title:               c.Title,
			Type:                c.Type,
			Volume:              c.Volume,
		}

		// 拍卖类活跃合同获取竞标
		if (c.Status == "outstanding" || c.Status == "in_progress") && c.Type == "auction" {
			bidPath := fmt.Sprintf("/characters/%d/contracts/%d/bids/", ctx.CharacterID, c.ContractID)
			var bids []ContractBid
			if err := ctx.Client.Get(bgCtx, bidPath, ctx.AccessToken, &bids); err != nil {
				global.Logger.Warn("[ESI] 获取合同竞标失败",
					zap.Int64("contract_id", c.ContractID),
					zap.Error(err),
				)
			} else if len(bids) > 0 {
				bidsJSON, _ := json.Marshal(bids)
				s := string(bidsJSON)
				record.BidsJSON = &s
			}
		}

		// 活跃合同获取物品
		if c.Status == "outstanding" || c.Status == "in_progress" {
			itemPath := fmt.Sprintf("/characters/%d/contracts/%d/items/", ctx.CharacterID, c.ContractID)
			var items []ContractItem
			if err := ctx.Client.Get(bgCtx, itemPath, ctx.AccessToken, &items); err != nil {
				global.Logger.Warn("[ESI] 获取合同物品失败",
					zap.Int64("contract_id", c.ContractID),
					zap.Error(err),
				)
			} else if len(items) > 0 {
				itemsJSON, _ := json.Marshal(items)
				s := string(itemsJSON)
				record.ItemsJSON = &s
			}
		}

		// Upsert
		var existing model.EveCharacterContract
		result := global.DB.Where("character_id = ? AND contract_id = ?", ctx.CharacterID, c.ContractID).First(&existing)
		if result.Error != nil {
			if err := global.DB.Create(&record).Error; err != nil {
				global.Logger.Warn("[ESI] 合同入库失败",
					zap.Int64("contract_id", c.ContractID),
					zap.Error(err),
				)
			}
		} else {
			record.ID = existing.ID
			if err := global.DB.Save(&record).Error; err != nil {
				global.Logger.Warn("[ESI] 合同更新失败",
					zap.Int64("contract_id", c.ContractID),
					zap.Error(err),
				)
			}
		}
	}

	global.Logger.Debug("[ESI] 角色合同入库完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(contracts)),
	)

	return nil
}
