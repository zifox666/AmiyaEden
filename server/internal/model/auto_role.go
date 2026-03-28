package model

import "time"

// ─────────────────────────────────────────────
//  ESI 自动权限映射
// ─────────────────────────────────────────────

// EsiRoleMapping ESI 军团角色 → 系统权限映射
// 一个 ESI role 可以映射到多个系统角色，一个系统角色也可被多个 ESI role 映射
type EsiRoleMapping struct {
	BaseModel
	EsiRole  string `gorm:"size:100;not null;index:idx_esi_role_mapping,unique" json:"esi_role"`  // ESI 军团角色名（如 Director、Accountant）
	RoleID   uint   `gorm:"not null;index:idx_esi_role_mapping,unique"          json:"role_id"`   // 系统角色 ID
	RoleCode string `gorm:"-"                                                   json:"role_code"` // 系统角色编码（仅展示用，不入库）
	RoleName string `gorm:"-"                                                   json:"role_name"` // 系统角色名称（仅展示用，不入库）
}

func (EsiRoleMapping) TableName() string { return "esi_role_mapping" }

// EsiTitleMapping ESI 军团头衔 → 系统权限映射
type EsiTitleMapping struct {
	BaseModel
	CorporationID int64  `gorm:"not null;index:idx_esi_title_mapping,unique"        json:"corporation_id"` // 军团 ID
	TitleID       int    `gorm:"not null;index:idx_esi_title_mapping,unique"         json:"title_id"`      // ESI 头衔 ID
	TitleName     string `gorm:"size:256"                                            json:"title_name"`    // 头衔名称（展示用）
	RoleID        uint   `gorm:"not null;index:idx_esi_title_mapping,unique"         json:"role_id"`       // 系统角色 ID
	RoleCode      string `gorm:"-"                                                   json:"role_code"`     // 系统角色编码（仅展示用）
	RoleName      string `gorm:"-"                                                   json:"role_name"`     // 系统角色名称（仅展示用）
}

func (EsiTitleMapping) TableName() string { return "esi_title_mapping" }

// EveCharacterCorpRole 角色的 ESI 军团角色快照（去重后的完整列表）
type EveCharacterCorpRole struct {
	ID          uint   `gorm:"primarykey"                                                    json:"id"`
	CharacterID int64  `gorm:"not null;index:idx_char_corp_role,unique"                       json:"character_id"`
	CorpRole    string `gorm:"size:100;not null;index:idx_char_corp_role,unique"               json:"corp_role"` // ESI 军团角色名
}

func (EveCharacterCorpRole) TableName() string { return "eve_character_corp_role" }

// AutoRoleLog 自动权限同步操作日志
type AutoRoleLog struct {
	ID       uint      `gorm:"primarykey"                    json:"id"`
	UserID   uint      `gorm:"not null;index"                json:"user_id"`
	Username string    `gorm:"size:128;default:''"           json:"username"`  // 冗余用户名，方便展示
	RoleID   uint      `gorm:"not null"                      json:"role_id"`
	RoleName string    `gorm:"size:128;default:''"           json:"role_name"` // 冗余角色名，方便展示
	RoleCode string    `gorm:"size:64;default:''"            json:"role_code"`
	Action   string    `gorm:"size:16;not null"              json:"action"`    // "add" | "remove"
	Reason   string    `gorm:"size:32;not null;default:''"   json:"reason"`    // "esi_role" | "title" | "director"
	CreatedAt time.Time `gorm:"autoCreateTime;index"          json:"created_at"`
}

func (AutoRoleLog) TableName() string { return "auto_role_log" }

// ─── ESI 军团角色名常量 ───

var AllEsiCorpRoles = []string{
	"Account_Take_1", "Account_Take_2", "Account_Take_3", "Account_Take_4",
	"Account_Take_5", "Account_Take_6", "Account_Take_7",
	"Accountant", "Auditor", "Brand_Manager", "Communications_Officer",
	"Config_Equipment", "Config_Starbase_Equipment",
	"Container_Take_1", "Container_Take_2", "Container_Take_3", "Container_Take_4",
	"Container_Take_5", "Container_Take_6", "Container_Take_7",
	"Contract_Manager", "Deliveries_Container_Take", "Deliveries_Query", "Deliveries_Take",
	"Diplomat", "Director", "Factory_Manager", "Fitting_Manager",
	"Hangar_Query_1", "Hangar_Query_2", "Hangar_Query_3", "Hangar_Query_4",
	"Hangar_Query_5", "Hangar_Query_6", "Hangar_Query_7",
	"Hangar_Take_1", "Hangar_Take_2", "Hangar_Take_3", "Hangar_Take_4",
	"Hangar_Take_5", "Hangar_Take_6", "Hangar_Take_7",
	"Junior_Accountant", "Personnel_Manager", "Project_Manager",
	"Rent_Factory_Facility", "Rent_Office", "Rent_Research_Facility",
	"Security_Officer", "Skill_Plan_Manager", "Starbase_Defense_Operator",
	"Starbase_Fuel_Technician", "Station_Manager", "Trader",
}
