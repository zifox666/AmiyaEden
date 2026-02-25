import { AppRouteRecord } from '@/types/router'

export const operationRoutes: AppRouteRecord = {
  path: '/operation',
  name: 'Operation',
  component: '/index/index',
  meta: {
    title: 'menus.operation.title',
    icon: 'ri:ship-line'
  },
  children: [
    {
      path: 'fleets',
      name: 'Fleets',
      component: '/operation/fleets',
      meta: {
        title: 'menus.operation.fleets',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN', 'R_FC']
      }
    },
    {
      path: 'fleet-detail/:id',
      name: 'FleetDetail',
      component: '/operation/fleet-detail',
      meta: {
        title: 'menus.operation.fleetDetail',
        isHide: true,
        roles: ['R_SUPER', 'R_ADMIN', 'R_FC']
      }
    },
    {
      path: 'pap',
      name: 'MyPap',
      component: '/operation/pap',
      meta: {
        title: 'menus.operation.pap',
        keepAlive: true
      }
    },
    {
      path: 'wallet',
      name: 'Wallet',
      component: '/operation/wallet',
      meta: {
        title: 'menus.operation.wallet',
        keepAlive: true
      }
    },
    {
      path: 'join',
      name: 'JoinFleet',
      component: '/operation/join',
      meta: {
        title: 'menus.operation.join',
        isHide: true
      }
    }
  ]
}
