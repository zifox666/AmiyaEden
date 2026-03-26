import request from '@/utils/http'

/** 上传图片，返回 base64 data URL（不写入文件系统，最大 2 MB） */
export function uploadImageAsDataUrl(file: File, url = '/api/v1/upload/image') {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<{ url: string }>({
    url,
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
