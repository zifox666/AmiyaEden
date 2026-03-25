import request from '@/utils/http'

/** 获取军团建筑列表 */
export function fetchCorpStructureList(data: Api.CorpStructure.ListRequest) {
  return request.post<Api.CorpStructure.StructureList>({
    url: '/api/v1/operation/corp-structures/list',
    data
  })
}

/** 获取用户关联的军团 ID 列表 */
export function fetchCorpIDs() {
  return request.get<number[]>({
    url: '/api/v1/operation/corp-structures/corps'
  })
}
