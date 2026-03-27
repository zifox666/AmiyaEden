package service

import (
	"amiya-eden/internal/model"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const newbroRecentAffiliationLimit = 10
const newbroActiveAffiliationUniqueIndex = "idx_newbro_captain_affiliation_active_player_user_id"

func normalizeRecentAffiliations(rows []model.NewbroCaptainAffiliation) []model.NewbroCaptainAffiliation {
	if len(rows) <= newbroRecentAffiliationLimit {
		return rows
	}
	return rows[:newbroRecentAffiliationLimit]
}

func shouldReuseCurrentAffiliation(current *model.NewbroCaptainAffiliation, captainUserID uint) bool {
	if current == nil || current.EndedAt != nil {
		return false
	}
	return current.CaptainUserID == captainUserID
}

func shouldBlockSelfAffiliation(playerUserID, captainUserID uint) bool {
	return playerUserID != 0 && playerUserID == captainUserID
}

func filterCaptainCandidateUsers(currentUserID uint, users []model.User) []model.User {
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

func buildNewbroCaptainAffiliation(
	playerUserID uint,
	playerPrimaryCharacterIDAtStart int64,
	captainUserID uint,
	createdBy uint,
	startedAt time.Time,
) model.NewbroCaptainAffiliation {
	return model.NewbroCaptainAffiliation{
		PlayerUserID:                    playerUserID,
		PlayerPrimaryCharacterIDAtStart: playerPrimaryCharacterIDAtStart,
		CaptainUserID:                   captainUserID,
		CreatedBy:                       createdBy,
		StartedAt:                       startedAt,
	}
}

func isActiveAffiliationConflictError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") && strings.Contains(msg, "newbro_captain_affiliation")
}

func resolveCaptainAffiliationCreateError(
	err error,
	current *model.NewbroCaptainAffiliation,
	captainUserID uint,
) (*SelectCaptainResponse, error) {
	if !isActiveAffiliationConflictError(err) {
		return nil, err
	}
	if shouldReuseCurrentAffiliation(current, captainUserID) {
		return &SelectCaptainResponse{
			AffiliationID: current.ID,
			CaptainUserID: current.CaptainUserID,
			StartedAt:     current.StartedAt,
		}, nil
	}
	return nil, errors.New("目标用户的帮扶关系刚刚被其他请求更新，请重试")
}
