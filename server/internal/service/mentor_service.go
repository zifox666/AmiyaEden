package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"
)

func shouldBlockSelfMentorApplication(menteeUserID, mentorUserID uint) bool {
	return menteeUserID != 0 && menteeUserID == mentorUserID
}

func filterMentorCandidateUsers(currentUserID uint, users []model.User) []model.User {
	if len(users) == 0 {
		return users
	}
	filtered := make([]model.User, 0, len(users))
	for _, user := range users {
		if user.ID == currentUserID {
			continue
		}
		filtered = append(filtered, user)
	}
	return filtered
}

func calculateMentorDaysActive(createdAt time.Time, lastLoginAt *time.Time) int {
	if lastLoginAt == nil {
		return 0
	}
	duration := lastLoginAt.Sub(createdAt)
	if duration < 0 {
		return 0
	}
	return int(duration.Hours() / 24)
}

type MentorCandidate struct {
	MentorUserID        uint       `json:"mentor_user_id"`
	MentorCharacterID   int64      `json:"mentor_character_id"`
	MentorCharacterName string     `json:"mentor_character_name"`
	MentorNickname      string     `json:"mentor_nickname"`
	MentorQQ            string     `json:"qq"`
	MentorDiscordID     string     `json:"discord_id"`
	ActiveMenteeCount   int64      `json:"active_mentee_count"`
	LastOnlineAt        *time.Time `json:"last_online_at"`
}

func buildMentorCandidate(user model.User, primaryChar model.EveCharacter, activeMenteeCount int64) MentorCandidate {
	return MentorCandidate{
		MentorUserID:        user.ID,
		MentorCharacterID:   primaryChar.CharacterID,
		MentorCharacterName: primaryChar.CharacterName,
		MentorNickname:      user.Nickname,
		MentorQQ:            user.QQ,
		MentorDiscordID:     user.DiscordID,
		ActiveMenteeCount:   activeMenteeCount,
		LastOnlineAt:        user.LastLoginAt,
	}
}

type MenteeMyStatusResponse struct {
	IsEligible          bool                    `json:"is_eligible"`
	DisqualifiedReason  string                  `json:"disqualified_reason"`
	CurrentRelationship *MentorRelationshipView `json:"current_relationship"`
}

type MentorRelationshipView struct {
	ID                  uint       `json:"id"`
	MenteeUserID        uint       `json:"mentee_user_id"`
	MentorUserID        uint       `json:"mentor_user_id"`
	Status              string     `json:"status"`
	AppliedAt           time.Time  `json:"applied_at"`
	RespondedAt         *time.Time `json:"responded_at"`
	RevokedAt           *time.Time `json:"revoked_at"`
	GraduatedAt         *time.Time `json:"graduated_at"`
	MentorCharacterID   int64      `json:"mentor_character_id"`
	MentorCharacterName string     `json:"mentor_character_name"`
	MentorNickname      string     `json:"mentor_nickname"`
	MentorQQ            string     `json:"mentor_qq"`
	MentorDiscordID     string     `json:"mentor_discord_id"`
	MenteeCharacterID   int64      `json:"mentee_character_id"`
	MenteeCharacterName string     `json:"mentee_character_name"`
	MenteeNickname      string     `json:"mentee_nickname"`
}

func buildMentorRelationshipView(
	rel model.MentorMenteeRelationship,
	mentorUser model.User,
	menteeUser model.User,
	mentorChar model.EveCharacter,
	menteeChar model.EveCharacter,
) MentorRelationshipView {
	return MentorRelationshipView{
		ID:                  rel.ID,
		MenteeUserID:        rel.MenteeUserID,
		MentorUserID:        rel.MentorUserID,
		Status:              rel.Status,
		AppliedAt:           rel.AppliedAt,
		RespondedAt:         rel.RespondedAt,
		RevokedAt:           rel.RevokedAt,
		GraduatedAt:         rel.GraduatedAt,
		MentorCharacterID:   mentorChar.CharacterID,
		MentorCharacterName: mentorChar.CharacterName,
		MentorNickname:      mentorUser.Nickname,
		MentorQQ:            mentorUser.QQ,
		MentorDiscordID:     mentorUser.DiscordID,
		MenteeCharacterID:   menteeChar.CharacterID,
		MenteeCharacterName: menteeChar.CharacterName,
		MenteeNickname:      menteeUser.Nickname,
	}
}

type MentorMenteeListItem struct {
	RelationshipID          uint       `json:"relationship_id"`
	MenteeUserID            uint       `json:"mentee_user_id"`
	MenteeCharacterID       int64      `json:"mentee_character_id"`
	MenteeCharacterName     string     `json:"mentee_character_name"`
	MenteeNickname          string     `json:"mentee_nickname"`
	MenteeQQ                string     `json:"mentee_qq"`
	MenteeDiscordID         string     `json:"mentee_discord_id"`
	MenteeTotalSP           int64      `json:"mentee_total_sp"`
	MenteeTotalPap          float64    `json:"mentee_total_pap"`
	MenteeDaysActive        int        `json:"mentee_days_active"`
	Status                  string     `json:"status"`
	AppliedAt               time.Time  `json:"applied_at"`
	RespondedAt             *time.Time `json:"responded_at"`
	GraduatedAt             *time.Time `json:"graduated_at"`
	DistributedStages       []int      `json:"distributed_stages"`
	DistributedRewardAmount float64    `json:"distributed_reward_amount"`
}

type MentorService struct {
	eligibilitySvc *MentorEligibilityService
	relRepo        *repository.MentorRelationshipRepository
	roleRepo       *repository.RoleRepository
	userRepo       *repository.UserRepository
	charRepo       *repository.EveCharacterRepository
	skillRepo      *repository.EveSkillRepository
	fleetRepo      *repository.FleetRepository
	distRepo       *repository.MentorRewardDistributionRepository
	now            func() time.Time
}

func NewMentorService() *MentorService {
	return &MentorService{
		eligibilitySvc: NewMentorEligibilityService(),
		relRepo:        repository.NewMentorRelationshipRepository(),
		roleRepo:       repository.NewRoleRepository(),
		userRepo:       repository.NewUserRepository(),
		charRepo:       repository.NewEveCharacterRepository(),
		skillRepo:      repository.NewEveSkillRepository(),
		fleetRepo:      repository.NewFleetRepository(),
		distRepo:       repository.NewMentorRewardDistributionRepository(),
		now:            time.Now,
	}
}

func (s *MentorService) ListMentorCandidates(menteeUserID uint) ([]MentorCandidate, error) {
	eligibility, err := s.eligibilitySvc.EvaluateEligibility(menteeUserID)
	if err != nil {
		return nil, err
	}
	if !eligibility.IsEligible {
		return nil, errors.New("当前用户不符合学员资格")
	}

	mentorUserIDs, err := s.roleRepo.GetRoleUserIDs(model.RoleMentor)
	if err != nil {
		return nil, err
	}
	users, err := s.userRepo.ListByIDs(mentorUserIDs)
	if err != nil {
		return nil, err
	}
	users = filterMentorCandidateUsers(menteeUserID, users)
	if len(users) == 0 {
		return []MentorCandidate{}, nil
	}

	sort.SliceStable(users, func(i, j int) bool {
		a, b := users[i].LastLoginAt, users[j].LastLoginAt
		switch {
		case a == nil && b == nil:
			return users[i].ID < users[j].ID
		case a == nil:
			return false
		case b == nil:
			return true
		case !a.Equal(*b):
			return a.After(*b)
		default:
			return users[i].ID < users[j].ID
		}
	})

	userIDs := make([]uint, 0, len(users))
	characterIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
		if user.PrimaryCharacterID != 0 {
			characterIDs = append(characterIDs, user.PrimaryCharacterID)
		}
	}

	chars, err := s.charRepo.ListByCharacterIDs(characterIDs)
	if err != nil {
		return nil, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}

	activeCounts, err := s.relRepo.CountActiveByMentorUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]MentorCandidate, 0, len(users))
	for _, user := range users {
		primaryChar := charByID[user.PrimaryCharacterID]
		result = append(result, buildMentorCandidate(user, primaryChar, activeCounts[user.ID]))
	}
	return result, nil
}

func (s *MentorService) GetMyMenteeStatus(userID uint) (*MenteeMyStatusResponse, error) {
	eligibility, err := s.eligibilitySvc.EvaluateEligibility(userID)
	if err != nil {
		return nil, err
	}

	resp := &MenteeMyStatusResponse{
		IsEligible:         eligibility.IsEligible,
		DisqualifiedReason: eligibility.DisqualifiedReason,
	}

	current, err := s.relRepo.GetActiveOrPendingByMenteeUserID(userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		current = nil
	}
	if current != nil {
		view, err := s.enrichRelationshipView(current)
		if err != nil {
			return nil, err
		}
		resp.CurrentRelationship = view
	}

	return resp, nil
}

func (s *MentorService) ApplyForMentor(menteeUserID, mentorUserID uint) (*model.MentorMenteeRelationship, error) {
	if shouldBlockSelfMentorApplication(menteeUserID, mentorUserID) {
		return nil, errors.New("不能选择自己作为导师")
	}

	eligibility, err := s.eligibilitySvc.EvaluateEligibility(menteeUserID)
	if err != nil {
		return nil, err
	}
	if !eligibility.IsEligible {
		return nil, errors.New("当前用户不符合学员资格")
	}

	roles, err := s.roleRepo.GetUserRoleCodes(mentorUserID)
	if err != nil {
		return nil, err
	}
	if !model.ContainsRole(roles, model.RoleMentor) {
		return nil, errors.New("目标用户不是导师")
	}

	current, err := s.relRepo.GetActiveOrPendingByMenteeUserID(menteeUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil && current != nil {
		return nil, errors.New("已有待处理或已建立的导师关系")
	}

	mentee, err := s.userRepo.GetByID(menteeUserID)
	if err != nil {
		return nil, err
	}

	row := &model.MentorMenteeRelationship{
		MenteeUserID:                    menteeUserID,
		MenteePrimaryCharacterIDAtStart: mentee.PrimaryCharacterID,
		MentorUserID:                    mentorUserID,
		Status:                          model.MentorRelationStatusPending,
		AppliedAt:                       s.now(),
	}
	if err := s.relRepo.Create(row); err != nil {
		if repository.IsActiveMentorRelationConflictError(err) {
			return nil, errors.New("已有待处理或已建立的导师关系")
		}
		return nil, err
	}
	return row, nil
}

func (s *MentorService) AcceptApplication(mentorUserID, relationshipID uint) error {
	rel, err := s.relRepo.GetByID(relationshipID)
	if err != nil {
		return err
	}
	if rel.MentorUserID != mentorUserID {
		return errors.New("无权操作此申请")
	}
	if rel.Status != model.MentorRelationStatusPending {
		return errors.New("该申请不在待处理状态")
	}
	return s.relRepo.UpdateStatus(relationshipID, model.MentorRelationStatusActive, map[string]any{"responded_at": s.now()})
}

func (s *MentorService) RejectApplication(mentorUserID, relationshipID uint) error {
	rel, err := s.relRepo.GetByID(relationshipID)
	if err != nil {
		return err
	}
	if rel.MentorUserID != mentorUserID {
		return errors.New("无权操作此申请")
	}
	if rel.Status != model.MentorRelationStatusPending {
		return errors.New("该申请不在待处理状态")
	}
	return s.relRepo.UpdateStatus(relationshipID, model.MentorRelationStatusRejected, map[string]any{"responded_at": s.now()})
}

func normalizeMentorStatuses(status string) []string {
	switch status {
	case "active":
		return []string{model.MentorRelationStatusActive}
	case "pending":
		return []string{model.MentorRelationStatusPending}
	case "rejected":
		return []string{model.MentorRelationStatusRejected}
	case "revoked":
		return []string{model.MentorRelationStatusRevoked}
	case "graduated":
		return []string{model.MentorRelationStatusGraduated}
	case "all", "":
		return nil
	default:
		return []string{model.MentorRelationStatusActive}
	}
}

func (s *MentorService) ListMyMentees(mentorUserID uint, status string, page, pageSize int) ([]MentorMenteeListItem, int64, error) {
	normalizePageRequest(&page, &pageSize, 20, 100)
	rows, total, err := s.relRepo.ListByMentorUserID(mentorUserID, normalizeMentorStatuses(status), page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.enrichMenteeListItems(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *MentorService) ListPendingApplications(mentorUserID uint) ([]MentorMenteeListItem, error) {
	rows, err := s.relRepo.ListPendingByMentorUserID(mentorUserID)
	if err != nil {
		return nil, err
	}
	return s.enrichMenteeListItems(rows)
}

func (s *MentorService) AdminRevokeRelationship(adminUserID, relationshipID uint) error {
	rel, err := s.relRepo.GetByID(relationshipID)
	if err != nil {
		return err
	}
	if rel.Status != model.MentorRelationStatusActive && rel.Status != model.MentorRelationStatusPending {
		return errors.New("只能撤销待处理或进行中的导师关系")
	}
	return s.relRepo.UpdateStatus(relationshipID, model.MentorRelationStatusRevoked, map[string]any{
		"revoked_at": s.now(),
		"revoked_by": adminUserID,
	})
}

func (s *MentorService) AdminListAllRelationships(status, keyword string, page, pageSize int) ([]MentorRelationshipView, int64, error) {
	normalizePageRequest(&page, &pageSize, 20, 200)
	rows, total, err := s.relRepo.ListAllPaged(repository.MentorRelationshipAdminFilter{Status: status, Keyword: keyword}, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	views := make([]MentorRelationshipView, 0, len(rows))
	for i := range rows {
		view, err := s.enrichRelationshipView(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		views = append(views, *view)
	}
	return views, total, nil
}

func (s *MentorService) enrichRelationshipView(rel *model.MentorMenteeRelationship) (*MentorRelationshipView, error) {
	users, err := s.userRepo.ListByIDs([]uint{rel.MentorUserID, rel.MenteeUserID})
	if err != nil {
		return nil, err
	}
	userByID := make(map[uint]model.User, len(users))
	charIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userByID[user.ID] = user
		if user.PrimaryCharacterID != 0 {
			charIDs = append(charIDs, user.PrimaryCharacterID)
		}
	}
	chars, err := s.charRepo.ListByCharacterIDs(charIDs)
	if err != nil {
		return nil, err
	}
	charByID := make(map[int64]model.EveCharacter, len(chars))
	for _, char := range chars {
		charByID[char.CharacterID] = char
	}

	mentorUser := userByID[rel.MentorUserID]
	menteeUser := userByID[rel.MenteeUserID]
	mentorChar := charByID[mentorUser.PrimaryCharacterID]
	menteeChar := charByID[menteeUser.PrimaryCharacterID]

	view := buildMentorRelationshipView(*rel, mentorUser, menteeUser, mentorChar, menteeChar)
	return &view, nil
}

func (s *MentorService) enrichMenteeListItems(rows []model.MentorMenteeRelationship) ([]MentorMenteeListItem, error) {
	if len(rows) == 0 {
		return []MentorMenteeListItem{}, nil
	}

	menteeUserIDs := make([]uint, 0, len(rows))
	relationshipIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		menteeUserIDs = append(menteeUserIDs, row.MenteeUserID)
		relationshipIDs = append(relationshipIDs, row.ID)
	}

	users, err := s.userRepo.ListByIDs(menteeUserIDs)
	if err != nil {
		return nil, err
	}
	userByID := make(map[uint]model.User, len(users))
	primaryCharIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userByID[user.ID] = user
		if user.PrimaryCharacterID != 0 {
			primaryCharIDs = append(primaryCharIDs, user.PrimaryCharacterID)
		}
	}

	primaryChars, err := s.charRepo.ListByCharacterIDs(primaryCharIDs)
	if err != nil {
		return nil, err
	}
	charByID := make(map[int64]model.EveCharacter, len(primaryChars))
	for _, char := range primaryChars {
		charByID[char.CharacterID] = char
	}

	allChars, err := s.charRepo.ListByUserIDs(menteeUserIDs)
	if err != nil {
		return nil, err
	}
	characterIDs := make([]int64, 0, len(allChars))
	for _, char := range allChars {
		characterIDs = append(characterIDs, char.CharacterID)
	}
	skillTotals, err := s.skillRepo.GetSkillTotalsByCharacterIDs(characterIDs)
	if err != nil {
		return nil, err
	}
	totalSPByUserID := make(map[uint]int64, len(menteeUserIDs))
	for _, char := range allChars {
		totalSPByUserID[char.UserID] += skillTotals[char.CharacterID]
	}

	papTotals, err := s.fleetRepo.SumPapTotalsByUserIDs(menteeUserIDs)
	if err != nil {
		return nil, err
	}

	distributedStages, err := s.distRepo.ListDistributedStageOrdersByRelationshipIDs(relationshipIDs)
	if err != nil {
		return nil, err
	}
	distributedRewardAmounts, err := s.distRepo.SumRewardAmountsByRelationshipIDs(relationshipIDs)
	if err != nil {
		return nil, err
	}

	items := make([]MentorMenteeListItem, 0, len(rows))
	for _, row := range rows {
		user := userByID[row.MenteeUserID]
		primaryChar := charByID[user.PrimaryCharacterID]
		stages := distributedStages[row.ID]
		if stages == nil {
			stages = []int{}
		}
		items = append(items, MentorMenteeListItem{
			RelationshipID:          row.ID,
			MenteeUserID:            row.MenteeUserID,
			MenteeCharacterID:       primaryChar.CharacterID,
			MenteeCharacterName:     primaryChar.CharacterName,
			MenteeNickname:          user.Nickname,
			MenteeQQ:                user.QQ,
			MenteeDiscordID:         user.DiscordID,
			MenteeTotalSP:           totalSPByUserID[row.MenteeUserID],
			MenteeTotalPap:          papTotals[row.MenteeUserID],
			MenteeDaysActive:        calculateMentorDaysActive(user.CreatedAt, user.LastLoginAt),
			Status:                  row.Status,
			AppliedAt:               row.AppliedAt,
			RespondedAt:             row.RespondedAt,
			GraduatedAt:             row.GraduatedAt,
			DistributedStages:       stages,
			DistributedRewardAmount: distributedRewardAmounts[row.ID],
		})
	}
	return items, nil
}
