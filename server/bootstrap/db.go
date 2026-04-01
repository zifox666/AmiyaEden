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

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// GORM 日志级别
	gormLogLevel := logger.Warn
	if global.Config.Server.Mode == "debug" {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 禁用表名复数
		},
		// 禁止通过事务进行外键处理
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		global.Logger.Fatal("数据库连接失败", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Logger.Fatal("获取底层数据库连接失败", zap.Error(err))
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	global.Logger.Info("数据库连接成功", zap.String("host", cfg.Host), zap.String("db", cfg.DBName))

	// 自动迁移
	autoMigrate(db)
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.User{},
		&model.OperationLog{},
		&model.EveCharacter{},
		&model.SdeVersion{},
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
		&model.SystemWallet{},
		&model.WalletTransaction{},
		&model.WalletLog{},
		// 商店相关表
		&model.ShopProduct{},
		&model.ShopOrder{},
		&model.ShopRedeemCode{},
		// SRP 补损相关表
		&model.SrpShipPrice{},
		&model.SrpApplication{},
		// 舰队配置相关表
		&model.FleetConfig{},
		&model.FleetConfigFitting{},
		&model.FleetConfigFittingItem{},
		&model.FleetConfigFittingItemReplacement{},
		// 军团技能计划相关表
		&model.SkillPlan{},
		&model.SkillPlanSkill{},
		&model.SkillPlanCheckCharacter{},
		&model.SkillPlanCheckPlan{},
		// 军团福利相关表
		&model.Welfare{},
		&model.WelfareSkillPlan{},
		&model.WelfareApplication{},
		// 新人帮扶相关表
		&model.NewbroPlayerState{},
		&model.NewbroCaptainAffiliation{},
		&model.CaptainBountyAttribution{},
		&model.CaptainBountySyncState{},
		&model.CaptainRewardSettlement{},
		&model.MentorMenteeRelationship{},
		&model.MentorRewardStage{},
		&model.MentorRewardDistribution{},
		// 联盟 PAP 相关表
		&model.AlliancePAPRecord{},
		&model.AlliancePAPSummary{},
		&model.PAPTypeRate{},
		// 系统配置表
		&model.SystemConfig{},
		// RBAC 权限相关表
		&model.UserRole{},
		// ESI 自动权限映射表
		&model.EsiRoleMapping{},
		&model.EsiTitleMapping{},
		&model.EveCharacterCorpRole{},
	); err != nil {
		global.Logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 清理旧列/旧表（GORM AutoMigrate 不会自动删除）
	dropObsoleteSchema(db)
	ensureCustomIndexes(db)

	// 数据迁移：user_role / esi 映射表从 role_id 迁移到 role_code，然后删除 role 表
	roleSvc := service.NewRoleService()
	roleSvc.MigrateUserRoleTableToCode()
	roleSvc.MigrateEsiMappingsToCode()

	// 删除旧的 role 表（迁移完成后不再需要）
	if db.Migrator().HasTable("role") {
		if err := db.Migrator().DropTable("role"); err != nil {
			global.Logger.Warn("删除旧 role 表失败", zap.Error(err))
		} else {
			global.Logger.Info("已删除旧 role 表")
		}
	}

	// 迁移旧 User.Role 字段到 user_role 表
	roleSvc.MigrateExistingUsers()
}

// dropObsoleteSchema 删除历史遗留的已被移除的列和表
func dropObsoleteSchema(db *gorm.DB) {
	migrator := db.Migrator()
	type colDrop struct {
		table string
		col   string
	}
	drops := []colDrop{
		{"fleet_config_fitting", "eft"},
		{"fleet_config_fitting", "ship_name"},
	}
	for _, d := range drops {
		if migrator.HasColumn(d.table, d.col) {
			if err := migrator.DropColumn(d.table, d.col); err != nil {
				global.Logger.Warn("删除旧列失败", zap.String("table", d.table), zap.String("col", d.col), zap.Error(err))
			} else {
				global.Logger.Info("已删除旧列", zap.String("table", d.table), zap.String("col", d.col))
			}
		}
	}

	obsoleteTables := []string{
		"shop_lottery_record",
		"shop_lottery_prize",
		"shop_lottery_activity",
	}
	for _, table := range obsoleteTables {
		if migrator.HasTable(table) {
			if err := migrator.DropTable(table); err != nil {
				global.Logger.Warn("删除旧表失败", zap.String("table", table), zap.Error(err))
			} else {
				global.Logger.Info("已删除旧表", zap.String("table", table))
			}
		}
	}
}

func newbroCustomIndexStatements() []string {
	return []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_newbro_captain_affiliation_active_player_user_id ON newbro_captain_affiliation (player_user_id) WHERE ended_at IS NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_mentor_rel_active_mentee ON mentor_mentee_relationship (mentee_user_id) WHERE status IN ('pending', 'active') AND deleted_at IS NULL`,
	}
}

func ensureCustomIndexes(db *gorm.DB) {
	for _, stmt := range newbroCustomIndexStatements() {
		if err := db.Exec(stmt).Error; err != nil {
			global.Logger.Warn("创建自定义索引失败", zap.String("statement", stmt), zap.Error(err))
		}
	}
}
