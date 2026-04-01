import { AppRouteRecord } from '@/types/router'

export const newbroRoutes: AppRouteRecord = {
  path: '/newbro',
  name: 'NewbroRoot',
  component: '/index/index',
  meta: {
    title: 'menus.newbro.title',
    icon: 'ri:user-heart-line',
    login: true
  },
  children: [
    {
      path: 'select-captain',
      name: 'NewbroSelectCaptain',
      component: '/newbro/select-captain',
      meta: {
        title: 'menus.newbro.selectCaptain',
        keepAlive: true,
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
        keepAlive: true,
        login: true,
        requiresMentorMenteeEligibility: true
      }
    },
    {
      path: 'captain',
      name: 'NewbroCaptainDashboard',
      component: '/newbro/captain',
      meta: {
        title: 'menus.newbro.captain',
        keepAlive: true,
        roles: ['captain']
      }
    },
    {
      path: 'mentor',
      name: 'NewbroMentorDashboard',
      component: '/newbro/mentor',
      meta: {
        title: 'menus.newbro.mentor',
        keepAlive: true,
        roles: ['mentor']
      }
    },
    {
      path: 'manage',
      name: 'NewbroManage',
      component: '/newbro/manage',
      meta: {
        title: 'menus.newbro.manage',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    },
    {
      path: 'mentor-manage',
      name: 'MentorManage',
      component: '/newbro/mentor-manage',
      meta: {
        title: 'menus.newbro.mentorManage',
        keepAlive: true,
        roles: ['super_admin', 'admin']
      }
    }
  ]
}
