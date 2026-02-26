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

// MapRegion 映射 SDE 中的 mapRegions 表
type MapRegion struct {
	RegionID   int     `gorm:"column:regionID;primaryKey" json:"region_id"`
	RegionName string  `gorm:"column:regionName"          json:"region_name"`
	X          float64 `gorm:"column:x"                   json:"x"`
	Y          float64 `gorm:"column:y"                   json:"y"`
	Z          float64 `gorm:"column:z"                   json:"z"`
	XMin       float64 `gorm:"column:xMin"                json:"x_min"`
	XMax       float64 `gorm:"column:xMax"                json:"x_max"`
	YMin       float64 `gorm:"column:yMin"                json:"y_min"`
	YMax       float64 `gorm:"column:yMax"                json:"y_max"`
	ZMin       float64 `gorm:"column:zMin"                json:"z_min"`
	ZMax       float64 `gorm:"column:zMax"                json:"z_max"`
	FactionID  int     `gorm:"column:factionID"           json:"faction_id"`
	Nebula     int     `gorm:"column:nebula"              json:"nebula"`
	Radius     float32 `gorm:"column:radius"              json:"radius"`
}

func (MapRegion) TableName() string { return "mapRegions" }

// MapConstellation 映射 SDE 中的 mapConstellations 表
type MapConstellation struct {
	RegionID          int     `gorm:"column:regionID"                   json:"region_id"`
	ConstellationID   int     `gorm:"column:constellationID;primaryKey" json:"constellation_id"`
	ConstellationName string  `gorm:"column:constellationName"          json:"constellation_name"`
	X                 float64 `gorm:"column:x"                          json:"x"`
	Y                 float64 `gorm:"column:y"                          json:"y"`
	Z                 float64 `gorm:"column:z"                          json:"z"`
	XMin              float64 `gorm:"column:xMin"                       json:"x_min"`
	XMax              float64 `gorm:"column:xMax"                       json:"x_max"`
	YMin              float64 `gorm:"column:yMin"                       json:"y_min"`
	YMax              float64 `gorm:"column:yMax"                       json:"y_max"`
	ZMin              float64 `gorm:"column:zMin"                       json:"z_min"`
	ZMax              float64 `gorm:"column:zMax"                       json:"z_max"`
	FactionID         int     `gorm:"column:factionID"                  json:"faction_id"`
	Radius            float32 `gorm:"column:radius"                     json:"radius"`
}

func (MapConstellation) TableName() string { return "mapConstellations" }

// MapSolarSystem 映射 SDE 中的 mapSolarSystems 表
type MapSolarSystem struct {
	RegionID        int     `gorm:"column:regionID"                      json:"region_id"`
	ConstellationID int     `gorm:"column:constellationID;index"         json:"constellation_id"`
	SolarSystemID   int     `gorm:"column:solarSystemID;primaryKey"      json:"solar_system_id"`
	SolarSystemName string  `gorm:"column:solarSystemName"               json:"solar_system_name"`
	X               float64 `gorm:"column:x"                             json:"x"`
	Y               float64 `gorm:"column:y"                             json:"y"`
	Z               float64 `gorm:"column:z"                             json:"z"`
	XMin            float64 `gorm:"column:xMin"                          json:"x_min"`
	XMax            float64 `gorm:"column:xMax"                          json:"x_max"`
	YMin            float64 `gorm:"column:yMin"                          json:"y_min"`
	YMax            float64 `gorm:"column:yMax"                          json:"y_max"`
	ZMin            float64 `gorm:"column:zMin"                          json:"z_min"`
	ZMax            float64 `gorm:"column:zMax"                          json:"z_max"`
	Luminosity      float64 `gorm:"column:luminosity"                    json:"luminosity"`
	Border          int8    `gorm:"column:border"                        json:"border"`
	Fringe          int8    `gorm:"column:fringe"                        json:"fringe"`
	Corridor        int8    `gorm:"column:corridor"                      json:"corridor"`
	Hub             int8    `gorm:"column:hub"                           json:"hub"`
	International   int8    `gorm:"column:international"                 json:"international"`
	Regional        int8    `gorm:"column:regional"                      json:"regional"`
	Constellation   int8    `gorm:"column:constellation"                 json:"constellation"`
	Security        float64 `gorm:"column:security;index"                json:"security"`
	FactionID       int     `gorm:"column:factionID"                     json:"faction_id"`
	Radius          float64 `gorm:"column:radius"                        json:"radius"`
	SunTypeID       int     `gorm:"column:sunTypeID"                     json:"sun_type_id"`
	SecurityClass   string  `gorm:"column:securityClass;type:varchar(2)" json:"security_class"`
}

func (MapSolarSystem) TableName() string { return "mapSolarSystems" }
