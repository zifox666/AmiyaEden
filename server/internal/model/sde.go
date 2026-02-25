package model

// SdeVersion 记录当前已导入的 SDE 版本信息
type SdeVersion struct {
	BaseModel
	Version   string `gorm:"type:varchar(100);not null;uniqueIndex" json:"version"`
	BuildHash string `gorm:"type:varchar(200)"                      json:"build_hash"`
	Note      string `gorm:"type:varchar(500)"                      json:"note"`
}

func (SdeVersion) TableName() string { return "sde_versions" }

// TrnTranslation 映射 SDE 中的 trnTranslations 表
// tcID=8 → invTypes, tcID=7 → invGroups, tcID=6 → invCategories
type TrnTranslation struct {
	TcID       int    `gorm:"column:tcID;primaryKey"       json:"tc_id"`
	KeyID      int    `gorm:"column:keyID;primaryKey"       json:"key_id"`
	LanguageID string `gorm:"column:languageID;primaryKey"  json:"language_id"`
	Text       string `gorm:"column:text"                   json:"text"`
}

func (TrnTranslation) TableName() string { return "trnTranslations" }

// InvType 映射 SDE 中的 invTypes 表
type InvType struct {
	TypeID        int     `gorm:"column:typeID;primaryKey"   json:"type_id"`
	GroupID       int     `gorm:"column:groupID"             json:"group_id"`
	TypeName      string  `gorm:"column:typeName"            json:"type_name"`
	Description   string  `gorm:"column:description"         json:"description"`
	Mass          float64 `gorm:"column:mass"                json:"mass"`
	Volume        float64 `gorm:"column:volume"              json:"volume"`
	Capacity      float64 `gorm:"column:capacity"            json:"capacity"`
	PortionSize   int     `gorm:"column:portionSize"         json:"portion_size"`
	RaceID        int     `gorm:"column:raceID"              json:"race_id"`
	Published     int8    `gorm:"column:published"           json:"published"`
	MarketGroupID int     `gorm:"column:marketGroupID"       json:"market_group_id"`
	IconID        int     `gorm:"column:iconID"              json:"icon_id"`
	GraphicID     int     `gorm:"column:graphicID"           json:"graphic_id"`
}

func (InvType) TableName() string { return "invTypes" }

// InvGroup 映射 SDE 中的 invGroups 表
type InvGroup struct {
	GroupID    int    `gorm:"column:groupID;primaryKey"  json:"group_id"`
	CategoryID int    `gorm:"column:categoryID"          json:"category_id"`
	GroupName  string `gorm:"column:groupName"           json:"group_name"`
	Published  int8   `gorm:"column:published"           json:"published"`
}

func (InvGroup) TableName() string { return "invGroups" }

// InvCategory 映射 SDE 中的 invCategories 表
type InvCategory struct {
	CategoryID   int    `gorm:"column:categoryID;primaryKey" json:"category_id"`
	CategoryName string `gorm:"column:categoryName"          json:"category_name"`
	Published    int8   `gorm:"column:published"             json:"published"`
}

func (InvCategory) TableName() string { return "invCategories" }
