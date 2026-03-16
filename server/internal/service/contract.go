package service

import (
	"amiya-eden/internal/repository"
	"errors"
	"time"
)

// ─────────────────────────────────────────────
//  请求 & 响应结构
// ─────────────────────────────────────────────

// InfoContractsRequest 合同列表请求（含分页与过滤）
type InfoContractsRequest struct {
	Current  int    `json:"current"`
	Size     int    `json:"size"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Language string `json:"language"`
}

// ContractBidItem 合同竞标条目
type ContractBidItem struct {
	Amount   float64   `json:"amount"`
	BidID    int64     `json:"bid_id"`
	BidderID int64     `json:"bidder_id"`
	DateBid  time.Time `json:"date_bid"`
}

// ContractItemDetail 合同物品条目（含类型名）
type ContractItemDetail struct {
	TypeID      int    `json:"type_id"`
	TypeName    string `json:"type_name"`
	GroupName   string `json:"group_name"`
	CategoryID  int    `json:"category_id"`
	Quantity    int    `json:"quantity"`
	IsIncluded  bool   `json:"is_included"`
	IsSingleton bool   `json:"is_singleton"`
}

// ContractResponse 单个合同响应（不含物品/竞标，通过详情接口获取）
type ContractResponse struct {
	CharacterID         int64      `json:"character_id"`
	CharacterName       string     `json:"character_name"`
	ContractID          int64      `json:"contract_id"`
	AcceptorID          int64      `json:"acceptor_id"`
	AssigneeID          int64      `json:"assignee_id"`
	Availability        string     `json:"availability"`
	Buyout              *float64   `json:"buyout,omitempty"`
	Collateral          *float64   `json:"collateral,omitempty"`
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

// InfoContractDetailRequest 合同详情请求
type InfoContractDetailRequest struct {
	CharacterID int64  `json:"character_id"`
	ContractID  int64  `json:"contract_id"`
	Language    string `json:"language"`
}

// ContractDetailResponse 合同详情响应（物品 + 竞标）
type ContractDetailResponse struct {
	Items []ContractItemDetail `json:"items"`
	Bids  []ContractBidItem    `json:"bids"`
}

// ─────────────────────────────────────────────
//  Service
// ─────────────────────────────────────────────

// ContractService 合同业务逻辑
type ContractService struct {
	charRepo     *repository.EveCharacterRepository
	contractRepo *repository.ContractRepository
	sdeRepo      *repository.SdeRepository
}

func NewContractService() *ContractService {
	return &ContractService{
		charRepo:     repository.NewEveCharacterRepository(),
		contractRepo: repository.NewContractRepository(),
		sdeRepo:      repository.NewSdeRepository(),
	}
}

// GetUserContracts 分页获取用户名下所有角色的合同
func (s *ContractService) GetUserContracts(userID uint, req *InfoContractsRequest) ([]ContractResponse, int64, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}
	page := req.Current
	if page <= 0 {
		page = 1
	}
	pageSize := req.Size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 1. 获取角色列表
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, 0, errors.New("获取角色列表失败")
	}
	if len(chars) == 0 {
		return []ContractResponse{}, 0, nil
	}

	charIDs := make([]int64, 0, len(chars))
	charNameMap := make(map[int64]string, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
		charNameMap[c.CharacterID] = c.CharacterName
	}

	// 2. 分页查询合同
	filter := repository.ContractFilter{Type: req.Type, Status: req.Status}
	dbContracts, total, err := s.contractRepo.ListContracts(page, pageSize, charIDs, filter)
	if err != nil {
		return nil, 0, errors.New("查询合同失败")
	}

	// 3. 组装响应
	result := make([]ContractResponse, 0, len(dbContracts))
	for _, c := range dbContracts {
		resp := ContractResponse{
			CharacterID:         c.CharacterID,
			CharacterName:       charNameMap[c.CharacterID],
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
		result = append(result, resp)
	}

	return result, total, nil
}

// GetContractDetail 获取指定合同的物品与竞标详情
func (s *ContractService) GetContractDetail(userID uint, req *InfoContractDetailRequest) (*ContractDetailResponse, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	// 1. 验证角色属于当前用户
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取角色列表失败")
	}
	owned := false
	for _, c := range chars {
		if c.CharacterID == req.CharacterID {
			owned = true
			break
		}
	}
	if !owned {
		return nil, errors.New("无权访问该合同")
	}

	// 2. 查询合同记录（权限校验通过后只需确认合同存在）
	_, err = s.contractRepo.GetContractByCharacterAndID(req.CharacterID, req.ContractID)
	if err != nil {
		return nil, errors.New("合同不存在")
	}

	detail := &ContractDetailResponse{
		Items: []ContractItemDetail{},
		Bids:  []ContractBidItem{},
	}

	// 3. 查物品表
	rawItems, err := s.contractRepo.GetContractItems(req.ContractID)
	if err == nil && len(rawItems) > 0 {
		typeIDSet := make(map[int]struct{}, len(rawItems))
		for _, it := range rawItems {
			typeIDSet[it.TypeID] = struct{}{}
		}
		typeIDSlice := make([]int, 0, len(typeIDSet))
		for id := range typeIDSet {
			typeIDSlice = append(typeIDSlice, id)
		}
		typeInfoMap := make(map[int]repository.TypeInfo)
		if typeInfos, e := s.sdeRepo.GetTypes(typeIDSlice, nil, lang); e == nil {
			for _, ti := range typeInfos {
				typeInfoMap[ti.TypeID] = ti
			}
		}
		for _, it := range rawItems {
			ti := typeInfoMap[it.TypeID]
			detail.Items = append(detail.Items, ContractItemDetail{
				TypeID:      it.TypeID,
				TypeName:    ti.TypeName,
				GroupName:   ti.GroupName,
				CategoryID:  ti.CategoryID,
				Quantity:    it.Quantity,
				IsIncluded:  it.IsIncluded,
				IsSingleton: it.IsSingleton,
			})
		}
	}

	// 4. 查竞标表
	rawBids, err := s.contractRepo.GetContractBids(req.ContractID)
	if err == nil {
		for _, b := range rawBids {
			detail.Bids = append(detail.Bids, ContractBidItem{
				Amount:   b.Amount,
				BidID:    b.BidID,
				BidderID: b.BidderID,
				DateBid:  b.DateBid,
			})
		}
	}

	return detail, nil
}
