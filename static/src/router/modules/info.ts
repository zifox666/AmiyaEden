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
    },
    {
      path: 'ships',
      name: 'EveInfoShips',
      component: '/info/ships',
      meta: { title: 'menus.info.ships', keepAlive: true }
    },
    {
      path: 'implants',
      name: 'EveInfoImplants',
      component: '/info/implants',
      meta: { title: 'menus.info.implants', keepAlive: true }
    },
    {
      path: 'fittings',
      name: 'EveInfoFittings',
      component: '/info/fittings',
      meta: { title: 'menus.info.fittings', keepAlive: true }
    },
    {
      path: 'npc-kills',
      name: 'NpcKillReport',
      component: '/info/npc-kills',
      meta: { title: 'menus.info.npcKills', keepAlive: true }
    }
  ]
}
