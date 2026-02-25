import { AppRouteRecord } from '@/types/router'

export const roleRoutes: AppRouteRecord = {
  path: '/role',
  name: 'Role',
  component: '/index/index',
  meta: {
    title: 'menus.role.title',
    icon: 'ri:shield-user-line'
  },
  children: [
    {
      path: 'explore',
      name: 'RoleExplore',
      component: '/role/explore',
      meta: {
        title: 'menus.role.explore',
        keepAlive: true
      }
    },
    {
      path: 'my-roles',
      name: 'RoleMyRoles',
      component: '/role/my-roles',
      meta: {
        title: 'menus.role.myRoles',
        keepAlive: true
      }
    },
    {
      path: 'applications',
      name: 'RoleApplications',
      component: '/role/applications',
      meta: {
        title: 'menus.role.applications',
        keepAlive: true
      }
    },
    {
      path: 'manage',
      name: 'RoleManage',
      component: '/role/manage',
      meta: {
        title: 'menus.role.manage',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    }
  ]
}
