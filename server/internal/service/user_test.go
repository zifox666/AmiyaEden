package service

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"
)

func TestValidatePrimaryCharacterTokenHealth(t *testing.T) {
	t.Run("rejects invalid primary character token", func(t *testing.T) {
		user := model.User{PrimaryCharacterID: 9001}
		characters := []model.EveCharacter{
			{CharacterID: 9001, TokenInvalid: true},
			{CharacterID: 9002, TokenInvalid: false},
		}

		err := validatePrimaryCharacterTokenHealth(user, characters)
		if err == nil || !strings.Contains(err.Error(), "ESI") {
			t.Fatalf("expected invalid primary token error, got %v", err)
		}
	})

	t.Run("allows invalid non primary character token", func(t *testing.T) {
		user := model.User{PrimaryCharacterID: 9001}
		characters := []model.EveCharacter{
			{CharacterID: 9001, TokenInvalid: false},
			{CharacterID: 9002, TokenInvalid: true},
		}

		if err := validatePrimaryCharacterTokenHealth(user, characters); err != nil {
			t.Fatalf("expected invalid non-primary token to pass, got %v", err)
		}
	})
}

func TestValidateImpersonationTargetPrimaryCharacterHealth(t *testing.T) {
	user := model.User{PrimaryCharacterID: 9001}
	characters := []model.EveCharacter{{CharacterID: 9001, TokenInvalid: true}}

	err := validateImpersonationTargetPrimaryCharacterHealth(user, characters)
	if err == nil || !strings.Contains(err.Error(), "ESI") {
		t.Fatalf("expected impersonation health error, got %v", err)
	}
}

func TestBuildUserPatchUpdates(t *testing.T) {
	t.Run("current profile requires nickname and contact", func(t *testing.T) {
		current := &model.User{}
		nickname := "Amiya"

		_, _, err := buildUserPatchUpdates(current, UserPatch{Nickname: &nickname}, true)
		if err == nil || !strings.Contains(err.Error(), "QQ") {
			t.Fatalf("expected contact validation error, got %v", err)
		}
	})

	t.Run("nickname length is limited by runes", func(t *testing.T) {
		current := &model.User{QQ: "12345"}
		nickname := "一二三四五六七八九十十一十二十三十四十五十六十七十八十九二十贰"

		_, _, err := buildUserPatchUpdates(current, UserPatch{Nickname: &nickname}, true)
		if err == nil || !strings.Contains(err.Error(), "昵称最多 20 个字符") {
			t.Fatalf("expected nickname length error, got %v", err)
		}
	})

	t.Run("admin can update status without forcing completion", func(t *testing.T) {
		current := &model.User{}
		status := int8(0)

		_, updates, err := buildUserPatchUpdates(current, UserPatch{Status: &status}, false)
		if err != nil {
			t.Fatalf("expected admin status update to pass, got %v", err)
		}
		if got := updates["status"]; got != status {
			t.Fatalf("expected status %d, got %+v", status, got)
		}
	})

	t.Run("current profile can switch contact methods", func(t *testing.T) {
		current := &model.User{Nickname: "Amiya", QQ: "12345"}
		qq := ""
		discordID := "998877"

		_, updates, err := buildUserPatchUpdates(current, UserPatch{
			QQ:        &qq,
			DiscordID: &discordID,
		}, true)
		if err != nil {
			t.Fatalf("expected contact switch to pass, got %v", err)
		}
		if got := updates["qq"]; got != qq {
			t.Fatalf("expected qq to clear, got %+v", got)
		}
		if got := updates["discord_id"]; got != discordID {
			t.Fatalf("expected discord_id %q, got %+v", discordID, got)
		}
	})

	t.Run("qq must be digits only", func(t *testing.T) {
		current := &model.User{Nickname: "Amiya", DiscordID: "998877"}
		qq := "12ab34"

		_, _, err := buildUserPatchUpdates(current, UserPatch{QQ: &qq}, true)
		if err == nil || !strings.Contains(err.Error(), "只能包含数字") {
			t.Fatalf("expected qq digits-only error, got %v", err)
		}
	})

	t.Run("qq must be unique across users", func(t *testing.T) {
		err := validateContactOwner(7, &model.User{BaseModel: model.BaseModel{ID: 9}}, "QQ 号码")
		if err == nil || !strings.Contains(err.Error(), "QQ 号码") {
			t.Fatalf("expected QQ uniqueness error, got %v", err)
		}
	})

	t.Run("same user can keep existing discord id", func(t *testing.T) {
		err := validateContactOwner(7, &model.User{BaseModel: model.BaseModel{ID: 7}}, "Discord ID")
		if err != nil {
			t.Fatalf("expected same user to keep contact, got %v", err)
		}
	})
}

func TestValidateManageUserPermission(t *testing.T) {
	t.Run("super admin can manage admin users", func(t *testing.T) {
		err := validateManageUserPermission(
			[]string{model.RoleSuperAdmin},
			[]string{model.RoleAdmin},
		)
		if err != nil {
			t.Fatalf("expected super admin manage to pass, got %v", err)
		}
	})

	t.Run("admin cannot edit super admin", func(t *testing.T) {
		err := validateManageUserPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleSuperAdmin},
		)
		if err == nil || !strings.Contains(err.Error(), "不能编辑或删除") {
			t.Fatalf("expected protected edit error, got %v", err)
		}
	})

	t.Run("admin cannot edit another admin", func(t *testing.T) {
		err := validateManageUserPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin, model.RoleUser},
		)
		if err == nil || !strings.Contains(err.Error(), "不能编辑或删除") {
			t.Fatalf("expected peer admin edit to be blocked, got %v", err)
		}
	})

	t.Run("admin can manage normal user", func(t *testing.T) {
		err := validateManageUserPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
		)
		if err != nil {
			t.Fatalf("expected normal user manage to pass, got %v", err)
		}
	})
}
