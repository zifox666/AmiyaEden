import assert from 'node:assert/strict'
import test from 'node:test'
import type { AppRouteRecord } from '../../types/router'
import { dashboardRoutes } from '../modules/dashboard'
import { newbroRoutes as actualNewbroRoutes } from '../modules/newbro'
import { skillPlanningRoutes } from '../modules/skill-planning'
import { shopRoutes } from '../modules/shop'
import { srpRoutes } from '../modules/srp'
import { systemRoutes } from '../modules/system'
import { applyMenuAccessFilter, pruneEmptyMenus } from './menuAccess'

const newbroRoutes: AppRouteRecord[] = [
  {
    path: '/newbro',
    name: 'NewbroRoot',
    component: '/index/index',
    meta: {
      title: 'menus.newbro.title',
      login: true
    },
    children: [
      {
        path: 'select-captain',
        name: 'NewbroSelectCaptain',
        component: '/newbro/select-captain',
        meta: {
          title: 'menus.newbro.selectCaptain',
          login: true,
          requiresNewbro: true
        }
      },
      {
        path: 'select-mentor',
        name: 'NewbroSelectMentor',
        component: '/newbro/select-mentor',
        meta: {
          title: 'menus.newbro.selectMentor',
          login: true,
          requiresMentorMenteeEligibility: true
        }
      }
    ]
  }
]

test('applyMenuAccessFilter hides requiresNewbro routes when status is unknown', () => {
  const filtered = applyMenuAccessFilter(newbroRoutes, ['user'], undefined)

  assert.deepEqual(filtered, [
    {
      path: '/newbro',
      name: 'NewbroRoot',
      component: '/index/index',
      meta: {
        title: 'menus.newbro.title',
        login: true
      },
      children: []
    }
  ])
})

test('pruneEmptyMenus removes directories whose children were fully filtered out', () => {
  const filtered = applyMenuAccessFilter(newbroRoutes, ['user'], undefined)
  const pruned = pruneEmptyMenus(filtered)

  assert.deepEqual(pruned, [])
})

test('applyMenuAccessFilter keeps SkillPlans for logged-in ordinary users', () => {
  const filtered = applyMenuAccessFilter([skillPlanningRoutes], ['user'], undefined)
  const skillPlanning = filtered[0]

  assert.equal(
    skillPlanning.children?.some((route) => route.name === 'SkillPlans'),
    true
  )
})

test('applyMenuAccessFilter hides mentor selection routes when mentor eligibility is unknown', () => {
  const filtered = applyMenuAccessFilter(newbroRoutes, ['user'], true, undefined)

  assert.equal(
    filtered[0].children?.some((route) => route.name === 'NewbroSelectMentor'),
    false
  )
})

test('newbro mentor selection route requires mentor mentee eligibility', () => {
  const mentorRoute = actualNewbroRoutes.children?.find(
    (route) => route.name === 'NewbroSelectMentor'
  )

  assert.equal(mentorRoute?.meta.requiresMentorMenteeEligibility, true)
})

test('applyMenuAccessFilter hides AutoRole from admins but keeps it for super admins', () => {
  const adminFiltered = applyMenuAccessFilter([systemRoutes], ['admin'])
  const adminSystemMenu = adminFiltered[0]

  assert.equal(
    adminSystemMenu.children?.some((route) => route.name === 'AutoRole'),
    false
  )

  const superAdminFiltered = applyMenuAccessFilter([systemRoutes], ['super_admin'])
  const superAdminSystemMenu = superAdminFiltered[0]

  assert.equal(
    superAdminSystemMenu.children?.some((route) => route.name === 'AutoRole'),
    true
  )
})

test('applyMenuAccessFilter hides BasicConfig from admins but keeps it for super admins', () => {
  const adminFiltered = applyMenuAccessFilter([systemRoutes], ['admin'])
  const adminSystemMenu = adminFiltered[0]

  assert.equal(
    adminSystemMenu.children?.some((route) => route.name === 'BasicConfig'),
    false
  )

  const superAdminFiltered = applyMenuAccessFilter([systemRoutes], ['super_admin'])
  const superAdminSystemMenu = superAdminFiltered[0]

  assert.equal(
    superAdminSystemMenu.children?.some((route) => route.name === 'BasicConfig'),
    true
  )
})

test('CorpNpcKillReport lives under Dashboard for admins only', () => {
  const adminDashboard = applyMenuAccessFilter([dashboardRoutes], ['admin'])[0]
  const userDashboard = applyMenuAccessFilter([dashboardRoutes], ['user'])[0]
  const adminSystem = applyMenuAccessFilter([systemRoutes], ['admin'])[0]

  const adminNpcKillsRoute = adminDashboard.children?.find(
    (route) => route.name === 'CorpNpcKillReport'
  )

  assert.equal(adminNpcKillsRoute?.path, 'npc-kills')
  assert.deepEqual(adminNpcKillsRoute?.meta.roles, ['super_admin', 'admin'])
  assert.equal(
    userDashboard.children?.some((route) => route.name === 'CorpNpcKillReport'),
    false
  )
  assert.equal(
    adminSystem.children?.some((route) => route.name === 'CorpNpcKillReport'),
    false
  )
})

test('applyMenuAccessFilter keeps SRP prices for SRP, admin, senior fc, and super admins', () => {
  const adminSrpMenu = applyMenuAccessFilter([srpRoutes], ['admin'])[0]
  const seniorFCSrpMenu = applyMenuAccessFilter([srpRoutes], ['senior_fc'])[0]
  const srpOfficerMenu = applyMenuAccessFilter([srpRoutes], ['srp'])[0]
  const superAdminSrpMenu = applyMenuAccessFilter([srpRoutes], ['super_admin'])[0]

  assert.equal(
    adminSrpMenu.children?.some((route) => route.name === 'SrpPrices'),
    true
  )
  assert.equal(
    seniorFCSrpMenu.children?.some((route) => route.name === 'SrpPrices'),
    true
  )
  assert.equal(
    srpOfficerMenu.children?.some((route) => route.name === 'SrpPrices'),
    true
  )
  assert.equal(
    superAdminSrpMenu.children?.some((route) => route.name === 'SrpPrices'),
    true
  )
})

test('applyMenuAccessFilter hides ShopOrderManage from welfare officers', () => {
  const welfareShopMenu = applyMenuAccessFilter([shopRoutes], ['welfare'])[0]

  assert.equal(
    welfareShopMenu.children?.some((route) => route.name === 'ShopOrderManage'),
    false
  )
})
