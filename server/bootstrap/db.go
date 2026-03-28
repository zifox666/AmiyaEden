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
		// RBAC 权限相关表
		&model.Role{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.UserRole{},
		// ESI 自动权限映射表
		&model.EsiRoleMapping{},
		&model.EsiTitleMapping{},
		&model.EveCharacterCorpRole{},
		&model.AutoRoleLog{},
	); err != nil {
		global.Logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 清理旧列（GORM AutoMigrate 不会自动删除列）
	dropObsoleteColumns(db)

	// 种子数据：系统角色 → 系统菜单 → 默认角色权限 → 迁移已有用户
	roleSvc := service.NewRoleService()
	roleSvc.SeedSystemRoles()
	roleSvc.SeedSystemMenus()
	roleSvc.MigrateExistingUsers()
}

// dropObsoleteColumns 删除历史遗留的已被移除的列
func dropObsoleteColumns(db *gorm.DB) {
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
}
