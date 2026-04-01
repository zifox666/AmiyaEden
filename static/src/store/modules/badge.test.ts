import assert from 'node:assert/strict'
import test from 'node:test'
import type { AppRouteRecord } from '../../types/router'
import { applyBadgeCountsToMenu } from './badge.helpers'

function createMenuList(): AppRouteRecord[] {
  return [
    {
      path: '/newbro',
      name: 'NewbroRoot',
      component: '/index/index',
      meta: { title: 'menus.newbro.title' },
      children: [
        {
          path: 'mentor',
          name: 'NewbroMentorDashboard',
          component: '/newbro/mentor',
          meta: { title: 'menus.newbro.mentor' }
        }
      ]
    },
    {
      path: '/welfare',
      name: 'WelfareRoot',
      component: '/index/index',
      meta: { title: 'menus.welfare.title' },
      children: [
        {
          path: 'my',
          name: 'WelfareMy',
          component: '/welfare/my',
          meta: { title: 'menus.welfare.my', showTextBadge: '99' }
        },
        {
          path: 'approval',
          name: 'WelfareApproval',
          component: '/welfare/approval',
          meta: { title: 'menus.welfare.approval', showTextBadge: '88' }
        },
        {
          path: 'settings',
          name: 'WelfareSettings',
          component: '/welfare/settings',
          meta: { title: 'menus.welfare.settings', showTextBadge: '77' }
        }
      ]
    },
    {
      path: '/srp',
      name: 'SRP',
      component: '/index/index',
      meta: { title: 'menus.srp.title' },
      children: [
        {
          path: 'srp-manage',
          name: 'SrpManage',
          component: '/srp/manage',
          meta: { title: 'menus.srp.srpManage' }
        }
      ]
    },
    {
      path: '/shop',
      name: 'ShopRoot',
      component: '/index/index',
      meta: { title: 'menus.shop.title' },
      children: [
        {
          path: 'order-manage',
          name: 'ShopOrderManage',
          component: '/shop/order-manage',
          meta: { title: 'menus.shop.orderManage' }
        }
      ]
    }
  ]
}

test('applyBadgeCountsToMenu maps counts to leaves and sums parent badges', () => {
  const menuList = createMenuList()

  applyBadgeCountsToMenu(menuList, {
    mentor_pending_applications: 4,
    welfare_eligible: 2,
    welfare_pending: 3,
    srp_pending: 5,
    order_pending: 1
  })

  assert.equal(menuList[0].meta.showTextBadge, '4')
  assert.equal(menuList[0].children?.[0].meta.showTextBadge, '4')
  assert.equal(menuList[1].meta.showTextBadge, '5')
  assert.equal(menuList[1].children?.[0].meta.showTextBadge, '2')
  assert.equal(menuList[1].children?.[1].meta.showTextBadge, '3')
  assert.equal(menuList[2].meta.showTextBadge, '5')
  assert.equal(menuList[2].children?.[0].meta.showTextBadge, '5')
  assert.equal(menuList[3].meta.showTextBadge, '1')
  assert.equal(menuList[3].children?.[0].meta.showTextBadge, '1')
})

test('applyBadgeCountsToMenu clears missing or zero badge counts', () => {
  const menuList = createMenuList()

  applyBadgeCountsToMenu(menuList, {})

  assert.equal(menuList[0].meta.showTextBadge, undefined)
  assert.equal(menuList[0].children?.[0].meta.showTextBadge, undefined)
  assert.equal(menuList[1].meta.showTextBadge, undefined)
  assert.equal(menuList[1].children?.[0].meta.showTextBadge, undefined)
  assert.equal(menuList[1].children?.[1].meta.showTextBadge, undefined)
  assert.equal(menuList[1].children?.[2].meta.showTextBadge, undefined)
})
