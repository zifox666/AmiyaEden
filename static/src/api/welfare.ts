import request from '@/utils/http'
import { uploadImageAsDataUrl } from '@/api/upload'

/** 上传福利证明图片（含示例），返回 base64 data URL（不写入文件系统，最大 2 MB） */
export function uploadWelfareEvidence(file: File) {
  return uploadImageAsDataUrl(file, '/api/v1/welfare/upload-evidence')
}

// ─── 管理员福利设置 ───

/** 管理员查询福利列表 */
export function adminListWelfares(data?: Api.Welfare.SearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Welfare.WelfareItem>>({
    url: '/api/v1/system/welfare/list',
    data: data ?? { current: 1, size: 20 }
  })
}

/** 管理员创建福利 */
export function adminCreateWelfare(data: Api.Welfare.CreateParams) {
  return request.post<Api.Welfare.WelfareItem>({
    url: '/api/v1/system/welfare/add',
    data
  })
}

/** 管理员更新福利 */
export function adminUpdateWelfare(data: Api.Welfare.UpdateParams) {
  return request.post<Api.Welfare.WelfareItem>({
    url: '/api/v1/system/welfare/edit',
    data
  })
}

/** 管理员批量更新福利排序 */
export function adminReorderWelfares(ids: number[]) {
  return request.post({
    url: '/api/v1/system/welfare/reorder',
    data: { ids }
  })
}

/** 管理员删除福利 */
export function adminDeleteWelfare(id: number) {
  return request.post({
    url: '/api/v1/system/welfare/delete',
    data: { id }
  })
}

/** 管理员导入福利历史记录 */
export function adminImportWelfareRecords(data: Api.Welfare.ImportRecordsParams) {
  return request.post<Api.Welfare.ImportRecordsResult>({
    url: '/api/v1/system/welfare/import',
    data
  })
}

// ─── 管理端福利审批 ───

/** 管理端查询福利申请列表 */
export function adminListApplications(data?: Api.Welfare.AdminApplicationSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Welfare.AdminApplication>>({
    url: '/api/v1/system/welfare/applications',
    data: data ?? { current: 1, size: 50 }
  })
}

/** 管理端审批福利申请 */
export function adminReviewApplication(data: Api.Welfare.ReviewParams) {
  return request.post({
    url: '/api/v1/system/welfare/review',
    data
  })
}

/** 管理端删除福利申请记录（仅 admin） */
export function adminDeleteApplication(id: number) {
  return request.post({
    url: '/api/v1/system/welfare/applications/delete',
    data: { id }
  })
}

// ─── 用户端福利 ───

/** 获取可申请的福利列表 */
export function getEligibleWelfares() {
  return request.post<Api.Welfare.EligibleWelfare[]>({
    url: '/api/v1/welfare/eligible'
  })
}

/** 申请福利 */
export function applyForWelfare(data: Api.Welfare.ApplyParams) {
  return request.post({
    url: '/api/v1/welfare/apply',
    data
  })
}

/** 查询我的福利申请 */
export function getMyApplications(data?: Api.Welfare.MyApplicationSearchParams) {
  return request.post<Api.Common.PaginatedResponse<Api.Welfare.MyApplication>>({
    url: '/api/v1/welfare/my-applications',
    data
  })
}
