import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { AdminEntitiesResponse, AdminEntityRow, AdminListResponse } from './types'

const token = () => authState.token

export function fetchAdminEntities() {
  return apiRequest<AdminEntitiesResponse>('/admin/entities', { token: token() })
}

export function fetchAdminRows(entity: string, params: { page?: number; perPage?: number; search?: string } = {}) {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.perPage) query.set('perPage', String(params.perPage))
  if (params.search) query.set('search', params.search)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return apiRequest<AdminListResponse>(`/admin/${entity}${suffix}`, { token: token() })
}

export function createAdminRow(entity: string, data: AdminEntityRow) {
  return apiRequest<{ item: AdminEntityRow }>(`/admin/${entity}`, {
    method: 'POST',
    token: token(),
    body: JSON.stringify({ data })
  })
}

export function updateAdminRow(entity: string, id: string, data: AdminEntityRow) {
  return apiRequest<{ item: AdminEntityRow }>(`/admin/${entity}/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify({ data })
  })
}

export function deleteAdminRow(entity: string, id: string) {
  return apiRequest<void>(`/admin/${entity}/${id}`, {
    method: 'DELETE',
    token: token()
  })
}
