import request from '@/utils/http'

/** 获取所有 ESI 刷新任务定义 */
export function fetchESIRefreshTasks() {
  return request.get<Api.ESIRefresh.TaskInfo[]>({
    url: '/api/v1/esi/refresh/tasks'
  })
}

/** 获取任务运行时状态（分页 + 筛选） */
export function fetchESIRefreshStatuses(params?: Api.ESIRefresh.TaskStatusSearchParams) {
  return request.get<Api.ESIRefresh.TaskStatusList>({
    url: '/api/v1/esi/refresh/statuses',
    params
  })
}

/** 手动触发指定任务（单角色） */
export function runESIRefreshTask(params: Api.ESIRefresh.RunTaskParams) {
  return request.post<{ message: string }>({
    url: '/api/v1/esi/refresh/run',
    data: params
  })
}

/** 手动触发指定任务（所有角色） */
export function runESIRefreshTaskByName(params: Api.ESIRefresh.RunTaskByNameParams) {
  return request.post<{ message: string }>({
    url: '/api/v1/esi/refresh/run-task',
    data: params
  })
}

/** 手动触发全量刷新 */
export function runESIRefreshAll() {
  return request.post<{ message: string }>({
    url: '/api/v1/esi/refresh/run-all'
  })
}
