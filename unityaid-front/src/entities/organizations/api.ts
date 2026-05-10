import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { Organization, OrganizationPayload } from './types'

const token = () => authState.token

export function fetchOrganizations() {
  return apiRequest<{ items: Organization[] }>('/organizations', { token: token() })
}

export function createOrganization(payload: OrganizationPayload) {
  return apiRequest<{ item: Organization }>('/organizations', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateOrganization(id: string, payload: OrganizationPayload) {
  return apiRequest<{ item: Organization }>(`/organizations/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function deleteOrganization(id: string) {
  return apiRequest<void>(`/organizations/${id}`, {
    method: 'DELETE',
    token: token()
  })
}
