import assert from 'node:assert/strict'
import test from 'node:test'
import type { AppRouteRecord } from '../../types/router'
import { newbroRoutes as actualNewbroRoutes } from '../modules/newbro'
import { skillPlanningRoutes } from '../modules/skill-planning'
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
