import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { NotificationsResponse } from './types'

const token = () => authState.token

export const fetchNotifications = () =>
  apiRequest<NotificationsResponse>('/notifications', { token: token() })

export const markNotificationRead = (id: string) =>
  apiRequest<{ status: string }>(`/notifications/${id}/read`, {
    method: 'POST',
    token: token()
  })

export const markAllNotificationsRead = () =>
  apiRequest<{ status: string }>('/notifications/read-all', {
    method: 'POST',
    token: token()
  })
