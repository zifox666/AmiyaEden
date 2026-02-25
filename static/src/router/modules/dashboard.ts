import { AppRouteRecord } from '@/types/router'

export const dashboardRoutes: AppRouteRecord = {
  name: 'Dashboard',
  path: '/dashboard',
  component: '/index/index',
  meta: {
    title: 'menus.dashboard.title',
    icon: 'ri:pie-chart-line',
    roles: ['super_admin', 'admin', 'srp', 'fc', 'user', 'guest']
  },
  children: [
    {
      path: 'console',
      name: 'Console',
      component: '/dashboard/console',
      meta: {
        title: 'menus.dashboard.console',
        keepAlive: false,
        fixedTab: true
      }
    },
    {
      path: 'characters',
      name: 'Characters',
      component: '/dashboard/characters',
      meta: {
        title: 'menus.characters.title',
        keepAlive: true
      }
    },
  ]
}
