package service

import (
	"amiya-eden/internal/repository"
)

type CorpStructureService struct {
	repo     *repository.CorpStructureRepository
	charRepo *repository.EveCharacterRepository
}

func NewCorpStructureService() *CorpStructureService {
	return &CorpStructureService{
		repo:     repository.NewCorpStructureRepository(),
		charRepo: repository.NewEveCharacterRepository(),
	}
}

// CorpStructureListRequest 建筑列表请求
type CorpStructureListRequest struct {
	Current         int    `json:"current"           binding:"required,min=1"`
	Size            int    `json:"size"              binding:"required,min=1,max=100"`
	CorpID          int64  `json:"corp_id"`
	State           string `json:"state"`
	FuelExpiresSoon bool   `json:"fuel_expires_soon"`
}

// ListCorpStructures 获取用户可见的军团建筑列表
func (s *CorpStructureService) ListCorpStructures(userID uint, req *CorpStructureListRequest) (interface{}, error) {
	corpID := req.CorpID

	// 如果未指定军团 ID，取用户第一个角色的军团
	if corpID == 0 {
		corpIDs, err := s.repo.GetCorpIDsByUserID(userID)
		if err != nil {
			return nil, err
		}
		if len(corpIDs) == 0 {
			return map[string]interface{}{
				"list":     []interface{}{},
				"total":    0,
				"page":     req.Current,
				"pageSize": req.Size,
			}, nil
		}
		corpID = corpIDs[0]
	}

	list, total, err := s.repo.ListByCorpID(corpID, req.Current, req.Size, req.State, req.FuelExpiresSoon)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     req.Current,
		"pageSize": req.Size,
	}, nil
}

// GetUserCorpIDs 获取用户关联的所有军团 ID
func (s *CorpStructureService) GetUserCorpIDs(userID uint) ([]int64, error) {
	return s.repo.GetCorpIDsByUserID(userID)
}
