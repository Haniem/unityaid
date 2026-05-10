import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { EventItem, EventPayload } from './types'

const token = () => authState.token

export function fetchEvents() {
  return apiRequest<{ items: EventItem[] }>('/events', { token: token() })
}

export function fetchEvent(id: string) {
  return apiRequest<{ item: EventItem }>(`/events/${id}`, { token: token() })
}

export function createEvent(payload: EventPayload) {
  return apiRequest<{ item: EventItem }>('/events', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateEvent(id: string, payload: EventPayload) {
  return apiRequest<{ item: EventItem }>(`/events/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function deleteEvent(id: string) {
  return apiRequest<void>(`/events/${id}`, {
    method: 'DELETE',
    token: token()
  })
}
