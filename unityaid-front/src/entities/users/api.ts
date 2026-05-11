import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { Skill, SkillPayload, User, VolunteerPayload, VolunteerProfile } from './types'

const token = () => authState.token

export function fetchUsers(params: { search?: string; role?: string; organizationId?: string } = {}) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.role) query.set('role', params.role)
  if (params.organizationId) query.set('organizationId', params.organizationId)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return apiRequest<{ items: User[] }>(`/users${suffix}`, { token: token() })
}

export function fetchVolunteers(params: { search?: string; skillId?: string; organizationId?: string } = {}) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.skillId) query.set('skillId', params.skillId)
  if (params.organizationId) query.set('organizationId', params.organizationId)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return apiRequest<{ items: VolunteerProfile[] }>(`/volunteers${suffix}`, { token: token() })
}

export function fetchVolunteer(userId: string) {
  return apiRequest<{ item: VolunteerProfile }>(`/volunteers/${userId}`, { token: token() })
}

export function updateVolunteer(userId: string, payload: VolunteerPayload) {
  return apiRequest<{ item: VolunteerProfile }>(`/volunteers/${userId}`, {
    method: 'PATCH',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function fetchSkills() {
  return apiRequest<{ items: Skill[] }>('/skills', { token: token() })
}

export function createSkill(payload: SkillPayload) {
  return apiRequest<{ item: Skill }>('/skills', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function deleteSkill(id: string) {
  return apiRequest<void>(`/skills/${id}`, {
    method: 'DELETE',
    token: token()
  })
}
