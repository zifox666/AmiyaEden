import { AppRouteRecord } from '@/types/router'

export const skillPlanningRoutes: AppRouteRecord = {
  path: '/skill-planning',
  name: 'SkillPlanning',
  component: '/index/index',
  meta: {
    title: 'menus.skillPlanning.title',
    icon: 'ri:brain-line',
    login: true
  },
  children: [
    {
      path: 'completion-check',
      name: 'SkillPlanCompletionCheck',
      component: '/skill-planning/completion-check',
      meta: {
        title: 'menus.skillPlanning.completionCheck',
        keepAlive: true,
        login: true
      }
    },
    {
      path: 'skill-plans',
      name: 'SkillPlans',
      component: '/skill-planning/skill-plans',
      meta: {
        title: 'menus.skillPlanning.skillPlans',
        keepAlive: true,
        roles: ['super_admin', 'admin', 'senior_fc', 'fc']
      }
    }
  ]
}
