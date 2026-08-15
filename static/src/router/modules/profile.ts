import { AppRouteRecord } from '@/types/router'

export const profileRoutes: AppRouteRecord = {
  path: '/profile',
  name: 'UserCenter',
  component: '/system/user-center',
  meta: {
    title: 'menus.system.userCenter',
    icon: 'ri:user-settings-line',
    keepAlive: true
  }
}
