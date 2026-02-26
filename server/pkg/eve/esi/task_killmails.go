package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────
//  Character Killmails 角色击杀邮件
//  GET /characters/{character_id}/killmails/recent
//  GET /killmails/{killmail_id}/{killmail_hash}  (详情)
//  默认刷新间隔: 20 Minutes / 不活跃: 3 Days
// ─────────────────────────────────────────────

func init() {
	Register(&KillmailsTask{})
}

// KillmailsTask 角色击杀邮件刷新任务
type KillmailsTask struct{}

func (t *KillmailsTask) Name() string        { return "character_killmails" }
func (t *KillmailsTask) Description() string { return "角色击杀/损失邮件" }
func (t *KillmailsTask) Priority() Priority  { return PriorityCritical } // 高频关键任务

func (t *KillmailsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   20 * time.Minute,
		Inactive: 3 * 24 * time.Hour,
	}
}

func (t *KillmailsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-killmails.read_killmails.v1", Description: "读取击杀邮件"},
	}
}

// KillmailRef 击杀邮件引用（recent 接口返回）
type KillmailRef struct {
	KillmailHash string `json:"killmail_hash"`
	KillmailID   int64  `json:"killmail_id"`
}

// KillmailDetail 击杀邮件详情
type KillmailDetail struct {
	Attackers []struct {
		AllianceID     *int64  `json:"alliance_id,omitempty"`
		CharacterID    *int64  `json:"character_id,omitempty"`
		CorporationID  *int64  `json:"corporation_id,omitempty"`
		DamageDone     int     `json:"damage_done"`
		FactionID      *int64  `json:"faction_id,omitempty"`
		FinalBlow      bool    `json:"final_blow"`
		SecurityStatus float64 `json:"security_status"`
		ShipTypeID     *int    `json:"ship_type_id,omitempty"`
		WeaponTypeID   *int    `json:"weapon_type_id,omitempty"`
	} `json:"attackers"`
	KillmailID    int64     `json:"killmail_id"`
	KillmailTime  time.Time `json:"killmail_time"`
	MoonID        *int64    `json:"moon_id,omitempty"`
	SolarSystemID int64     `json:"solar_system_id"`
	Victim        struct {
		AllianceID    *int64 `json:"alliance_id,omitempty"`
		CharacterID   *int64 `json:"character_id,omitempty"`
		CorporationID *int64 `json:"corporation_id,omitempty"`
		DamageTaken   int    `json:"damage_taken"`
		FactionID     *int64 `json:"faction_id,omitempty"`
		Items         []struct {
			Flag              int  `json:"flag"`
			ItemTypeID        int  `json:"item_type_id"`
			QuantityDestroyed *int `json:"quantity_destroyed,omitempty"`
			QuantityDropped   *int `json:"quantity_dropped,omitempty"`
			Singleton         int  `json:"singleton"`
		} `json:"items,omitempty"`
		Position *struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position,omitempty"`
		ShipTypeID int `json:"ship_type_id"`
	} `json:"victim"`
	WarID *int64 `json:"war_id,omitempty"`
}

func (t *KillmailsTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	// 1. 获取最近的 killmail 列表（自动分页）
	recentPath := fmt.Sprintf("/characters/%d/killmails/recent/", ctx.CharacterID)
	var refs []KillmailRef
	if _, err := ctx.Client.GetPaginated(bgCtx, recentPath, ctx.AccessToken, &refs); err != nil {
		return fmt.Errorf("fetch recent killmails: %w", err)
	}

	global.Logger.Debug("[ESI] 角色击杀邮件引用获取完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(refs)),
	)

	// 2. 逐个获取 killmail 详情并入库
	for _, ref := range refs {
		// 先检查数据库中是否已存在该 killmail
		var count int64
		global.DB.Model(&model.EveKillmailList{}).Where("kill_mail_id = ?", ref.KillmailID).Count(&count)
		if count > 0 {
			// 已存在，只需确保关联关系
			var linkCount int64
			global.DB.Model(&model.EveCharacterKillmail{}).
				Where("character_id = ? AND killmail_id = ?", ctx.CharacterID, ref.KillmailID).
				Count(&linkCount)
			if linkCount == 0 {
				// 查询主记录判断是否受害者
				var km model.EveKillmailList
				isVictim := false
				if err := global.DB.Where("kill_mail_id = ?", ref.KillmailID).First(&km).Error; err == nil {
					isVictim = km.CharacterID == ctx.CharacterID
				}
				global.DB.Create(&model.EveCharacterKillmail{
					CharacterID: ctx.CharacterID,
					KillmailID:  ref.KillmailID,
					Victim:      isVictim,
				})
			}
			continue
		}

		detailPath := fmt.Sprintf("/killmails/%d/%s/", ref.KillmailID, ref.KillmailHash)
		var detail KillmailDetail
		if err := ctx.Client.Get(bgCtx, detailPath, "", &detail); err != nil {
			global.Logger.Warn("[ESI] 获取 killmail 详情失败",
				zap.Int64("killmail_id", ref.KillmailID),
				zap.Error(err),
			)
			continue
		}

		// 提取 victim 信息
		var victimCharID, victimCorpID, victimAllianceID int64
		if detail.Victim.CharacterID != nil {
			victimCharID = *detail.Victim.CharacterID
		}
		if detail.Victim.CorporationID != nil {
			victimCorpID = *detail.Victim.CorporationID
		}
		if detail.Victim.AllianceID != nil {
			victimAllianceID = *detail.Victim.AllianceID
		}
		isVictim := victimCharID == ctx.CharacterID

		// 在事务中写入 killmail 主记录 + items + 关联
		err := global.DB.Transaction(func(tx *gorm.DB) error {
			km := model.EveKillmailList{
				KillmailID:    ref.KillmailID,
				KillmailHash:  ref.KillmailHash,
				KillmailTime:  detail.KillmailTime,
				SolarSystemID: detail.SolarSystemID,
				ShipTypeID:    int64(detail.Victim.ShipTypeID),
				CharacterID:   victimCharID,
				CorporationID: victimCorpID,
				AllianceID:    victimAllianceID,
			}
			if err := tx.Create(&km).Error; err != nil {
				return err
			}

			// 将 victim items 写入 eve_killmail_item 表（按掉落/损毁拆分成独立记录）
			if len(detail.Victim.Items) > 0 {
				var items []model.EveKillmailItem
				for _, it := range detail.Victim.Items {
					// 损毁的物品
					if it.QuantityDestroyed != nil && *it.QuantityDestroyed > 0 {
						dropType := false
						items = append(items, model.EveKillmailItem{
							KillmailID: ref.KillmailID,
							ItemID:     it.ItemTypeID,
							ItemNum:    int64(*it.QuantityDestroyed),
							DropType:   &dropType,
							Flag:       it.Flag,
						})
					}
					// 掉落的物品
					if it.QuantityDropped != nil && *it.QuantityDropped > 0 {
						dropType := true
						items = append(items, model.EveKillmailItem{
							KillmailID: ref.KillmailID,
							ItemID:     it.ItemTypeID,
							ItemNum:    int64(*it.QuantityDropped),
							DropType:   &dropType,
							Flag:       it.Flag,
						})
					}
				}
				if len(items) > 0 {
					if err := tx.Create(&items).Error; err != nil {
						return err
					}
				}
			}

			// 创建角色-killmail 关联
			return tx.Create(&model.EveCharacterKillmail{
				CharacterID: ctx.CharacterID,
				KillmailID:  ref.KillmailID,
				Victim:      isVictim,
			}).Error
		})

		if err != nil {
			global.Logger.Warn("[ESI] killmail 入库失败",
				zap.Int64("killmail_id", ref.KillmailID),
				zap.Error(err),
			)
			continue
		}

		global.Logger.Debug("[ESI] killmail 入库成功",
			zap.Int64("killmail_id", ref.KillmailID),
			zap.Int("items", len(detail.Victim.Items)),
			zap.Time("killmail_time", detail.KillmailTime),
		)
	}

	return nil
}
