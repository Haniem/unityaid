import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { ProfileField, ProfileFieldGroup, ProfileFieldPayload, ProfileGroupPayload, ProfileValue } from './types'

const token = () => authState.token

export function fetchProfileSchema(organizationId: string) {
  return apiRequest<{ items: ProfileFieldGroup[] }>(`/profile-fields/schema?organizationId=${encodeURIComponent(organizationId)}`, { token: token() })
}

export function fetchProfileValues(organizationId: string, userId: string) {
  return apiRequest<{ values: Record<string, ProfileValue> }>(`/profile-fields/users/${userId}?organizationId=${encodeURIComponent(organizationId)}`, { token: token() })
}

export function saveProfileValues(organizationId: string, userId: string, values: Record<string, ProfileValue>) {
  return apiRequest<{ values: Record<string, ProfileValue> }>(`/profile-fields/users/${userId}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify({ organizationId, values })
  })
}

export function createProfileGroup(payload: ProfileGroupPayload) {
  return apiRequest<{ item: ProfileFieldGroup }>('/profile-fields/groups', { method: 'POST', token: token(), body: JSON.stringify(payload) })
}

export function updateProfileGroup(id: string, payload: ProfileGroupPayload) {
  return apiRequest<{ item: ProfileFieldGroup }>(`/profile-fields/groups/${id}`, { method: 'PUT', token: token(), body: JSON.stringify(payload) })
}

export function deleteProfileGroup(id: string, organizationId: string) {
  return apiRequest<void>(`/profile-fields/groups/${id}?organizationId=${encodeURIComponent(organizationId)}`, { method: 'DELETE', token: token() })
}

export function createProfileField(payload: ProfileFieldPayload) {
  return apiRequest<{ item: ProfileField }>('/profile-fields/fields', { method: 'POST', token: token(), body: JSON.stringify(payload) })
}

export function updateProfileField(id: string, payload: ProfileFieldPayload) {
  return apiRequest<{ item: ProfileField }>(`/profile-fields/fields/${id}`, { method: 'PUT', token: token(), body: JSON.stringify(payload) })
}

export function deleteProfileField(id: string, organizationId: string) {
  return apiRequest<void>(`/profile-fields/fields/${id}?organizationId=${encodeURIComponent(organizationId)}`, { method: 'DELETE', token: token() })
}
