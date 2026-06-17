import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type {
  AddOrganizationMemberPayload,
  Organization,
  OrganizationMember,
  OrganizationPayload,
  UpdateOrganizationMemberPayload
} from './types'

const token = () => authState.token

export function fetchOrganizations(params: { search?: string; includeDeleted?: boolean } = {}) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.includeDeleted) query.set('includeDeleted', 'true')
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return apiRequest<{ items: Organization[] }>(`/organizations${suffix}`, { token: token() })
}

export function fetchOrganization(id: string) {
  return apiRequest<{ item: Organization }>(`/organizations/${id}`, { token: token() })
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

export function fetchOrganizationMembers(id: string) {
  return apiRequest<{ items: OrganizationMember[] }>(`/organizations/${id}/members`, { token: token() })
}

export function addOrganizationMember(id: string, payload: AddOrganizationMemberPayload) {
  return apiRequest<{ item: OrganizationMember }>(`/organizations/${id}/members`, {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateOrganizationMember(id: string, memberId: string, payload: UpdateOrganizationMemberPayload) {
  return apiRequest<{ item: OrganizationMember }>(`/organizations/${id}/members/${memberId}`, {
    method: 'PATCH',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function removeOrganizationMember(id: string, memberId: string) {
  return apiRequest<void>(`/organizations/${id}/members/${memberId}`, {
    method: 'DELETE',
    token: token()
  })
}
