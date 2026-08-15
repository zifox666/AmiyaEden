package bootstrap

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// InitDB 初始化 GORM PostgreSQL 数据库连接
func InitDB() {
	cfg := global.Config.Database
	dataDBName := cfg.DataDBName
	if dataDBName == "" {
		dataDBName = cfg.DBName
	}
	if dataDBName == "" || cfg.SDEDBName == "" {
		global.Logger.Fatal("数据库配置不完整", zap.String("required", "database.data_dbname,database.sde_dbname"))
	}

	// GORM 日志级别
	gormLogLevel := logger.Warn
	if global.Config.Server.Mode == "debug" {
		gormLogLevel = logger.Info
	}

	openDB := func(name string) *gorm.DB {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, name, cfg.SSLMode,
		)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(gormLogLevel),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 禁用表名复数
			},
			// 禁止通过事务进行外键处理
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			global.Logger.Fatal("数据库连接失败", zap.String("db", name), zap.Error(err))
		}

		sqlDB, err := db.DB()
		if err != nil {
			global.Logger.Fatal("获取底层数据库连接失败", zap.String("db", name), zap.Error(err))
		}

		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return db
	}

	global.DB = openDB(dataDBName)
	global.SDEDB = openDB(cfg.SDEDBName)
	global.Logger.Info("数据库连接成功", zap.String("host", cfg.Host), zap.String("data_db", dataDBName), zap.String("sde_db", cfg.SDEDBName))

	autoMigrateData(global.DB)
	autoMigrateSDE(global.SDEDB)
}

// autoMigrateData migrates business data only. Legacy tables are retained for
// compatibility even though their routes and jobs are not loaded in AT mode.
func autoMigrateData(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.User{},
		&model.OperationLog{},
		&model.EveCharacter{},
		// ESI 数据表
		&model.EveCharacterAsset{},
		&model.EveCharacterNotification{},
		&model.EveCharacterTitle{},
		&model.EveCharacterCloneBaseInfo{},
		&model.EveCharacterImplants{},
		&model.EveStructure{},
		&model.CorpStructureInfo{},
		&model.EveStation{},

		&model.EveKillmailList{},
		&model.EveKillmailItem{},
		&model.EveCharacterKillmail{},
		&model.EveKillmailAttacker{},

		&model.EveCharacterContract{},
		&model.EveCharacterContractItem{},
		&model.EveCharacterContractBid{},

		&model.EVECharacterWallet{},
		&model.EVECharacterWalletJournal{},
		&model.EVECharacterWalletTransaction{},

		&model.EveCharacterSkill{},
		&model.EveCharacterSkills{},
		&model.EveCharacterSkillQueue{},

		&model.EveCharacterFitting{},
		&model.EveCharacterFittingItem{},
		// Fleet / Operation 相关表
		&model.Fleet{},
		&model.FleetMember{},
		&model.FleetPapLog{},
		&model.FleetInvite{},
		&model.FleetBattleIncentive{},
		&model.SystemWallet{},
		&model.WalletTransaction{},
		&model.WalletLog{},
		// 商店相关表
		&model.ShopProduct{},
		&model.ShopOrder{},
		&model.ShopRedeemCode{},
		// 抽奖相关表
		&model.ShopLotteryActivity{},
		&model.ShopLotteryPrize{},
		&model.ShopLotteryRecord{},
		// SRP 补损相关表
		&model.SrpShipPrice{},
		&model.SrpApplication{},
		// 舰队配置相关表
		&model.FleetConfig{},
		&model.FleetConfigFitting{},
		&model.FleetConfigFittingItem{},
		&model.FleetConfigFittingItemReplacement{},
		// 联盟 PAP 相关表
		&model.AlliancePAPRecord{},
		&model.AlliancePAPSummary{},
		// 技能规划相关表
		&model.SkillPlan{},
		&model.SkillPlanItem{},
		// 系统配置表
		&model.SystemConfig{},
		// Mumble 语音中心
		&model.MumbleAccount{},
		&model.VoiceRoleGroupMapping{},
		// RBAC 权限相关表
		&model.Role{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.UserRole{},
		// ESI / SeAT 自动权限映射表
		&model.SeatRoleMapping{},
		&model.EsiRoleMapping{},
		&model.EsiTitleMapping{},
		&model.EveCharacterCorpRole{},
		&model.AutoRoleLog{},
		// 准入名单表
		&model.AllowedEntity{},
		// SeAT 用户绑定表
		&model.SeatUser{},
		// AT recruitment
		&model.RecruitmentQuestionnaire{},
		&model.RecruitmentField{},
		&model.RecruitmentLink{},
		&model.RecruitmentApplication{},
		&model.RecruitmentAnswer{},
		&model.RecruitmentSkillSnapshot{},
		&model.RecruitmentReviewEvent{},
		// AT score rules
		&model.ATScoreRuleVersion{},
		&model.ATScoreShip{},
		&model.ATScoreBannedShip{},
		&model.ATScoreDuplicatePenalty{},
		// AT fits and composition sets
		&model.ATFit{},
		&model.ATSet{},
		&model.ATSetSlot{},
		&model.ATFitReplacement{},
		// AT seasons, matches and contribution ledger
		&model.ATSeason{},
		&model.ATMatchType{},
		&model.ATCleanupShip{},
		&model.ATMatch{},
		&model.ATMatchRegistration{},
		&model.ATMatchRound{},
		&model.ATMatchRoundRoster{},
		&model.ATMatchStateEvent{},
		&model.ATContributionLedger{},
		// AT tournament-server killmails and battle reports
		&model.ATKillmail{},
		&model.ATKillmailAttacker{},
		&model.ATKillmailAssignmentEvent{},
		&model.ATBRVersion{},
		&model.ATBRTeamSummary{},
		&model.ATBRParticipantSnapshot{},
	); err != nil {
		global.Logger.Fatal("DATA 数据库迁移失败", zap.Error(err))
	}
	// An early AT draft accidentally made field_key globally unique. Replace
	// only that obsolete index; no application or questionnaire rows are removed.
	if db.Migrator().HasIndex(&model.RecruitmentField{}, "idx_recruitment_field_key") {
		if err := db.Migrator().DropIndex(&model.RecruitmentField{}, "idx_recruitment_field_key"); err != nil {
			global.Logger.Fatal("修复报名字段复合索引失败", zap.Error(err))
		}
	}

	// 种子数据：系统角色 → 系统菜单 → 默认角色权限 → 迁移已有用户
	roleSvc := service.NewRoleService()
	roleSvc.SeedSystemRoles()
	roleSvc.SeedSystemMenus()
	roleSvc.MigrateExistingUsers()
	seedATMatchTypes(db)
}

func autoMigrateSDE(db *gorm.DB) {
	if err := db.AutoMigrate(&model.SdeVersion{}); err != nil {
		global.Logger.Fatal("SDE 数据库迁移失败", zap.Error(err))
	}
}

func seedATMatchTypes(db *gorm.DB) {
	defaults := []model.ATMatchType{
		{Name: "training", ContributionValue: 1, Active: true},
		{Name: "official", ContributionValue: 2, Active: true},
		{Name: "other", ContributionValue: 1, Active: true},
	}
	for _, item := range defaults {
		if err := db.Where("name = ?", item.Name).FirstOrCreate(&item).Error; err != nil {
			global.Logger.Warn("初始化 AT 比赛类型失败", zap.String("name", item.Name), zap.Error(err))
		}
	}
}
