package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"
)

const (
	newbroAttributionLookbackDays    = 30
	newbroAttributionWindow          = 5 * time.Minute
	captainAttributionSyncKey        = "default"
	captainAttributionSyncFetchLimit = 500
)

type NewbroCaptainCandidate struct {
	CaptainUserID        uint       `json:"captain_user_id"`
	CaptainCharacterID   int64      `json:"captain_character_id"`
	CaptainCharacterName string     `json:"captain_character_name"`
	CaptainNickname      string     `json:"captain_nickname"`
	ActiveNewbroCount    int64      `json:"active_newbro_count"`
	LastOnlineAt         *time.Time `json:"last_online_at"`
}

type NewbroAffiliationSummary struct {
	AffiliationID        uint       `json:"affiliation_id"`
	CaptainUserID        uint       `json:"captain_user_id"`
	CaptainCharacterID   int64      `json:"captain_character_id"`
	CaptainCharacterName string     `json:"captain_character_name"`
	StartedAt            time.Time  `json:"started_at"`
	EndedAt              *time.Time `json:"ended_at"`
}

type NewbroMyAffiliationResponse struct {
	IsCurrentlyNewbro  bool                       `json:"is_currently_newbro"`
	EvaluatedAt        time.Time                  `json:"evaluated_at"`
	RuleVersion        string                     `json:"rule_version"`
	DisqualifiedReason string                     `json:"disqualified_reason"`
	CurrentAffiliation *NewbroAffiliationSummary  `json:"current_affiliation"`
	RecentAffiliations []NewbroAffiliationSummary `json:"recent_affiliations"`
}

type SelectCaptainResponse struct {
	AffiliationID uint      `json:"affiliation_id"`
	CaptainUserID uint      `json:"captain_user_id"`
	StartedAt     time.Time `json:"started_at"`
}

type CaptainEligiblePlayerCurrentAffiliation struct {
	AffiliationID        uint      `json:"affiliation_id"`
	CaptainUserID        uint      `json:"captain_user_id"`
	CaptainCharacterID   int64     `json:"captain_character_id"`
	CaptainCharacterName string    `json:"captain_character_name"`
	CaptainNickname      string    `json:"captain_nickname"`
	StartedAt            time.Time `json:"started_at"`
}

type CaptainEligiblePlayerListItem struct {
	PlayerUserID        uint                                     `json:"player_user_id"`
	PlayerCharacterID   int64                                    `json:"player_character_id"`
	PlayerCharacterName string                                   `json:"player_character_name"`
	PlayerNickname      string                                   `json:"player_nickname"`
	CurrentAffiliation  *CaptainEligiblePlayerCurrentAffiliation `json:"current_affiliation"`
}

type CaptainOverview struct {
	CaptainUserID          uint    `json:"captain_user_id"`
	CaptainCharacterID     int64   `json:"captain_character_id"`
	CaptainCharacterName   string  `json:"captain_character_name"`
	CaptainNickname        string  `json:"captain_nickname"`
	ActivePlayerCount      int64   `json:"active_player_count"`
	HistoricalPlayerCount  int64   `json:"historical_player_count"`
	AttributedBountyTotal  float64 `json:"attributed_bounty_total"`
	AttributionRecordCount int64   `json:"attribution_record_count"`
}

type CaptainPlayerListItem struct {
	PlayerUserID          uint       `json:"player_user_id"`
	PlayerCharacterID     int64      `json:"player_character_id"`
	PlayerCharacterName   string     `json:"player_character_name"`
	PlayerNickname        string     `json:"player_nickname"`
	StartedAt             time.Time  `json:"started_at"`
	EndedAt               *time.Time `json:"ended_at"`
	AttributedBountyTotal float64    `json:"attributed_bounty_total"`
}

type CaptainAttributionListItem struct {
	ID                     uint       `json:"id"`
	PlayerUserID           uint       `json:"player_user_id"`
	PlayerCharacterID      int64      `json:"player_character_id"`
	PlayerCharacterName    string     `json:"player_character_name"`
	CaptainCharacterID     int64      `json:"captain_character_id"`
	CaptainCharacterName   string     `json:"captain_character_name"`
	CaptainWalletJournalID int64      `json:"captain_wallet_journal_id"`
	WalletJournalID        int64      `json:"wallet_journal_id"`
	RefType                string     `json:"ref_type"`
	SystemID               int64      `json:"system_id"`
	JournalAt              time.Time  `json:"journal_at"`
	Amount                 float64    `json:"amount"`
	ProcessedAt            *time.Time `json:"processed_at"`
}

type CaptainAttributionSummary struct {
	AttributedBountyTotal float64 `json:"attributed_bounty_total"`
	RecordCount           int64   `json:"record_count"`
}

type CaptainAttributionSyncResult struct {
	ProcessedCount      int   `json:"processed_count"`
	InsertedCount       int   `json:"inserted_count"`
	SkippedCount        int   `json:"skipped_count"`
	LastWalletJournalID int64 `json:"last_wallet_journal_id"`
}

type CaptainRewardSettlementItem struct {
	ID                   uint      `json:"id"`
	CaptainUserID        uint      `json:"captain_user_id"`
	CaptainCharacterID   int64     `json:"captain_character_id"`
	CaptainCharacterName string    `json:"captain_character_name"`
	CaptainNickname      string    `json:"captain_nickname"`
	AttributionCount     int64     `json:"attribution_count"`
	AttributedISKTotal   float64   `json:"attributed_isk_total"`
	BonusRate            float64   `json:"bonus_rate"`
	CreditedValue        float64   `json:"credited_value"`
	ProcessedAt          time.Time `json:"processed_at"`
}

type CaptainRewardSummary struct {
	SettlementCount    int64      `json:"settlement_count"`
	TotalCreditedValue float64    `json:"total_credited_value"`
	LastProcessedAt    *time.Time `json:"last_processed_at"`
}

type CaptainRewardProcessResult struct {
	ProcessedAt               time.Time `json:"processed_at"`
	ProcessedCaptainCount     int       `json:"processed_captain_count"`
	ProcessedAttributionCount int       `json:"processed_attribution_count"`
	SettlementCount           int       `json:"settlement_count"`
	TotalCreditedValue        float64   `json:"total_credited_value"`
}

type NewbroEligibilityService struct {
	stateRepo              *repository.NewbroPlayerStateRepository
	charRepo               *repository.EveCharacterRepository
	skillRepo              *repository.EveSkillRepository
	settingsSvc            *NewbroSettingsService
	endAffiliationByUserID func(userID uint, endedAt time.Time) error
}

func NewNewbroEligibilityService() *NewbroEligibilityService {
	affRepo := repository.NewNewbroCaptainAffiliationRepository()
	return &NewbroEligibilityService{
		stateRepo:              repository.NewNewbroPlayerStateRepository(),
		charRepo:               repository.NewEveCharacterRepository(),
		skillRepo:              repository.NewEveSkillRepository(),
		settingsSvc:            NewNewbroSettingsService(),
		endAffiliationByUserID: affRepo.EndActiveByPlayerUserID,
	}
}

func (s *NewbroEligibilityService) CurrentSettings() NewbroSettings {
	return s.settingsSvc.GetSettings()
}

func (s *NewbroEligibilityService) CurrentRules() NewbroEligibilityRules {
	return s.CurrentSettings().ToEligibilityRules()
}

// GetCachedState returns the stored eligibility state without triggering a refresh.
// Use this on hot paths (e.g. GetMe, GetUserMenuTree) where blocking on a recalculation
// is unacceptable. Returns nil if no state has been computed yet.
func (s *NewbroEligibilityService) GetCachedState(userID uint) *model.NewbroPlayerState {
	state, err := s.stateRepo.GetByUserID(userID)
	if err != nil {
		return nil
	}
	return state
}

func (s *NewbroEligibilityService) EnsureCurrentState(userID uint) (*model.NewbroPlayerState, error) {
	state, err := s.stateRepo.GetByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	settings := s.CurrentSettings()
	rules := settings.ToEligibilityRules()
	ruleVersion := BuildNewbroRuleVersion(rules)
	if state != nil && !NeedsNewbroEligibilityRefresh(state, ruleVersion, time.Now(), settings.RefreshInterval()) {
		return state, nil
	}

	return s.recalculateState(userID, ruleVersion, rules)
}

func (s *NewbroEligibilityService) recalculateState(userID uint, ruleVersion string, rules NewbroEligibilityRules) (*model.NewbroPlayerState, error) {
	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	characterIDs := make([]int64, 0, len(characters))
	for _, character := range characters {
		characterIDs = append(characterIDs, character.CharacterID)
	}

	skillTotals, err := s.skillRepo.GetSkillTotalsByCharacterIDs(characterIDs)
	if err != nil {
		return nil, err
	}

	snapshots := make([]NewbroCharacterSnapshot, 0, len(characters))
	for _, character := range characters {
		snapshots = append(snapshots, NewbroCharacterSnapshot{
			CharacterID:   character.CharacterID,
			CorporationID: character.CorporationID,
			TotalSP:       skillTotals[character.CharacterID],
		})
	}

	result := EvaluateNewbroEligibility(snapshots, rules)
	evaluatedAt := time.Now()
	state := &model.NewbroPlayerState{
		UserID:             userID,
		IsCurrentlyNewbro:  result.IsCurrentlyNewbro,
		EvaluatedAt:        evaluatedAt,
		RuleVersion:        ruleVersion,
		DisqualifiedReason: result.DisqualifiedReason,
	}
	if err := s.stateRepo.Save(state); err != nil {
		return nil, err
	}
	if err := s.syncAffiliationWithEligibility(userID, result.IsCurrentlyNewbro, evaluatedAt); err != nil {
		return nil, err
	}
	return state, nil
}

func (s *NewbroEligibilityService) syncAffiliationWithEligibility(
	userID uint,
	isCurrentlyNewbro bool,
	evaluatedAt time.Time,
) error {
	if isCurrentlyNewbro || s.endAffiliationByUserID == nil {
		return nil
	}
	return s.endAffiliationByUserID(userID, evaluatedAt)
}

type NewbroAffiliationService struct {
	eligibilitySvc *NewbroEligibilityService
	roleRepo       *repository.RoleRepository
	userRepo       *repository.UserRepository
	charRepo       *repository.EveCharacterRepository
	affRepo        *repository.NewbroCaptainAffiliationRepository
}

func NewNewbroAffiliationService() *NewbroAffiliationService {
	return &NewbroAffiliationService{
		eligibilitySvc: NewNewbroEligibilityService(),
		roleRepo:       repository.NewRoleRepository(),
		userRepo:       repository.NewUserRepository(),
		charRepo:       repository.NewEveCharacterRepository(),
		affRepo:        repository.NewNewbroCaptainAffiliationRepository(),
	}
}

func (s *NewbroAffiliationService) ListCaptainCandidates(userID uint) ([]NewbroCaptainCandidate, error) {
	state, err := s.eligibilitySvc.EnsureCurrentState(userID)
	if err != nil {
		return nil, err
	}
	if !state.IsCurrentlyNewbro {
		return nil, errors.New("当前用户不符合新人资格")
	}

	captainUserIDs, err := s.roleRepo.GetRoleUserIDs(model.RoleCaptain)
	if err != nil {
		return nil, err
	}
	users, err := s.userRepo.ListByIDs(captainUserIDs)
	if err != nil {
		return nil, err
	}
	users = filterCaptainCandidateUsers(userID, users)
	if len(users) == 0 {
		return []NewbroCaptainCandidate{}, nil
	}

	primaryCharacterIDs := make([]int64, 0, len(users))
	for _, user := range users {
		if user.PrimaryCharacterID != 0 {
			primaryCharacterIDs = append(primaryCharacterIDs, user.PrimaryCharacterID)
		}
	}
	chars, err := s.charRepo.ListByCharacterIDs(primaryCharacterIDs)
	if err != nil {
		return nil, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}

	activeCounts, err := s.affRepo.CountActiveByCaptainUserIDs(captainUserIDs)
	if err != nil {
		return nil, err
	}

	result := make([]NewbroCaptainCandidate, 0, len(users))
	for _, user := range users {
		primaryChar := charByID[user.PrimaryCharacterID]
		result = append(result, NewbroCaptainCandidate{
			CaptainUserID:        user.ID,
			CaptainCharacterID:   primaryChar.CharacterID,
			CaptainCharacterName: primaryChar.CharacterName,
			CaptainNickname:      user.Nickname,
			ActiveNewbroCount:    activeCounts[user.ID],
			LastOnlineAt:         user.LastLoginAt,
		})
	}
	sortCaptainCandidatesByLastOnline(result)
	return result, nil
}

func sortCaptainCandidatesByLastOnline(result []NewbroCaptainCandidate) {
	sort.Slice(result, func(i, j int) bool {
		left := result[i].LastOnlineAt
		right := result[j].LastOnlineAt
		switch {
		case left == nil && right == nil:
			return result[i].CaptainUserID < result[j].CaptainUserID
		case left == nil:
			return false
		case right == nil:
			return true
		case left.Equal(*right):
			return result[i].CaptainUserID < result[j].CaptainUserID
		default:
			return left.After(*right)
		}
	})
}

func (s *NewbroAffiliationService) GetMyAffiliation(userID uint) (*NewbroMyAffiliationResponse, error) {
	state, err := s.eligibilitySvc.EnsureCurrentState(userID)
	if err != nil {
		return nil, err
	}

	active, err := s.affRepo.GetActiveByPlayerUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	recent, err := s.affRepo.ListRecentByPlayerUserID(userID, newbroRecentAffiliationLimit)
	if err != nil {
		return nil, err
	}

	resp := &NewbroMyAffiliationResponse{
		IsCurrentlyNewbro:  state.IsCurrentlyNewbro,
		EvaluatedAt:        state.EvaluatedAt,
		RuleVersion:        state.RuleVersion,
		DisqualifiedReason: state.DisqualifiedReason,
		RecentAffiliations: []NewbroAffiliationSummary{},
	}

	captainUserIDs := make([]uint, 0, len(recent)+1)
	if active != nil {
		captainUserIDs = append(captainUserIDs, active.CaptainUserID)
	}
	for _, item := range recent {
		captainUserIDs = append(captainUserIDs, item.CaptainUserID)
	}
	enrichment, err := s.loadCaptainPrimaryData(captainUserIDs)
	if err != nil {
		return nil, err
	}

	if active != nil {
		summary := buildNewbroAffiliationSummary(*active, enrichment)
		resp.CurrentAffiliation = &summary
	}
	for _, item := range normalizeRecentAffiliations(recent) {
		resp.RecentAffiliations = append(resp.RecentAffiliations, buildNewbroAffiliationSummary(item, enrichment))
	}
	return resp, nil
}

func (s *NewbroAffiliationService) ListMyAffiliationHistory(userID uint, page, pageSize int) ([]NewbroAffiliationSummary, int64, error) {
	normalizeLedgerPageRequest(&page, &pageSize)

	rows, total, err := s.affRepo.ListByPlayerUserIDPaged(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	captainUserIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		captainUserIDs = append(captainUserIDs, r.CaptainUserID)
	}
	enrichment, err := s.loadCaptainPrimaryData(captainUserIDs)
	if err != nil {
		return nil, 0, err
	}
	result := make([]NewbroAffiliationSummary, 0, len(rows))
	for _, r := range rows {
		result = append(result, buildNewbroAffiliationSummary(r, enrichment))
	}
	return result, total, nil
}

func (s *NewbroAffiliationService) SelectCaptain(userID, captainUserID uint) (*SelectCaptainResponse, error) {
	return s.changeCaptainAffiliation(userID, userID, captainUserID)
}

func (s *NewbroAffiliationService) EnrollPlayer(captainUserID, playerUserID uint) (*SelectCaptainResponse, error) {
	return s.changeCaptainAffiliation(captainUserID, playerUserID, captainUserID)
}

func (s *NewbroAffiliationService) changeCaptainAffiliation(actorUserID, playerUserID, captainUserID uint) (*SelectCaptainResponse, error) {
	if shouldBlockSelfAffiliation(playerUserID, captainUserID) {
		return nil, errors.New("不能选择自己作为队长或帮扶对象")
	}

	state, err := s.eligibilitySvc.EnsureCurrentState(playerUserID)
	if err != nil {
		return nil, err
	}
	if !state.IsCurrentlyNewbro {
		return nil, errors.New("目标用户当前不符合新人资格")
	}

	targetRoles, err := s.roleRepo.GetUserRoleCodes(captainUserID)
	if err != nil {
		return nil, err
	}
	if !model.ContainsRole(targetRoles, model.RoleCaptain) {
		return nil, errors.New("目标用户不是队长")
	}

	var result *SelectCaptainResponse
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		player, err := s.userRepo.GetByIDForUpdateTx(tx, playerUserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		current, err := s.affRepo.GetActiveByPlayerUserIDTx(tx, playerUserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if shouldReuseCurrentAffiliation(current, captainUserID) {
			result = &SelectCaptainResponse{
				AffiliationID: current.ID,
				CaptainUserID: current.CaptainUserID,
				StartedAt:     current.StartedAt,
			}
			return nil
		}

		now := time.Now()
		if current != nil && current.EndedAt == nil {
			if err := s.affRepo.EndActiveByPlayerUserIDTx(tx, playerUserID, now); err != nil {
				return err
			}
		}

		row := buildNewbroCaptainAffiliation(playerUserID, player.PrimaryCharacterID, captainUserID, actorUserID, now)
		if err := s.affRepo.CreateTx(tx, &row); err != nil {
			return err
		}

		result = &SelectCaptainResponse{
			AffiliationID: row.ID,
			CaptainUserID: row.CaptainUserID,
			StartedAt:     row.StartedAt,
		}
		return nil
	})
	if err != nil {
		if !isActiveAffiliationConflictError(err) {
			return nil, err
		}
		// The transaction rolled back due to a concurrent cross-process request
		// inserting an active affiliation for the same player. Read the current
		// state with a fresh connection (tx is no longer usable) and resolve.
		latestCurrent, getErr := s.affRepo.GetActiveByPlayerUserID(playerUserID)
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
		return resolveCaptainAffiliationCreateError(err, latestCurrent, captainUserID)
	}
	return result, nil
}

func (s *NewbroAffiliationService) EndAffiliation(actorUserID, playerUserID uint) error {
	current, err := s.affRepo.GetActiveByPlayerUserID(playerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("当前没有有效的帮扶关系")
		}
		return err
	}

	if actorUserID != playerUserID {
		targetRoles, err := s.roleRepo.GetUserRoleCodes(actorUserID)
		if err != nil {
			return err
		}
		if !model.ContainsRole(targetRoles, model.RoleCaptain) {
			return errors.New("只有队长可以结束他人帮扶关系")
		}
		if current.CaptainUserID != actorUserID {
			return errors.New("只能结束当前属于自己的帮扶关系")
		}
	}

	now := time.Now()
	return s.affRepo.EndActiveByPlayerUserID(playerUserID, now)
}

func (s *NewbroAffiliationService) ListCaptainEligiblePlayers(
	captainUserID uint,
	keyword string,
	page int,
	pageSize int,
) ([]CaptainEligiblePlayerListItem, int64, error) {
	normalizePageRequest(&page, &pageSize, 20, 100)

	users, total, err := s.affRepo.ListCaptainEligiblePlayers(captainUserID, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(users) == 0 {
		return []CaptainEligiblePlayerListItem{}, total, nil
	}

	playerUserIDs := make([]uint, 0, len(users))
	playerCharacterIDs := make([]int64, 0, len(users))
	for _, user := range users {
		playerUserIDs = append(playerUserIDs, user.ID)
		if user.PrimaryCharacterID != 0 {
			playerCharacterIDs = append(playerCharacterIDs, user.PrimaryCharacterID)
		}
	}

	playerChars, err := s.charRepo.ListByCharacterIDs(playerCharacterIDs)
	if err != nil {
		return nil, 0, err
	}
	playerCharByID := make(map[int64]model.EveCharacter, len(playerChars))
	for _, char := range playerChars {
		playerCharByID[char.CharacterID] = char
	}

	activeAffiliations, err := s.affRepo.ListActiveByPlayerUserIDs(playerUserIDs)
	if err != nil {
		return nil, 0, err
	}
	activeAffByPlayer := make(map[uint]model.NewbroCaptainAffiliation, len(activeAffiliations))
	currentCaptainUserIDs := make([]uint, 0, len(activeAffiliations))
	for _, row := range activeAffiliations {
		if _, exists := activeAffByPlayer[row.PlayerUserID]; exists {
			continue
		}
		activeAffByPlayer[row.PlayerUserID] = row
		currentCaptainUserIDs = append(currentCaptainUserIDs, row.CaptainUserID)
	}

	currentCaptainProfiles, err := s.loadCaptainProfiles(currentCaptainUserIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]CaptainEligiblePlayerListItem, 0, len(users))
	for _, user := range users {
		char := playerCharByID[user.PrimaryCharacterID]
		item := CaptainEligiblePlayerListItem{
			PlayerUserID:        user.ID,
			PlayerCharacterID:   char.CharacterID,
			PlayerCharacterName: char.CharacterName,
			PlayerNickname:      user.Nickname,
		}
		if currentAffiliation, ok := activeAffByPlayer[user.ID]; ok {
			profile := currentCaptainProfiles[currentAffiliation.CaptainUserID]
			item.CurrentAffiliation = &CaptainEligiblePlayerCurrentAffiliation{
				AffiliationID:        currentAffiliation.ID,
				CaptainUserID:        currentAffiliation.CaptainUserID,
				CaptainCharacterID:   profile.PrimaryCharacterID,
				CaptainCharacterName: profile.PrimaryCharacterName,
				CaptainNickname:      profile.Nickname,
				StartedAt:            currentAffiliation.StartedAt,
			}
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (s *NewbroAffiliationService) loadCaptainPrimaryData(userIDs []uint) (map[uint]model.EveCharacter, error) {
	result := make(map[uint]model.EveCharacter)
	if len(userIDs) == 0 {
		return result, nil
	}
	seen := make(map[uint]struct{}, len(userIDs))
	uniqueUserIDs := make([]uint, 0, len(userIDs))
	for _, userID := range userIDs {
		if _, ok := seen[userID]; ok || userID == 0 {
			continue
		}
		seen[userID] = struct{}{}
		uniqueUserIDs = append(uniqueUserIDs, userID)
	}
	users, err := s.userRepo.ListByIDs(uniqueUserIDs)
	if err != nil {
		return nil, err
	}
	primaryIDs := make([]int64, 0, len(users))
	userByPrimaryID := make(map[int64]uint, len(users))
	for _, user := range users {
		if user.PrimaryCharacterID == 0 {
			continue
		}
		primaryIDs = append(primaryIDs, user.PrimaryCharacterID)
		userByPrimaryID[user.PrimaryCharacterID] = user.ID
	}
	chars, err := s.charRepo.ListByCharacterIDs(primaryIDs)
	if err != nil {
		return nil, err
	}
	for _, char := range chars {
		result[userByPrimaryID[char.CharacterID]] = char
	}
	return result, nil
}

func (s *NewbroAffiliationService) loadCaptainProfiles(userIDs []uint) (map[uint]captainProfile, error) {
	return loadCaptainProfiles(s.userRepo, s.charRepo, userIDs)
}

func buildNewbroAffiliationSummary(
	row model.NewbroCaptainAffiliation,
	captainData map[uint]model.EveCharacter,
) NewbroAffiliationSummary {
	char := captainData[row.CaptainUserID]
	return NewbroAffiliationSummary{
		AffiliationID:        row.ID,
		CaptainUserID:        row.CaptainUserID,
		CaptainCharacterID:   char.CharacterID,
		CaptainCharacterName: char.CharacterName,
		StartedAt:            row.StartedAt,
		EndedAt:              row.EndedAt,
	}
}

type CaptainBountySyncService struct {
	attrRepo       *repository.CaptainBountyAttributionRepository
	charRepo       *repository.EveCharacterRepository
	userRepo       *repository.UserRepository
	affRepo        *repository.NewbroCaptainAffiliationRepository
	eligibilitySvc *NewbroEligibilityService
	runGuard       exclusiveRunGuard
}

func NewCaptainBountySyncService() *CaptainBountySyncService {
	return &CaptainBountySyncService{
		attrRepo:       repository.NewCaptainBountyAttributionRepository(),
		charRepo:       repository.NewEveCharacterRepository(),
		userRepo:       repository.NewUserRepository(),
		affRepo:        repository.NewNewbroCaptainAffiliationRepository(),
		eligibilitySvc: NewNewbroEligibilityService(),
	}
}

func (s *CaptainBountySyncService) RunSync(now time.Time) (*CaptainAttributionSyncResult, error) {
	if err := s.runGuard.Start("新人赏金归因同步"); err != nil {
		return nil, err
	}
	defer s.runGuard.Finish()

	state, err := s.attrRepo.GetSyncState(captainAttributionSyncKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if state == nil {
		state = &model.CaptainBountySyncState{SyncKey: captainAttributionSyncKey}
	}

	result := &CaptainAttributionSyncResult{}
	lookbackStart := now.AddDate(0, 0, -newbroAttributionLookbackDays)
	refTypes := supportedPlayerAttributionRefTypeList()
	userStateCache := make(map[uint]*model.NewbroPlayerState)
	captainPrimaryCache := make(map[uint]int64)

	// Reset cursor so previously-skipped journals are re-evaluated each run.
	// The LEFT JOIN in the query already excludes already-attributed journals.
	state.LastWalletJournalID = 0
	state.LastJournalAt = nil

	for {
		journals, err := s.attrRepo.ListUnattributedPlayerJournalsFromLookback(
			state.LastWalletJournalID,
			lookbackStart,
			refTypes,
			captainAttributionSyncFetchLimit,
		)
		if err != nil {
			return nil, err
		}
		if len(journals) == 0 {
			break
		}

		playerChars, err := s.charRepo.ListByCharacterIDs(extractPlayerCharacterIDs(journals))
		if err != nil {
			return nil, err
		}
		charByID := make(map[int64]model.EveCharacter, len(playerChars))
		for _, char := range playerChars {
			charByID[char.CharacterID] = char
		}

		for _, journal := range journals {
			result.ProcessedCount++
			state.LastWalletJournalID = journal.ID
			lastJournalAt := journal.Date
			state.LastJournalAt = &lastJournalAt

			if !shouldConsiderAttributionJournal(journal, now, newbroAttributionLookbackDays) {
				result.SkippedCount++
				continue
			}
			char, ok := charByID[journal.CharacterID]
			if !ok {
				result.SkippedCount++
				continue
			}

			currentState, ok := userStateCache[char.UserID]
			if !ok {
				currentState, err = s.eligibilitySvc.EnsureCurrentState(char.UserID)
				if err != nil {
					return nil, err
				}
				userStateCache[char.UserID] = currentState
			}
			if !currentState.IsCurrentlyNewbro {
				result.SkippedCount++
				continue
			}

			affiliation, err := s.affRepo.GetActiveAt(char.UserID, journal.Date)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					result.SkippedCount++
					continue
				}
				return nil, err
			}

			captainCharacterID, ok := captainPrimaryCache[affiliation.CaptainUserID]
			if !ok {
				captainUser, err := s.userRepo.GetByID(affiliation.CaptainUserID)
				if err != nil {
					return nil, err
				}
				captainCharacterID = captainUser.PrimaryCharacterID
				captainPrimaryCache[affiliation.CaptainUserID] = captainCharacterID
			}
			if captainCharacterID == 0 {
				result.SkippedCount++
				continue
			}

			candidates, err := s.attrRepo.ListCaptainCandidateJournals(
				captainCharacterID,
				journal.ContextID,
				journal.Date.Add(-newbroAttributionWindow),
				journal.Date.Add(newbroAttributionWindow),
				refTypes,
			)
			if err != nil {
				return nil, err
			}

			match := selectCaptainWalletJournalMatch(journal, candidates)
			if match == nil {
				result.SkippedCount++
				continue
			}

			exists, err := s.attrRepo.ExistsByWalletJournalID(journal.ID)
			if err != nil {
				return nil, err
			}
			if exists {
				result.SkippedCount++
				continue
			}

			row := &model.CaptainBountyAttribution{
				AffiliationID:          affiliation.ID,
				PlayerUserID:           char.UserID,
				PlayerCharacterID:      journal.CharacterID,
				CaptainUserID:          affiliation.CaptainUserID,
				CaptainCharacterID:     captainCharacterID,
				CaptainWalletJournalID: match.ID,
				WalletJournalID:        journal.ID,
				RefType:                journal.RefType,
				SystemID:               journal.ContextID,
				JournalAt:              journal.Date,
				Amount:                 journal.Amount,
			}
			if err := s.attrRepo.CreateIgnoreDuplicate(row); err != nil {
				return nil, err
			}
			result.InsertedCount++
		}
	}

	if err := s.attrRepo.SaveSyncState(state); err != nil {
		return nil, err
	}
	result.LastWalletJournalID = state.LastWalletJournalID
	return result, nil
}

func extractPlayerCharacterIDs(journals []model.EVECharacterWalletJournal) []int64 {
	seen := make(map[int64]struct{}, len(journals))
	result := make([]int64, 0, len(journals))
	for _, journal := range journals {
		if _, ok := seen[journal.CharacterID]; ok {
			continue
		}
		seen[journal.CharacterID] = struct{}{}
		result = append(result, journal.CharacterID)
	}
	return result
}
