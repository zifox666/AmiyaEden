package model

// MenuMeta 路由元数据
type MenuMeta struct {
	Title     string `json:"title"`
	Icon      string `json:"icon,omitempty"`
	KeepAlive bool   `json:"keepAlive,omitempty"`
	IsHide    bool   `json:"isHide,omitempty"`
	IsHideTab bool   `json:"isHideTab,omitempty"`
	FixedTab  bool   `json:"fixedTab,omitempty"`
}

// MenuItem 路由菜单项（与前端 AppRouteRecord 格式一致）
type MenuItem struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Component string      `json:"component,omitempty"`
	Meta      MenuMeta    `json:"meta"`
	Children  []*MenuItem `json:"children,omitempty"`
}

// menuItemWithRoles 内部用于权限过滤的路由定义
type menuItemWithRoles struct {
	MenuItem
	// 访问此菜单所需的最低角色（空表示任意已登录用户均可）
	requiredRole string
	children     []*menuItemWithRoles
}

// allMenus 全量路由定义，与前端 router/modules 保持一致
// requiredRole 填写最低访问角色（使用 model.Role* 常量），空串表示所有已登录用户可见
var allMenus = []*menuItemWithRoles{
	{
		MenuItem: MenuItem{
			Path:      "/dashboard",
			Name:      "Dashboard",
			Component: "/index/index",
			Meta: MenuMeta{
				Title: "menus.dashboard.title",
				Icon:  "ri:pie-chart-line",
			},
		},
		requiredRole: "", // 所有已登录用户
		children: []*menuItemWithRoles{
			{
				MenuItem: MenuItem{
					Path:      "console",
					Name:      "Console",
					Component: "/dashboard/console",
					Meta: MenuMeta{
						Title:    "menus.dashboard.console",
						FixedTab: true,
					},
				},
				requiredRole: "",
			},
			{
				MenuItem: MenuItem{
					Path:      "characters",
					Name:      "Characters",
					Component: "/dashboard/characters",
					Meta: MenuMeta{
						Title:     "menus.characters.title",
						KeepAlive: true,
					},
				},
				requiredRole: "",
			},
		},
	},
	{
		MenuItem: MenuItem{
			Path:      "/system",
			Name:      "System",
			Component: "/index/index",
			Meta: MenuMeta{
				Title: "menus.system.title",
				Icon:  "ri:user-3-line",
			},
		},
		requiredRole: RoleAdmin, // Admin 及以上
		children: []*menuItemWithRoles{
			{
				MenuItem: MenuItem{
					Path:      "user",
					Name:      "User",
					Component: "/system/user",
					Meta: MenuMeta{
						Title:     "menus.system.user",
						KeepAlive: true,
					},
				},
				requiredRole: RoleAdmin,
			},
			{
				MenuItem: MenuItem{
					Path:      "role",
					Name:      "Role",
					Component: "/system/role",
					Meta: MenuMeta{
						Title:     "menus.system.role",
						KeepAlive: true,
					},
				},
				requiredRole: RoleSuperAdmin,
			},
			{
				MenuItem: MenuItem{
					Path:      "esi-refresh",
					Name:      "ESIRefresh",
					Component: "/system/esi-refresh",
					Meta: MenuMeta{
						Title:     "menus.system.esiRefresh",
						KeepAlive: true,
					},
				},
				requiredRole: RoleAdmin,
			},
			{
				MenuItem: MenuItem{
					Path:      "user-center",
					Name:      "UserCenter",
					Component: "/system/user-center",
					Meta: MenuMeta{
						Title:     "menus.system.userCenter",
						IsHide:    true,
						KeepAlive: true,
						IsHideTab: true,
					},
				},
				requiredRole: "", // 所有已登录用户可访问（个人中心）
			},
			{
				MenuItem: MenuItem{
					Path:      "menu",
					Name:      "Menus",
					Component: "/system/menu",
					Meta: MenuMeta{
						Title:     "menus.system.menu",
						KeepAlive: true,
					},
				},
				requiredRole: RoleSuperAdmin,
			},
		},
	},
	{
		MenuItem: MenuItem{
			Path:      "/operation",
			Name:      "Operation",
			Component: "/index/index",
			Meta: MenuMeta{
				Title: "menus.operation.title",
				Icon:  "ri:ship-line",
			},
		},
		requiredRole: "", // 所有已登录用户可查看
		children: []*menuItemWithRoles{
			{
				MenuItem: MenuItem{
					Path:      "fleets",
					Name:      "Fleets",
					Component: "/operation/fleets",
					Meta: MenuMeta{
						Title:     "menus.operation.fleets",
						KeepAlive: true,
					},
				},
				requiredRole: RoleFC,
			},
			{
				MenuItem: MenuItem{
					Path:      "fleet-detail/:id",
					Name:      "FleetDetail",
					Component: "/operation/fleet-detail",
					Meta: MenuMeta{
						Title:  "menus.operation.fleetDetail",
						IsHide: true,
					},
				},
				requiredRole: RoleFC,
			},
			{
				MenuItem: MenuItem{
					Path:      "pap",
					Name:      "MyPap",
					Component: "/operation/pap",
					Meta: MenuMeta{
						Title:     "menus.operation.pap",
						KeepAlive: true,
					},
				},
				requiredRole: "", // 所有已登录用户
			},
			{
				MenuItem: MenuItem{
					Path:      "wallet",
					Name:      "Wallet",
					Component: "/operation/wallet",
					Meta: MenuMeta{
						Title:     "menus.operation.wallet",
						KeepAlive: true,
					},
				},
				requiredRole: "", // 所有已登录用户
			},
		},
	},
	{
		MenuItem: MenuItem{
			Path:      "/result",
			Name:      "Result",
			Component: "/index/index",
			Meta: MenuMeta{
				Title:  "menus.result.title",
				Icon:   "ri:checkbox-circle-line",
				IsHide: true,
			},
		},
		requiredRole: "",
		children: []*menuItemWithRoles{
			{
				MenuItem: MenuItem{
					Path:      "success",
					Name:      "ResultSuccess",
					Component: "/result/success",
					Meta: MenuMeta{
						Title:     "menus.result.success",
						KeepAlive: true,
						IsHide:    true,
					},
				},
				requiredRole: "",
			},
			{
				MenuItem: MenuItem{
					Path:      "fail",
					Name:      "ResultFail",
					Component: "/result/fail",
					Meta: MenuMeta{
						Title:     "menus.result.fail",
						KeepAlive: true,
						IsHide:    true,
					},
				},
				requiredRole: "",
			},
		},
	},
	{
		MenuItem: MenuItem{
			Path:      "/srp",
			Name:      "SRP",
			Component: "/index/index",
			Meta: MenuMeta{
				Title: "menus.srp.title",
				Icon:  "ri:money-dollar-box-line",
			},
		},
		requiredRole: "", // 所有已登录用户
		children: []*menuItemWithRoles{
			{
				MenuItem: MenuItem{
					Path:      "srp-apply",
					Name:      "SrpApply",
					Component: "/srp/apply",
					Meta: MenuMeta{
						Title:     "menus.srp.srpApply",
						KeepAlive: true,
					},
				},
				requiredRole: "", // 所有已登录用户
			},
			{
				MenuItem: MenuItem{
					Path:      "srp-manage",
					Name:      "SrpManage",
					Component: "/srp/manage",
					Meta: MenuMeta{
						Title:     "menus.srp.srpManage",
						KeepAlive: true,
					},
				},
				requiredRole: RoleSRP,
			},
			{
				MenuItem: MenuItem{
					Path:      "srp-prices",
					Name:      "SrpPrices",
					Component: "/srp/prices",
					Meta: MenuMeta{
						Title:     "menus.srp.srpPrices",
						KeepAlive: true,
					},
				},
				requiredRole: RoleSRP,
			},
		},
	},
}

// GetMenuByRole 根据角色返回过滤后的菜单树
func GetMenuByRole(role string) []*MenuItem {
	return filterMenus(allMenus, role)
}

// filterMenus 递归过滤菜单，只保留当前角色有权访问的项
func filterMenus(items []*menuItemWithRoles, role string) []*MenuItem {
	var result []*MenuItem
	for _, item := range items {
		// 校验父菜单权限
		if item.requiredRole != "" && !HasRole(role, item.requiredRole) {
			continue
		}

		node := item.MenuItem // 值复制，避免修改全局数据

		if len(item.children) > 0 {
			filteredChildren := filterMenus(item.children, role)
			if len(filteredChildren) > 0 {
				node.Children = filteredChildren
			}
		}

		result = append(result, &node)
	}
	return result
}
