import { AppRouteRecord } from '@/types/router'

export const infoRoutes: AppRouteRecord = {
  path: '/info',
  name: 'EveInfo',
  component: '/index/index',
  meta: { title: 'menus.info.title', icon: 'ri:user-star-line' },
  children: [
    {
      path: 'wallet',
      name: 'EveInfoWallet',
      component: '/info/wallet',
      meta: { title: 'menus.info.wallet', keepAlive: true }
    },
    {
      path: 'skill',
      name: 'EveInfoSkill',
      component: '/info/skill',
      meta: { title: 'menus.info.skill', keepAlive: true }
    }
  ]
}
