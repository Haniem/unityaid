import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { EventApplication, EventAttendance, EventFeedback, EventFeedbackResponse, EventItem, EventPayload, EventShift, EventTemplate } from './types'

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

export function createRecurringEvents(payload: { event: EventPayload; frequency: string; count: number }) {
  return apiRequest<{ items: EventItem[] }>('/events/recurring', {
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

export const fetchEventApplications = (id: string) =>
  apiRequest<{ items: EventApplication[] }>(`/events/${id}/applications`, { token: token() })

export const createEventApplication = (id: string, payload: { userId?: string; message?: string }) =>
  apiRequest<{ item: EventApplication }>(`/events/${id}/applications`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const updateEventApplication = (id: string, applicationId: string, status: EventApplication['status'], rejectionReason = '') =>
  apiRequest<{ item: EventApplication }>(`/events/${id}/applications/${applicationId}`, { method: 'PATCH', token: token(), body: JSON.stringify({ status, rejectionReason }) })

export const bulkUpdateEventApplications = (id: string, payload: { applicationIds: string[]; status: EventApplication['status']; rejectionReason?: string }) =>
  apiRequest<{ items: EventApplication[] }>(`/events/${id}/applications/bulk-status`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const deleteEventApplication = (id: string, applicationId: string) =>
  apiRequest<void>(`/events/${id}/applications/${applicationId}`, { method: 'DELETE', token: token() })

export const fetchEventAttendance = (id: string) =>
  apiRequest<{ items: EventAttendance[] }>(`/events/${id}/attendance`, { token: token() })

export const markEventAttendance = (id: string, payload: { userId: string; checkinCode?: string; hours?: number }) =>
  apiRequest<{ item: EventAttendance }>(`/events/${id}/attendance`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const bulkMarkEventAttendance = (id: string, payload: { userIds: string[]; hours?: number }) =>
  apiRequest<{ items: EventAttendance[] }>(`/events/${id}/attendance/bulk`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const updateEventAttendance = (id: string, attendanceId: string, payload: { hours?: number; checkOutAt?: string }) =>
  apiRequest<{ item: EventAttendance }>(`/events/${id}/attendance/${attendanceId}`, { method: 'PATCH', token: token(), body: JSON.stringify(payload) })

export const fetchEventShifts = (id: string) =>
  apiRequest<{ items: EventShift[] }>(`/events/${id}/shifts`, { token: token() })

export const createEventShift = (id: string, payload: { title: string; startsAt: string; endsAt: string; capacity?: number | null }) =>
  apiRequest<{ item: EventShift }>(`/events/${id}/shifts`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const fetchEventFeedback = (id: string) =>
  apiRequest<EventFeedbackResponse>(`/events/${id}/feedback`, { token: token() })

export const createEventFeedback = (id: string, payload: { rating: number; comment: string }) =>
  apiRequest<{ item: EventFeedback }>(`/events/${id}/feedback`, { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const completeEvent = (id: string) =>
  apiRequest<{ status: string }>(`/events/${id}/complete`, { method: 'POST', token: token() })

export const fetchEventTemplates = () =>
  apiRequest<{ items: EventTemplate[] }>('/event-templates', { token: token() })

export const createEventTemplate = (payload: {
  organizationId: string
  name: string
  title: string
  description: string
  format: EventPayload['format']
  location?: string | null
  maxParticipants?: number | null
  defaultDurationMinutes: number
}) =>
  apiRequest<{ item: EventTemplate }>('/event-templates', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
