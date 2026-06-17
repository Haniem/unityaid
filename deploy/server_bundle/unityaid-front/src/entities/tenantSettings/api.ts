import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { TenantSettings, TenantSettingsPayload } from './types'

const token = () => authState.token

export function fetchTenantSettings() {
  return apiRequest<{ item: TenantSettings }>('/tenant-settings', { token: token() })
}

export function updateTenantSettings(payload: TenantSettingsPayload) {
  return apiRequest<{ item: TenantSettings }>('/tenant-settings', {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })
}
