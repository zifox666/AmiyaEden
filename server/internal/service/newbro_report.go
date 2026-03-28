package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"sort"
	"strings"
	"time"
)

type CaptainAttributionListRequest struct {
	Page         int
	PageSize     int
	PlayerUserID *uint
	RefType      string
	StartDate    *time.Time
	EndDate      *time.Time
}

type AdminAffiliationHistoryListRequest struct {
	Page            int
	PageSize        int
	CaptainSearch   string
	PlayerSearch    string
	ChangeStartDate *time.Time
	ChangeEndDate   *time.Time
}

type AdminCaptainDetail struct {
	Overview           CaptainOverview              `json:"overview"`
	Players            []CaptainPlayerListItem      `json:"players"`
	PlayersTotal       int64                        `json:"players_total"`
	Attributions       []CaptainAttributionListItem `json:"attributions"`
	AttributionsTotal  int64                        `json:"attributions_total"`
	AttributionSummary CaptainAttributionSummary    `json:"attribution_summary"`
}

type AdminAffiliationHistoryItem struct {
	AffiliationID          uint       `json:"affiliation_id"`
	PlayerUserID           uint       `json:"player_user_id"`
	PlayerCharacterID      int64      `json:"player_character_id"`
	PlayerCharacterName    string     `json:"player_character_name"`
	PlayerNickname         string     `json:"player_nickname"`
	CaptainUserID          uint       `json:"captain_user_id"`
	CaptainCharacterID     int64      `json:"captain_character_id"`
	CaptainCharacterName   string     `json:"captain_character_name"`
	CaptainNickname        string     `json:"captain_nickname"`
	ChangedByCharacterName string     `json:"changed_by_character_name"`
	StartedAt              time.Time  `json:"started_at"`
	EndedAt                *time.Time `json:"ended_at"`
	CreatedAt              time.Time  `json:"created_at"`
}

type captainProfile struct {
	Nickname             string
	PrimaryCharacterID   int64
	PrimaryCharacterName string
}

type NewbroReportService struct {
	roleRepo       *repository.RoleRepository
	userRepo       *repository.UserRepository
	charRepo       *repository.EveCharacterRepository
	affRepo        *repository.NewbroCaptainAffiliationRepository
	attrRepo       *repository.CaptainBountyAttributionRepository
	settlementRepo *repository.CaptainRewardSettlementRepository
}

func NewNewbroReportService() *NewbroReportService {
	return &NewbroReportService{
		roleRepo:       repository.NewRoleRepository(),
		userRepo:       repository.NewUserRepository(),
		charRepo:       repository.NewEveCharacterRepository(),
		affRepo:        repository.NewNewbroCaptainAffiliationRepository(),
		attrRepo:       repository.NewCaptainBountyAttributionRepository(),
		settlementRepo: repository.NewCaptainRewardSettlementRepository(),
	}
}

func (s *NewbroReportService) GetCaptainOverview(captainUserID uint) (*CaptainOverview, error) {
	activeCount, err := s.affRepo.CountDistinctPlayersByCaptainUserID(captainUserID, true)
	if err != nil {
		return nil, err
	}
	historicalCount, err := s.affRepo.CountDistinctPlayersByCaptainUserID(captainUserID, false)
	if err != nil {
		return nil, err
	}
	refTypes := supportedPlayerAttributionRefTypeList()
	bountyTotal, recordCount, err := s.attrRepo.SumByCaptainUserID(captainUserID, refTypes)
	if err != nil {
		return nil, err
	}
	profile, err := s.getCaptainProfileByUserID(captainUserID)
	if err != nil {
		return nil, err
	}
	return buildCaptainOverview(captainUserID, profile, activeCount, historicalCount, bountyTotal, recordCount), nil
}

func (s *NewbroReportService) ListCaptainPlayers(captainUserID uint, status string, page, pageSize int) ([]CaptainPlayerListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	rows, total, err := s.affRepo.ListByCaptainUserID(captainUserID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]uint, 0, len(rows))
	charIDs := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		userIDs = append(userIDs, row.PlayerUserID)
		charIDs = append(charIDs, row.PlayerPrimaryCharacterIDAtStart)
	}
	users, err := s.userRepo.ListByIDs(userIDs)
	if err != nil {
		return nil, 0, err
	}
	userByID := make(map[uint]model.User, len(users))
	for _, user := range users {
		userByID[user.ID] = user
		if user.PrimaryCharacterID != 0 {
			charIDs = append(charIDs, user.PrimaryCharacterID)
		}
	}
	chars, err := s.charRepo.ListByCharacterIDs(charIDs)
	if err != nil {
		return nil, 0, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}
	refTypes := supportedPlayerAttributionRefTypeList()
	items, err := buildCaptainPlayerListItems(rows, userByID, charByID, captainUserID, func(captainUserID, playerUserID uint) (float64, error) {
		return s.attrRepo.SumByCaptainAndPlayerUserID(captainUserID, playerUserID, refTypes)
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *NewbroReportService) ListCaptainAttributions(captainUserID uint, req CaptainAttributionListRequest) (CaptainAttributionSummary, []CaptainAttributionListItem, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	bountyTotal, recordCount, err := s.attrRepo.SummarizeByCaptainUserIDFiltered(
		captainUserID,
		req.PlayerUserID,
		req.RefType,
		req.StartDate,
		req.EndDate,
		supportedPlayerAttributionRefTypeList(),
	)
	if err != nil {
		return CaptainAttributionSummary{}, nil, 0, err
	}
	rows, total, err := s.attrRepo.ListByCaptainUserIDFiltered(
		captainUserID,
		req.Page,
		req.PageSize,
		req.PlayerUserID,
		req.RefType,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		return CaptainAttributionSummary{}, nil, 0, err
	}
	charIDs := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		charIDs = append(charIDs, row.PlayerCharacterID, row.CaptainCharacterID)
	}
	chars, err := s.charRepo.ListByCharacterIDs(charIDs)
	if err != nil {
		return CaptainAttributionSummary{}, nil, 0, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}
	items := make([]CaptainAttributionListItem, 0, len(rows))
	for _, row := range rows {
		playerChar := charByID[row.PlayerCharacterID]
		captainChar := charByID[row.CaptainCharacterID]
		items = append(items, CaptainAttributionListItem{
			ID:                     row.ID,
			PlayerUserID:           row.PlayerUserID,
			PlayerCharacterID:      row.PlayerCharacterID,
			PlayerCharacterName:    playerChar.CharacterName,
			CaptainCharacterID:     row.CaptainCharacterID,
			CaptainCharacterName:   captainChar.CharacterName,
			CaptainWalletJournalID: row.CaptainWalletJournalID,
			WalletJournalID:        row.WalletJournalID,
			RefType:                row.RefType,
			SystemID:               row.SystemID,
			JournalAt:              row.JournalAt,
			Amount:                 row.Amount,
			ProcessedAt:            row.ProcessedAt,
		})
	}
	return CaptainAttributionSummary{
		AttributedBountyTotal: bountyTotal,
		RecordCount:           recordCount,
	}, items, total, nil
}

func (s *NewbroReportService) ListCaptainRewardSettlements(
	captainUserID uint,
	page,
	pageSize int,
) (CaptainRewardSummary, []CaptainRewardSettlementItem, int64, error) {
	return s.listRewardSettlements(&captainUserID, page, pageSize)
}

func (s *NewbroReportService) ListAdminRewardSettlements(
	page,
	pageSize int,
) (CaptainRewardSummary, []CaptainRewardSettlementItem, int64, error) {
	return s.listRewardSettlements(nil, page, pageSize)
}

func (s *NewbroReportService) listRewardSettlements(
	captainUserID *uint,
	page,
	pageSize int,
) (CaptainRewardSummary, []CaptainRewardSettlementItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}

	settlementCount, totalCreditedValue, lastProcessedAt, err := s.settlementRepo.Summarize(captainUserID)
	if err != nil {
		return CaptainRewardSummary{}, nil, 0, err
	}
	rows, total, err := s.settlementRepo.List(captainUserID, page, pageSize)
	if err != nil {
		return CaptainRewardSummary{}, nil, 0, err
	}

	captainUserIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		captainUserIDs = append(captainUserIDs, row.CaptainUserID)
	}
	profiles, err := s.loadCaptainProfiles(captainUserIDs)
	if err != nil {
		return CaptainRewardSummary{}, nil, 0, err
	}

	return CaptainRewardSummary{
		SettlementCount:    settlementCount,
		TotalCreditedValue: totalCreditedValue,
		LastProcessedAt:    lastProcessedAt,
	}, buildCaptainRewardSettlementItems(rows, profiles), total, nil
}

func (s *NewbroReportService) ListAllCaptainOverviews(page, pageSize int, keyword string) ([]CaptainOverview, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userIDs, err := s.roleRepo.GetRoleUserIDs(model.RoleCaptain)
	if err != nil {
		return nil, 0, err
	}
	overviews := make([]CaptainOverview, 0, len(userIDs))
	for _, userID := range userIDs {
		overview, err := s.GetCaptainOverview(userID)
		if err != nil {
			return nil, 0, err
		}
		if keyword != "" && !strings.Contains(strings.ToLower(overview.CaptainCharacterName), strings.ToLower(keyword)) {
			continue
		}
		overviews = append(overviews, *overview)
	}
	sort.Slice(overviews, func(i, j int) bool {
		if overviews[i].AttributedBountyTotal != overviews[j].AttributedBountyTotal {
			return overviews[i].AttributedBountyTotal > overviews[j].AttributedBountyTotal
		}
		return overviews[i].CaptainUserID < overviews[j].CaptainUserID
	})
	total := int64(len(overviews))
	start := (page - 1) * pageSize
	if start >= len(overviews) {
		return []CaptainOverview{}, total, nil
	}
	end := start + pageSize
	if end > len(overviews) {
		end = len(overviews)
	}
	return overviews[start:end], total, nil
}

func (s *NewbroReportService) GetAdminCaptainDetail(captainUserID uint) (*AdminCaptainDetail, error) {
	overview, err := s.GetCaptainOverview(captainUserID)
	if err != nil {
		return nil, err
	}
	players, playersTotal, err := s.ListCaptainPlayers(captainUserID, "all", 1, 20)
	if err != nil {
		return nil, err
	}
	summary, attributions, attributionsTotal, err := s.ListCaptainAttributions(captainUserID, CaptainAttributionListRequest{Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	return &AdminCaptainDetail{
		Overview:           *overview,
		Players:            players,
		PlayersTotal:       playersTotal,
		Attributions:       attributions,
		AttributionsTotal:  attributionsTotal,
		AttributionSummary: summary,
	}, nil
}

func (s *NewbroReportService) ListAdminAffiliationHistory(req AdminAffiliationHistoryListRequest) ([]AdminAffiliationHistoryItem, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	rows, total, err := s.affRepo.ListAdminAffiliationHistory(repository.AdminAffiliationHistoryFilter{
		CaptainSearch:       req.CaptainSearch,
		PlayerSearch:        req.PlayerSearch,
		ChangeStartedAtFrom: req.ChangeStartDate,
		ChangeStartedAtTo:   req.ChangeEndDate,
	}, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []AdminAffiliationHistoryItem{}, total, nil
	}

	captainUserIDs := make([]uint, 0, len(rows))
	actorUserIDs := make([]uint, 0, len(rows))
	playerUserIDs := make([]uint, 0, len(rows))
	characterIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		captainUserIDs = append(captainUserIDs, row.CaptainUserID)
		actorUserIDs = append(actorUserIDs, row.CreatedBy)
		playerUserIDs = append(playerUserIDs, row.PlayerUserID)
		characterIDs = append(characterIDs, row.PlayerPrimaryCharacterIDAtStart)
	}

	captainProfiles, err := s.loadCaptainProfiles(captainUserIDs)
	if err != nil {
		return nil, 0, err
	}
	actorProfiles, err := s.loadCaptainProfiles(actorUserIDs)
	if err != nil {
		return nil, 0, err
	}
	users, err := s.userRepo.ListByIDs(playerUserIDs)
	if err != nil {
		return nil, 0, err
	}
	userByID := make(map[uint]model.User, len(users))
	for _, user := range users {
		userByID[user.ID] = user
	}
	chars, err := s.charRepo.ListByCharacterIDs(characterIDs)
	if err != nil {
		return nil, 0, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}

	return buildAdminAffiliationHistoryItems(rows, captainProfiles, actorProfiles, userByID, charByID), total, nil
}

func (s *NewbroReportService) getPrimaryCharacterByUserID(userID uint) (model.EveCharacter, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return model.EveCharacter{}, err
	}
	if user.PrimaryCharacterID == 0 {
		return model.EveCharacter{}, nil
	}
	char, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
	if err != nil {
		return model.EveCharacter{}, err
	}
	return *char, nil
}

func (s *NewbroReportService) getCaptainProfileByUserID(userID uint) (captainProfile, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return captainProfile{}, err
	}
	char, err := s.getPrimaryCharacterByUserID(userID)
	if err != nil {
		return captainProfile{}, err
	}
	return captainProfile{
		Nickname:             user.Nickname,
		PrimaryCharacterID:   char.CharacterID,
		PrimaryCharacterName: char.CharacterName,
	}, nil
}

func (s *NewbroReportService) loadCaptainProfiles(userIDs []uint) (map[uint]captainProfile, error) {
	return loadCaptainProfiles(s.userRepo, s.charRepo, userIDs)
}

func buildCaptainOverview(
	captainUserID uint,
	profile captainProfile,
	activeCount, historicalCount int64,
	bountyTotal float64,
	recordCount int64,
) *CaptainOverview {
	return &CaptainOverview{
		CaptainUserID:          captainUserID,
		CaptainCharacterID:     profile.PrimaryCharacterID,
		CaptainCharacterName:   profile.PrimaryCharacterName,
		CaptainNickname:        profile.Nickname,
		ActivePlayerCount:      activeCount,
		HistoricalPlayerCount:  historicalCount,
		AttributedBountyTotal:  bountyTotal,
		AttributionRecordCount: recordCount,
	}
}

func buildCaptainPlayerListItems(
	rows []model.NewbroCaptainAffiliation,
	userByID map[uint]model.User,
	charByID map[int64]model.EveCharacter,
	captainUserID uint,
	sumTotals func(captainUserID, playerUserID uint) (float64, error),
) ([]CaptainPlayerListItem, error) {
	items := make([]CaptainPlayerListItem, 0, len(rows))
	for _, row := range rows {
		bountyTotal, err := sumTotals(captainUserID, row.PlayerUserID)
		if err != nil {
			return nil, err
		}

		user := userByID[row.PlayerUserID]
		characterID := row.PlayerPrimaryCharacterIDAtStart
		char := charByID[characterID]
		if user.PrimaryCharacterID != 0 {
			if currentChar, ok := charByID[user.PrimaryCharacterID]; ok {
				characterID = user.PrimaryCharacterID
				char = currentChar
			}
		}

		items = append(items, CaptainPlayerListItem{
			PlayerUserID:          row.PlayerUserID,
			PlayerCharacterID:     characterID,
			PlayerCharacterName:   char.CharacterName,
			PlayerNickname:        user.Nickname,
			PlayerPortraitURL:     char.PortraitURL,
			StartedAt:             row.StartedAt,
			EndedAt:               row.EndedAt,
			AttributedBountyTotal: bountyTotal,
		})
	}
	return items, nil
}

func buildAdminAffiliationHistoryItems(
	rows []model.NewbroCaptainAffiliation,
	captainProfiles map[uint]captainProfile,
	actorProfiles map[uint]captainProfile,
	userByID map[uint]model.User,
	charByID map[int64]model.EveCharacter,
) []AdminAffiliationHistoryItem {
	items := make([]AdminAffiliationHistoryItem, 0, len(rows))
	for _, row := range rows {
		captain := captainProfiles[row.CaptainUserID]
		actor := actorProfiles[row.CreatedBy]
		player := userByID[row.PlayerUserID]
		playerChar := charByID[row.PlayerPrimaryCharacterIDAtStart]

		items = append(items, AdminAffiliationHistoryItem{
			AffiliationID:          row.ID,
			PlayerUserID:           row.PlayerUserID,
			PlayerCharacterID:      row.PlayerPrimaryCharacterIDAtStart,
			PlayerCharacterName:    playerChar.CharacterName,
			PlayerNickname:         player.Nickname,
			CaptainUserID:          row.CaptainUserID,
			CaptainCharacterID:     captain.PrimaryCharacterID,
			CaptainCharacterName:   captain.PrimaryCharacterName,
			CaptainNickname:        captain.Nickname,
			ChangedByCharacterName: actor.PrimaryCharacterName,
			StartedAt:              row.StartedAt,
			EndedAt:                row.EndedAt,
			CreatedAt:              row.CreatedAt,
		})
	}
	return items
}

func buildCaptainRewardSettlementItems(
	rows []model.CaptainRewardSettlement,
	captainProfiles map[uint]captainProfile,
) []CaptainRewardSettlementItem {
	items := make([]CaptainRewardSettlementItem, 0, len(rows))
	for _, row := range rows {
		profile := captainProfiles[row.CaptainUserID]
		items = append(items, CaptainRewardSettlementItem{
			ID:                   row.ID,
			CaptainUserID:        row.CaptainUserID,
			CaptainCharacterID:   profile.PrimaryCharacterID,
			CaptainCharacterName: profile.PrimaryCharacterName,
			CaptainNickname:      profile.Nickname,
			AttributionCount:     row.AttributionCount,
			AttributedISKTotal:   row.AttributedISKTotal,
			BonusRate:            row.BonusRate,
			CreditedValue:        row.CreditedValue,
			ProcessedAt:          row.ProcessedAt,
		})
	}
	return items
}
