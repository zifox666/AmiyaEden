import { AppRouteRecord } from '@/types/router'

export const operationRoutes: AppRouteRecord = {
  path: '/operation',
  name: 'Operation',
  component: '/index/index',
  meta: {
    title: 'menus.operation.title',
    icon: 'ri:ship-line',
    login: true
  },
  children: [
    {
      path: 'fleets',
      name: 'Fleets',
      component: '/operation/fleets',
      meta: {
        title: 'menus.operation.fleets',
        keepAlive: true,
        roles: ['super_admin', 'admin', 'fc', 'senior_fc']
      }
    },
    {
      path: 'fleet-configs',
      name: 'FleetConfigs',
      component: '/operation/fleet-configs',
      meta: {
        title: 'menus.operation.fleetConfigs',
        keepAlive: true,
        login: true
      }
    },
    {
      path: 'fleet-detail/:id',
      name: 'FleetDetail',
      component: '/operation/fleet-detail',
      meta: {
        title: 'menus.operation.fleetDetail',
        isHide: true,
        roles: ['super_admin', 'admin', 'fc', 'senior_fc']
      }
    },
    {
      path: 'corporation-pap',
      name: 'CorporationPap',
      component: '/operation/corporation-pap',
      meta: {
        title: 'menus.operation.corporationPap',
        keepAlive: true,
        login: true
      }
    },
    {
      path: 'pap',
      name: 'MyPap',
      component: '/operation/pap',
      meta: {
        title: 'menus.operation.pap',
        keepAlive: true,
        login: true
      }
    },
    {
      path: 'join',
      name: 'JoinFleet',
      component: '/operation/join',
      meta: {
        title: 'menus.operation.join',
        isHide: true,
        login: true
      }
    }
  ]
}
