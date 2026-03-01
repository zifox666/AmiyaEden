package service

import (
	"errors"

	"amiya-eden/internal/repository"
)

// EveInfoService EVE 角色信息业务逻辑层
type EveInfoService struct {
	charRepo   *repository.EveCharacterRepository
	walletRepo *repository.EveWalletRepository
	skillRepo  *repository.EveSkillRepository
	sdeRepo    *repository.SdeRepository
}

func NewEveInfoService() *EveInfoService {
	return &EveInfoService{
		charRepo:   repository.NewEveCharacterRepository(),
		walletRepo: repository.NewEveWalletRepository(),
		skillRepo:  repository.NewEveSkillRepository(),
		sdeRepo:    repository.NewSdeRepository(),
	}
}

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// InfoWalletRequest 钱包流水请求
type InfoWalletRequest struct {
	CharacterID int64 `json:"character_id" binding:"required"`
	Page        int   `json:"page" binding:"required,min=1"`
	PageSize    int   `json:"page_size" binding:"required,min=1,max=100"`
}

// InfoWalletResponse 钱包流水响应
type InfoWalletResponse struct {
	Balance  float64             `json:"balance"`
	Journals []InfoWalletJournal `json:"journals"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// InfoWalletJournal 钱包流水条目
type InfoWalletJournal struct {
	ID            int64   `json:"id"`
	Amount        float64 `json:"amount"`
	Balance       float64 `json:"balance"`
	Date          string  `json:"date"`
	Description   string  `json:"description"`
	FirstPartyID  int64   `json:"first_party_id"`
	SecondPartyID int64   `json:"second_party_id"`
	RefType       string  `json:"ref_type"`
	Reason        string  `json:"reason"`
}

// InfoSkillRequest 技能列表请求
type InfoSkillRequest struct {
	CharacterID int64  `json:"character_id" binding:"required"`
	Language    string `json:"language"`
}

// InfoSkillItem 技能条目（含名称）
type InfoSkillItem struct {
	SkillID            int    `json:"skill_id"`
	SkillName          string `json:"skill_name"`
	GroupID            int    `json:"group_id"`
	GroupName          string `json:"group_name"`
	ActiveLevel        int    `json:"active_level"`
	TrainedLevel       int    `json:"trained_level"`
	SkillpointsInSkill int64  `json:"skillpoints_in_skill"`
}

// InfoSkillQueueItem 技能队列条目（含名称）
type InfoSkillQueueItem struct {
	QueuePosition   int    `json:"queue_position"`
	SkillID         int    `json:"skill_id"`
	SkillName       string `json:"skill_name"`
	FinishedLevel   int    `json:"finished_level"`
	LevelStartSP    int64  `json:"level_start_sp"`
	LevelEndSP      int64  `json:"level_end_sp"`
	TrainingStartSP int64  `json:"training_start_sp"`
	StartDate       int64  `json:"start_date"`
	FinishDate      int64  `json:"finish_date"`
}

// InfoSkillResponse 技能列表响应
type InfoSkillResponse struct {
	TotalSP       int64                `json:"total_sp"`
	UnallocatedSP int64                `json:"unallocated_sp"`
	SkillCount    int                  `json:"skill_count"`
	Skills        []InfoSkillItem      `json:"skills"`
	SkillQueue    []InfoSkillQueueItem `json:"skill_queue"`
}

// ─────────────────────────────────────────────
//  业务方法
// ─────────────────────────────────────────────

// validateCharacterOwnership 校验角色归属
func (s *EveInfoService) validateCharacterOwnership(userID uint, characterID int64) error {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return errors.New("获取角色列表失败")
	}
	for _, c := range chars {
		if c.CharacterID == characterID {
			return nil
		}
	}
	return errors.New("该角色不属于当前用户")
}

// GetWalletJournal 获取指定角色的钱包流水
func (s *EveInfoService) GetWalletJournal(userID uint, req *InfoWalletRequest) (*InfoWalletResponse, error) {
	// 校验角色归属
	if err := s.validateCharacterOwnership(userID, req.CharacterID); err != nil {
		return nil, err
	}

	result := &InfoWalletResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 获取余额
	wallet, err := s.walletRepo.GetWallet(int(req.CharacterID))
	if err == nil {
		result.Balance = wallet.Balance
	}

	// 获取流水
	journals, total, err := s.walletRepo.GetWalletJournals(req.CharacterID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	result.Total = total
	result.Journals = make([]InfoWalletJournal, 0, len(journals))
	for _, j := range journals {
		result.Journals = append(result.Journals, InfoWalletJournal{
			ID:            j.ID,
			Amount:        j.Amount,
			Balance:       j.Balance,
			Date:          j.Date.Format("2006-01-02 15:04:05"),
			Description:   j.Description,
			FirstPartyID:  j.FirstPartyID,
			SecondPartyID: j.SecondPartyID,
			RefType:       j.RefType,
			Reason:        j.Reason,
		})
	}

	return result, nil
}

// GetCharacterSkills 获取指定角色的技能列表与队列
func (s *EveInfoService) GetCharacterSkills(userID uint, req *InfoSkillRequest) (*InfoSkillResponse, error) {
	// 校验角色归属
	if err := s.validateCharacterOwnership(userID, req.CharacterID); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	result := &InfoSkillResponse{}

	// 获取技能总览
	skill, err := s.skillRepo.GetSkill(int(req.CharacterID))
	if err == nil {
		result.TotalSP = skill.TotalSP
		result.UnallocatedSP = skill.UnallocatedSP
	}

	// 获取技能列表
	skillList, err := s.skillRepo.GetSkillList(int(req.CharacterID))
	if err != nil {
		return nil, err
	}

	// 收集所有 skill_id 用于 SDE 查询（skill_id = type_id）
	skillIDs := make([]int, 0, len(skillList))
	for _, sk := range skillList {
		skillIDs = append(skillIDs, sk.SkillID)
	}

	// 通过 SDE 获取技能名称 (skill_id = type_id)
	published := true
	typeInfoMap := make(map[int]struct {
		TypeName, GroupName string
		GroupID             int
	})
	if len(skillIDs) > 0 {
		typeInfos, err := s.sdeRepo.GetTypes(skillIDs, &published, lang)
		if err == nil {
			for _, t := range typeInfos {
				typeInfoMap[t.TypeID] = struct {
					TypeName, GroupName string
					GroupID             int
				}{
					TypeName:  t.TypeName,
					GroupName: t.GroupName,
					GroupID:   t.GroupID,
				}
			}
		}
	}

	result.Skills = make([]InfoSkillItem, 0, len(skillList))
	for _, sk := range skillList {
		item := InfoSkillItem{
			SkillID:            sk.SkillID,
			ActiveLevel:        sk.ActiveLevel,
			TrainedLevel:       sk.TrainedLevel,
			SkillpointsInSkill: sk.SkillpointsInSkill,
		}
		if info, ok := typeInfoMap[sk.SkillID]; ok {
			item.SkillName = info.TypeName
			item.GroupName = info.GroupName
			item.GroupID = info.GroupID
		}
		result.Skills = append(result.Skills, item)
	}
	result.SkillCount = len(result.Skills)

	// 获取技能队列
	queue, err := s.skillRepo.GetSkillQueue(int(req.CharacterID))
	if err == nil {
		// 收集队列中的 skill_id
		queueSkillIDs := make([]int, 0, len(queue))
		for _, q := range queue {
			if _, ok := typeInfoMap[q.SkillID]; !ok {
				queueSkillIDs = append(queueSkillIDs, q.SkillID)
			}
		}
		// 查询队列中尚未查到名称的 skill
		if len(queueSkillIDs) > 0 {
			queueTypeInfos, err := s.sdeRepo.GetTypes(queueSkillIDs, &published, lang)
			if err == nil {
				for _, t := range queueTypeInfos {
					typeInfoMap[t.TypeID] = struct {
						TypeName, GroupName string
						GroupID             int
					}{
						TypeName:  t.TypeName,
						GroupName: t.GroupName,
						GroupID:   t.GroupID,
					}
				}
			}
		}

		result.SkillQueue = make([]InfoSkillQueueItem, 0, len(queue))
		for _, q := range queue {
			item := InfoSkillQueueItem{
				QueuePosition:   q.QueuePosition,
				SkillID:         q.SkillID,
				FinishedLevel:   q.FinishedLevel,
				LevelStartSP:    q.LevelStartSP,
				LevelEndSP:      q.LevelEndSP,
				TrainingStartSP: q.TrainingStartSP,
				StartDate:       q.StartDate,
				FinishDate:      q.FinishDate,
			}
			if info, ok := typeInfoMap[q.SkillID]; ok {
				item.SkillName = info.TypeName
			}
			result.SkillQueue = append(result.SkillQueue, item)
		}
	}

	return result, nil
}
