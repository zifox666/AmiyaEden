import request from '@/utils/http'

/** 获取 Webhook 配置 */
export function fetchWebhookConfig() {
  return request.get<Api.Webhook.Config>({
    url: '/api/v1/system/webhook/config'
  })
}

/** 保存 Webhook 配置 */
export function setWebhookConfig(data: Api.Webhook.Config) {
  return request.put({
    url: '/api/v1/system/webhook/config',
    data
  })
}

/** 发送测试消息 */
export function testWebhook(data: {
  url: string
  type: string
  content?: string
  ob_target_type?: string
  ob_target_id?: number
  ob_token?: string
}) {
  return request.post({
    url: '/api/v1/system/webhook/test',
    data
  })
}
