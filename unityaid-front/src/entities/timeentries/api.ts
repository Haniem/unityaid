import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { TimeEntry, TimeEntryPayload } from './types'

const token = () => authState.token

export function fetchTimeEntries(params: { status?: string; userId?: string } = {}) {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.userId) query.set('userId', params.userId)
  return apiRequest<{ items: TimeEntry[] }>(`/time-entries${query.toString() ? `?${query}` : ''}`, { token: token() })
}

export function createTimeEntry(payload: TimeEntryPayload) {
  return apiRequest<{ item: TimeEntry }>('/time-entries', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateTimeEntry(id: string, payload: TimeEntryPayload) {
  return apiRequest<{ item: TimeEntry }>(`/time-entries/${id}`, {
    method: 'PATCH',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export const approveTimeEntry = (id: string) =>
  apiRequest<{ item: TimeEntry }>(`/time-entries/${id}/approve`, { method: 'POST', token: token() })

export const rejectTimeEntry = (id: string) =>
  apiRequest<{ item: TimeEntry }>(`/time-entries/${id}/reject`, { method: 'POST', token: token() })
