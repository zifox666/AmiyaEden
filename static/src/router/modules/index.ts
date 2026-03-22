import { AppRouteRecord } from '@/types/router'
import { dashboardRoutes } from './dashboard'
import { systemRoutes } from './system'
import { operationRoutes } from './operation'
import { skillPlanningRoutes } from './skill-planning'
import { exceptionRoutes } from './exception'
import { srpRoutes } from './srp'
import { shopRoutes } from './shop'
import { infoRoutes } from './info'

/**
 * 导出所有模块化路由
 */
export const routeModules: AppRouteRecord[] = [
  dashboardRoutes,
  operationRoutes,
  skillPlanningRoutes,
  infoRoutes,
  shopRoutes,
  systemRoutes,
  exceptionRoutes,
  srpRoutes
]
