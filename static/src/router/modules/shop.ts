import { AppRouteRecord } from '@/types/router'

export const shopRoutes: AppRouteRecord = {
  path: '/shop',
  name: 'ShopRoot',
  component: '/index/index',
  meta: {
    title: 'menus.shop.title',
    icon: 'ri:shopping-bag-line'
  },
  children: [
    {
      path: 'browse',
      name: 'Shop',
      component: '/shop/browse',
      meta: {
        title: 'menus.shop.browse',
        keepAlive: true
      }
    },
    {
      path: 'manage',
      name: 'ShopManage',
      component: '/shop/manage',
      meta: {
        title: 'menus.shop.manage',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    }
  ]
}
