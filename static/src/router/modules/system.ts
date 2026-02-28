import { AppRouteRecord } from '@/types/router'

export const systemRoutes: AppRouteRecord = {
  path: '/system',
  name: 'System',
  component: '/index/index',
  meta: {
    title: 'menus.system.title',
    icon: 'ri:user-3-line',
    roles: ['super_admin', 'admin']
  },
  children: [
    {
      path: 'user',
      name: 'User',
      component: '/system/user',
      meta: {
        title: 'menus.system.user',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'esi-refresh',
      name: 'ESIRefresh',
      component: '/system/esi-refresh',
      meta: {
        title: 'menus.system.esiRefresh',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'user-center',
      name: 'UserCenter',
      component: '/system/user-center',
      meta: {
        title: 'menus.system.userCenter',
        isHide: true,
        keepAlive: true,
        isHideTab: true
      }
    },
    {
      path: 'menu',
      name: 'Menus',
      component: '/system/menu',
      meta: {
        title: 'menus.system.menu',
        keepAlive: true,
        roles: ['super_admin'],
        authList: [
          { title: '新增', authMark: 'add' },
          { title: '编辑', authMark: 'edit' },
          { title: '删除', authMark: 'delete' }
        ]
      }
    },
    {
      path: 'wallet',
      name: 'SystemWallet',
      component: '/system/wallet',
      meta: {
        title: 'menus.system.wallet',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'pap',
      name: 'AlliancePAP',
      component: '/system/pap',
      meta: {
        title: 'menus.system.alliancePap',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    }
  ]
}
