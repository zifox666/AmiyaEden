import { AppRouteRecord } from '@/types/router'

export const shopRoutes: AppRouteRecord = {
  path: '/shop',
  name: 'ShopRoot',
  component: '/index/index',
  meta: {
    title: 'menus.shop.title',
    icon: 'ri:shopping-bag-line',
    login: true
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
        roles: ['super_admin', 'admin'],
        authList: [
          { title: '新增商品', authMark: 'add_product' },
          { title: '编辑商品', authMark: 'edit_product' },
          { title: '删除商品', authMark: 'delete_product' }
        ]
      }
    },
    {
      path: 'order-manage',
      name: 'ShopOrderManage',
      component: '/shop/order-manage',
      meta: {
        title: 'menus.shop.orderManage',
        keepAlive: true,
        roles: ['super_admin', 'admin'],
        authList: [{ title: '审批订单', authMark: 'approve_order' }]
      }
    },
    {
      path: 'wallet',
      name: 'Wallet',
      component: '/shop/wallet',
      meta: {
        title: 'menus.shop.wallet',
        keepAlive: true
      }
    }
  ]
}
