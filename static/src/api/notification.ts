import request from '@/utils/http'

/** 获取通知列表（当前用户所有角色） */
export function fetchNotifications(data?: Api.Notification.ListParams) {
  return request.post<Api.Notification.NotificationSummary>({
    url: '/api/v1/notification/list',
    data: data ?? {}
  })
}

/** 获取未读通知数量 */
export function fetchUnreadCount() {
  return request.post<Api.Notification.UnreadCountResponse>({
    url: '/api/v1/notification/unread-count'
  })
}

/** 标记指定通知为已读 */
export function markAsRead(data: Api.Notification.MarkAsReadParams) {
  return request.post({
    url: '/api/v1/notification/read',
    data
  })
}

/** 标记所有通知为已读 */
export function markAllAsRead() {
  return request.post({
    url: '/api/v1/notification/read-all'
  })
}
