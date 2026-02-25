import { AppRouteRecord } from '@/types/router'

export const srpRoutes: AppRouteRecord = {
  path: '/srp',
  name: 'SRP',
  component: '/index/index',
  meta: {
    title: 'menus.srp.title',
    icon: 'ri:shield-user-line'
  },
  children: [
    {
      path: 'srp-apply',
      name: 'SrpApply',
      component: '/srp/apply',
      meta: {
        title: 'menus.srp.srpApply',
        keepAlive: true
      }
    },
    {
      path: 'srp-manage',
      name: 'SrpManage',
      component: '/srp/manage',
      meta: {
        title: 'menus.srp.srpManage',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN', 'R_SRP', 'R_FC']
      }
    },
    {
      path: 'srp-prices',
      name: 'SrpPrices',
      component: '/srp/prices',
      meta: {
        title: 'menus.srp.srpPrices',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN', 'R_SRP']
      }
    }
  ]
}
