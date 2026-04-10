package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/gorm"
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

func TestDeleteUser(t *testing.T) {
	t.Run("admin cannot delete user with registered qq", func(t *testing.T) {
		db := newUserDeletionTestDB(t)
		seedUserForDeletion(t, db, model.User{
			BaseModel: model.BaseModel{ID: 1},
			Nickname:  "Amiya",
			QQ:        "12345",
			Role:      model.RoleUser,
		}, []string{model.RoleUser})

		originalDB := global.DB
		global.DB = db
		defer func() { global.DB = originalDB }()

		err := NewUserService().DeleteUser(1, []string{model.RoleAdmin})
		if err == nil || !strings.Contains(err.Error(), "超级管理员") {
			t.Fatalf("expected super admin deletion restriction, got %v", err)
		}
	})

	t.Run("super admin can delete user with registered discord", func(t *testing.T) {
		db := newUserDeletionTestDB(t)
		seedUserForDeletion(t, db, model.User{
			BaseModel: model.BaseModel{ID: 1},
			Nickname:  "Amiya",
			DiscordID: "998877",
			Role:      model.RoleUser,
		}, []string{model.RoleUser})

		originalDB := global.DB
		global.DB = db
		defer func() { global.DB = originalDB }()

		if err := NewUserService().DeleteUser(1, []string{model.RoleSuperAdmin}); err != nil {
			t.Fatalf("expected super admin delete to pass, got %v", err)
		}
	})
}

func TestUpdateUserByAdmin(t *testing.T) {
	t.Run("admin cannot update contacts from system user management", func(t *testing.T) {
		db := newUserDeletionTestDB(t)
		seedUserForDeletion(t, db, model.User{
			BaseModel: model.BaseModel{ID: 1},
			Nickname:  "Amiya",
			QQ:        "12345",
			Role:      model.RoleUser,
		}, []string{model.RoleUser})

		originalDB := global.DB
		global.DB = db
		defer func() { global.DB = originalDB }()

		qq := "54321"
		err := NewUserService().UpdateUserByAdmin(1, []string{model.RoleAdmin}, UserPatch{QQ: &qq})
		if err == nil || !strings.Contains(err.Error(), "QQ") {
			t.Fatalf("expected admin qq edit to be rejected, got %v", err)
		}
	})

	t.Run("super admin can update nickname and contacts of non super admin user", func(t *testing.T) {
		db := newUserDeletionTestDB(t)
		seedUserForDeletion(t, db, model.User{
			BaseModel: model.BaseModel{ID: 1},
			Nickname:  "Amiya",
			QQ:        "12345",
			DiscordID: "old-discord",
			Role:      model.RoleUser,
		}, []string{model.RoleUser})

		originalDB := global.DB
		global.DB = db
		defer func() { global.DB = originalDB }()

		nickname := "Doctor"
		qq := "54321"
		discordID := "doctor-1001"
		err := NewUserService().UpdateUserByAdmin(1, []string{model.RoleSuperAdmin}, UserPatch{
			Nickname:  &nickname,
			QQ:        &qq,
			DiscordID: &discordID,
		})
		if err != nil {
			t.Fatalf("expected super admin contact edit to pass, got %v", err)
		}

		updated, err := NewUserService().GetUserByID(1)
		if err != nil {
			t.Fatalf("reload updated user: %v", err)
		}
		if updated.Nickname != nickname {
			t.Fatalf("expected nickname %q, got %q", nickname, updated.Nickname)
		}
		if updated.QQ != qq {
			t.Fatalf("expected qq %q, got %q", qq, updated.QQ)
		}
		if updated.DiscordID != discordID {
			t.Fatalf("expected discord id %q, got %q", discordID, updated.DiscordID)
		}
	})

	t.Run("super admin cannot update contacts of another super admin", func(t *testing.T) {
		db := newUserDeletionTestDB(t)
		seedUserForDeletion(t, db, model.User{
			BaseModel: model.BaseModel{ID: 1},
			Nickname:  "Amiya",
			QQ:        "12345",
			Role:      model.RoleSuperAdmin,
		}, []string{model.RoleSuperAdmin})

		originalDB := global.DB
		global.DB = db
		defer func() { global.DB = originalDB }()

		qq := "54321"
		err := NewUserService().UpdateUserByAdmin(1, []string{model.RoleSuperAdmin}, UserPatch{QQ: &qq})
		if err == nil || !strings.Contains(err.Error(), "超级管理员") {
			t.Fatalf("expected super admin contact edit on protected target to be rejected, got %v", err)
		}
	})
}

func newUserDeletionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := newUserServiceTestDB(t)
	if err := db.AutoMigrate(&model.UserRole{}); err != nil {
		t.Fatalf("auto migrate user_role: %v", err)
	}
	return db
}

func seedUserForDeletion(t *testing.T, db *gorm.DB, user model.User, roles []string) {
	t.Helper()

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	for _, roleCode := range roles {
		if err := db.Create(&model.UserRole{UserID: user.ID, RoleCode: roleCode}).Error; err != nil {
			t.Fatalf("create user role: %v", err)
		}
	}
}
