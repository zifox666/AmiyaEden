package model

// ─── 菜单类型 ───

const (
	MenuTypeDir    = "dir"    // 目录
	MenuTypeMenu   = "menu"   // 页面
	MenuTypeButton = "button" // 按钮/权限
)

// ─── 数据模型 ───

// Menu 菜单（目录/页面/按钮）
type Menu struct {
	BaseModel
	ParentID   uint    `gorm:"default:0;index"       json:"parent_id"`
	Type       string  `gorm:"size:20;index"         json:"type"`
	Name       string  `gorm:"size:100;uniqueIndex"  json:"name"`
	Path       string  `gorm:"size:200"              json:"path"`
	Component  string  `gorm:"size:200"              json:"component"`
	Permission string  `gorm:"size:100;index"        json:"permission"`
	Title      string  `gorm:"size:200"              json:"title"`
	Icon       string  `gorm:"size:100"              json:"icon"`
	Sort       int     `gorm:"default:0"             json:"sort"`
	IsHide     bool    `gorm:"default:false"         json:"is_hide"`
	KeepAlive  bool    `gorm:"default:false"         json:"keep_alive"`
	IsHideTab  bool    `gorm:"default:false"         json:"is_hide_tab"`
	FixedTab   bool    `gorm:"default:false"         json:"fixed_tab"`
	Status     int8    `gorm:"default:1"             json:"status"`
	Children   []*Menu `gorm:"-"                    json:"children,omitempty"`
}

func (Menu) TableName() string { return "menu" }

// ─── 菜单转前端路由格式 ───

// MenuMeta 前端路由元数据
type MenuMeta struct {
	Title     string         `json:"title"`
	Icon      string         `json:"icon,omitempty"`
	KeepAlive bool           `json:"keepAlive,omitempty"`
	IsHide    bool           `json:"isHide,omitempty"`
	IsHideTab bool           `json:"isHideTab,omitempty"`
	FixedTab  bool           `json:"fixedTab,omitempty"`
	AuthList  []MenuAuthItem `json:"authList,omitempty"`
}

// MenuAuthItem 按钮权限项
type MenuAuthItem struct {
	Title    string `json:"title"`
	AuthMark string `json:"authMark"`
}

// MenuItem 前端路由菜单项
type MenuItem struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Component string      `json:"component,omitempty"`
	Meta      MenuMeta    `json:"meta"`
	Children  []*MenuItem `json:"children,omitempty"`
}

// ToMenuItem 将 Menu 转换为前端 MenuItem 格式
func (m *Menu) ToMenuItem(buttons []*Menu) *MenuItem {
	item := &MenuItem{
		Path:      m.Path,
		Name:      m.Name,
		Component: m.Component,
		Meta: MenuMeta{
			Title:     m.Title,
			Icon:      m.Icon,
			KeepAlive: m.KeepAlive,
			IsHide:    m.IsHide,
			IsHideTab: m.IsHideTab,
			FixedTab:  m.FixedTab,
		},
	}

	// 将按钮类型子菜单转为 authList
	if len(buttons) > 0 {
		for _, btn := range buttons {
			item.Meta.AuthList = append(item.Meta.AuthList, MenuAuthItem{
				Title:    btn.Title,
				AuthMark: btn.Permission,
			})
		}
	}

	return item
}

// ─── 菜单种子数据 ───

// MenuSeed 用于种子数据的菜单定义
type MenuSeed struct {
	ParentName string // 父菜单 Name（空表示根菜单）
	Menu       Menu
}

// GetSystemMenuSeeds 返回系统默认菜单种子数据
func GetSystemMenuSeeds() []MenuSeed {
	return []MenuSeed{
		// ── Dashboard ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "Dashboard", Path: "/dashboard", Component: "/index/index", Title: "menus.dashboard.title", Icon: "ri:pie-chart-line", Sort: 100, Status: 1}},
		{ParentName: "Dashboard", Menu: Menu{Type: MenuTypeMenu, Name: "Console", Path: "console", Component: "/dashboard/console", Title: "menus.dashboard.console", Sort: 100, FixedTab: true, Status: 1}},
		{ParentName: "Dashboard", Menu: Menu{Type: MenuTypeMenu, Name: "Characters", Path: "characters", Component: "/dashboard/characters", Title: "menus.characters.title", Sort: 90, KeepAlive: true, Status: 1}},

		// ── Operation ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "Operation", Path: "/operation", Component: "/index/index", Title: "menus.operation.title", Icon: "ri:ship-line", Sort: 90, Status: 1}},
		{ParentName: "Operation", Menu: Menu{Type: MenuTypeMenu, Name: "Fleets", Path: "fleets", Component: "/operation/fleets", Title: "menus.operation.fleets", Sort: 100, KeepAlive: true, Status: 1}},
		{ParentName: "Operation", Menu: Menu{Type: MenuTypeMenu, Name: "FleetDetail", Path: "fleet-detail/:id", Component: "/operation/fleet-detail", Title: "menus.operation.fleetDetail", Sort: 90, IsHide: true, Status: 1}},
		{ParentName: "Operation", Menu: Menu{Type: MenuTypeMenu, Name: "MyPap", Path: "pap", Component: "/operation/pap", Title: "menus.operation.pap", Sort: 80, KeepAlive: true, Status: 1}},
		{ParentName: "Operation", Menu: Menu{Type: MenuTypeMenu, Name: "JoinFleet", Path: "join", Component: "/operation/join", Title: "menus.operation.join", Sort: 60, IsHide: true, Status: 1}},

		// ── Shop ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "ShopRoot", Path: "/shop", Component: "/index/index", Title: "menus.shop.title", Icon: "ri:shopping-bag-line", Sort: 85, Status: 1}},
		{ParentName: "ShopRoot", Menu: Menu{Type: MenuTypeMenu, Name: "Shop", Path: "browse", Component: "/shop/browse", Title: "menus.shop.browse", Sort: 100, KeepAlive: true, Status: 1}},
		{ParentName: "ShopRoot", Menu: Menu{Type: MenuTypeMenu, Name: "ShopManage", Path: "manage", Component: "/shop/manage", Title: "menus.shop.manage", Sort: 90, KeepAlive: true, Status: 1}},
		{ParentName: "ShopRoot", Menu: Menu{Type: MenuTypeMenu, Name: "Wallet", Path: "wallet", Component: "/shop/wallet", Title: "menus.shop.wallet", Sort: 70, KeepAlive: true, Status: 1}},
		{ParentName: "ShopManage", Menu: Menu{Type: MenuTypeButton, Name: "ShopProductAdd", Permission: "system:shop:product:add", Title: "新增商品", Sort: 100, Status: 1}},
		{ParentName: "ShopManage", Menu: Menu{Type: MenuTypeButton, Name: "ShopProductEdit", Permission: "system:shop:product:edit", Title: "编辑商品", Sort: 90, Status: 1}},
		{ParentName: "ShopManage", Menu: Menu{Type: MenuTypeButton, Name: "ShopProductDelete", Permission: "system:shop:product:delete", Title: "删除商品", Sort: 80, Status: 1}},
		{ParentName: "ShopManage", Menu: Menu{Type: MenuTypeButton, Name: "ShopOrderReview", Permission: "system:shop:order:review", Title: "审批订单", Sort: 70, Status: 1}},

		// ── SRP ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "SRP", Path: "/srp", Component: "/index/index", Title: "menus.srp.title", Icon: "ri:money-dollar-box-line", Sort: 80, Status: 1}},
		{ParentName: "SRP", Menu: Menu{Type: MenuTypeMenu, Name: "SrpApply", Path: "srp-apply", Component: "/srp/apply", Title: "menus.srp.srpApply", Sort: 100, KeepAlive: true, Status: 1}},
		{ParentName: "SRP", Menu: Menu{Type: MenuTypeMenu, Name: "SrpManage", Path: "srp-manage", Component: "/srp/manage", Title: "menus.srp.srpManage", Sort: 90, KeepAlive: true, Status: 1}},
		{ParentName: "SrpManage", Menu: Menu{Type: MenuTypeButton, Name: "SrpManageReview", Permission: "srp:manage:review", Title: "审批", Sort: 100, Status: 1}},
		{ParentName: "SRP", Menu: Menu{Type: MenuTypeMenu, Name: "SrpPrices", Path: "srp-prices", Component: "/srp/prices", Title: "menus.srp.srpPrices", Sort: 80, KeepAlive: true, Status: 1}},
		{ParentName: "SrpPrices", Menu: Menu{Type: MenuTypeButton, Name: "SrpPriceAdd", Permission: "srp:price:add", Title: "新增价格", Sort: 100, Status: 1}},
		{ParentName: "SrpPrices", Menu: Menu{Type: MenuTypeButton, Name: "SrpPriceDelete", Permission: "srp:price:delete", Title: "删除价格", Sort: 90, Status: 1}},

		// ── System ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "System", Path: "/system", Component: "/index/index", Title: "menus.system.title", Icon: "ri:user-3-line", Sort: 70, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "User", Path: "user", Component: "/system/user", Title: "menus.system.user", Sort: 100, KeepAlive: true, Status: 1}},
		{ParentName: "User", Menu: Menu{Type: MenuTypeButton, Name: "UserDelete", Permission: "system:user:delete", Title: "删除用户", Sort: 100, Status: 1}},
		{ParentName: "User", Menu: Menu{Type: MenuTypeButton, Name: "UserSetRole", Permission: "system:user:role", Title: "分配角色", Sort: 90, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "RoleManage", Path: "role", Component: "/system/role", Title: "menus.system.role", Sort: 90, KeepAlive: true, Status: 1}},
		{ParentName: "RoleManage", Menu: Menu{Type: MenuTypeButton, Name: "RoleAdd", Permission: "system:role:add", Title: "新增角色", Sort: 100, Status: 1}},
		{ParentName: "RoleManage", Menu: Menu{Type: MenuTypeButton, Name: "RoleEdit", Permission: "system:role:edit", Title: "编辑角色", Sort: 90, Status: 1}},
		{ParentName: "RoleManage", Menu: Menu{Type: MenuTypeButton, Name: "RoleDelete", Permission: "system:role:delete", Title: "删除角色", Sort: 80, Status: 1}},
		{ParentName: "RoleManage", Menu: Menu{Type: MenuTypeButton, Name: "RolePermission", Permission: "system:role:permission", Title: "权限设置", Sort: 70, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "Menus", Path: "menu", Component: "/system/menu", Title: "menus.system.menu", Sort: 80, KeepAlive: true, Status: 1}},
		{ParentName: "Menus", Menu: Menu{Type: MenuTypeButton, Name: "MenuAdd", Permission: "system:menu:add", Title: "新增菜单", Sort: 100, Status: 1}},
		{ParentName: "Menus", Menu: Menu{Type: MenuTypeButton, Name: "MenuEdit", Permission: "system:menu:edit", Title: "编辑菜单", Sort: 90, Status: 1}},
		{ParentName: "Menus", Menu: Menu{Type: MenuTypeButton, Name: "MenuDelete", Permission: "system:menu:delete", Title: "删除菜单", Sort: 80, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "ESIRefresh", Path: "esi-refresh", Component: "/system/esi-refresh", Title: "menus.system.esiRefresh", Sort: 70, KeepAlive: true, Status: 1}},
		{ParentName: "ESIRefresh", Menu: Menu{Type: MenuTypeButton, Name: "ESIRun", Permission: "system:esi:run", Title: "执行任务", Sort: 100, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "SystemWallet", Path: "wallet", Component: "/system/wallet", Title: "menus.system.wallet", Sort: 65, KeepAlive: true, Status: 1}},
		{ParentName: "SystemWallet", Menu: Menu{Type: MenuTypeButton, Name: "WalletAdjust", Permission: "system:wallet:adjust", Title: "调整余额", Sort: 100, Status: 1}},
		{ParentName: "SystemWallet", Menu: Menu{Type: MenuTypeButton, Name: "WalletViewLog", Permission: "system:wallet:log", Title: "查看日志", Sort: 90, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "AlliancePAP", Path: "pap", Component: "/system/pap", Title: "menus.system.alliancePap", Sort: 63, KeepAlive: true, Status: 1}},
		{ParentName: "AlliancePAP", Menu: Menu{Type: MenuTypeButton, Name: "AlliancePAPFetch", Permission: "system:pap:fetch", Title: "手动拉取", Sort: 100, Status: 1}},
		{ParentName: "System", Menu: Menu{Type: MenuTypeMenu, Name: "UserCenter", Path: "user-center", Component: "/system/user-center", Title: "menus.system.userCenter", Sort: 60, IsHide: true, KeepAlive: true, IsHideTab: true, Status: 1}},

		// ── Result ──
		{ParentName: "", Menu: Menu{Type: MenuTypeDir, Name: "Result", Path: "/result", Component: "/index/index", Title: "menus.result.title", Icon: "ri:checkbox-circle-line", Sort: 10, IsHide: true, Status: 1}},
		{ParentName: "Result", Menu: Menu{Type: MenuTypeMenu, Name: "ResultSuccess", Path: "success", Component: "/result/success", Title: "menus.result.success", Sort: 100, IsHide: true, Status: 1}},
		{ParentName: "Result", Menu: Menu{Type: MenuTypeMenu, Name: "ResultFail", Path: "fail", Component: "/result/fail", Title: "menus.result.fail", Sort: 90, IsHide: true, Status: 1}},
	}
}

// DefaultRoleMenuMap 默认角色-菜单映射（角色Code -> 菜单Name列表）
// super_admin 不在这里定义，代码中直接赋予所有权限
func DefaultRoleMenuMap() map[string][]string {
	return map[string][]string{
		RoleAdmin: {
			"Dashboard", "Console", "Characters",
			"Operation", "Fleets", "FleetDetail", "MyPap", "Wallet", "JoinFleet",
			"ShopRoot", "Shop", "ShopManage", "ShopProductAdd", "ShopProductEdit", "ShopProductDelete", "ShopOrderReview",
			"SRP", "SrpApply", "SrpManage", "SrpManageReview", "SrpPrices", "SrpPriceAdd", "SrpPriceDelete",
			"System", "User", "UserDelete", "UserSetRole",
			"RoleManage", "RoleAdd", "RoleEdit", "RoleDelete", "RolePermission",
			"Menus", "MenuAdd", "MenuEdit", "MenuDelete",
			"ESIRefresh", "ESIRun",
			"SystemWallet", "WalletAdjust", "WalletViewLog",
			"AlliancePAP", "AlliancePAPFetch",
			"UserCenter",
			"Result", "ResultSuccess", "ResultFail",
		},
		RoleFC: {
			"Dashboard", "Console", "Characters",
			"Operation", "Fleets", "FleetDetail", "MyPap", "Wallet", "JoinFleet",
			"ShopRoot", "Shop",
			"SRP", "SrpApply", "SrpManage", "SrpManageReview",
			"Result", "ResultSuccess", "ResultFail",
		},
		RoleSRP: {
			"Dashboard", "Console", "Characters",
			"Operation", "MyPap", "Wallet", "JoinFleet",
			"ShopRoot", "Shop",
			"SRP", "SrpApply", "SrpManage", "SrpManageReview", "SrpPrices", "SrpPriceAdd", "SrpPriceDelete",
			"Result", "ResultSuccess", "ResultFail",
		},
		RoleUser: {
			"Dashboard", "Console", "Characters",
			"Operation", "MyPap", "Wallet", "JoinFleet",
			"ShopRoot", "Shop",
			"SRP", "SrpApply",
			"Result", "ResultSuccess", "ResultFail",
			"UserCenter",
		},
		RoleGuest: {
			"Dashboard", "Console", "Characters",
			"Result", "ResultSuccess", "ResultFail",
		},
	}
}
