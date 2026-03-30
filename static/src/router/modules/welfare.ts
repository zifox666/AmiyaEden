import { AppRouteRecord } from '@/types/router'

export const welfareRoutes: AppRouteRecord = {
  path: '/welfare',
  name: 'WelfareRoot',
  component: '/index/index',
  meta: {
    title: 'menus.welfare.title',
    icon: 'ri:gift-line',
    login: true
  },
  children: [
    {
      path: 'my',
      name: 'WelfareMy',
      component: '/welfare/my',
      meta: {
        title: 'menus.welfare.my',
        keepAlive: true,
        login: true
      }
    },
    {
      path: 'approval',
      name: 'WelfareApproval',
      component: '/welfare/approval',
      meta: {
        title: 'menus.welfare.approval',
        keepAlive: true,
        roles: ['super_admin', 'admin', 'welfare']
      }
    },
    {
      path: 'settings',
      name: 'WelfareSettings',
      component: '/welfare/settings',
      meta: {
        title: 'menus.welfare.settings',
        keepAlive: true,
        roles: ['super_admin', 'admin', 'welfare']
      }
    }
  ]
}
