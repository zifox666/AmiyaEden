import { AppRouteRecord } from '@/types/router'
import { dashboardRoutes } from './dashboard'
import { systemRoutes } from './system'
import { operationRoutes } from './operation'
import { resultRoutes } from './result'
import { exceptionRoutes } from './exception'
import { srpRoutes } from './srp'

/**
 * 导出所有模块化路由
 */
export const routeModules: AppRouteRecord[] = [
  dashboardRoutes,
  operationRoutes,
  systemRoutes,
  resultRoutes,
  exceptionRoutes,
  srpRoutes
]
