import { AppRouteRecord } from '@/types/router'

export const dashboardRoutes: AppRouteRecord = {
  name: 'Dashboard',
  path: '/dashboard',
  component: '/index/index',
  meta: {
    title: 'menus.dashboard.title',
    icon: 'ri:pie-chart-line'
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
    {
      path: 'npc-kills',
      name: 'CorpNpcKillReport',
      component: '/dashboard/npc-kills',
      meta: {
        title: 'menus.dashboard.npcKills',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    }
  ]
}
