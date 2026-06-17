import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { TaskAttachment, TaskComment, TaskItem, TaskPayload, TaskStatusHistory, TaskTimeEntry } from './types'

const token = () => authState.token

export function fetchTasks(params: { status?: string; priority?: string; assigneeId?: string } = {}) {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.priority) query.set('priority', params.priority)
  if (params.assigneeId) query.set('assigneeId', params.assigneeId)
  return apiRequest<{ items: TaskItem[] }>(`/tasks${query.toString() ? `?${query}` : ''}`, { token: token() })
}

export function fetchTask(id: string) {
  return apiRequest<{ item: TaskItem }>(`/tasks/${id}`, { token: token() })
}

export function createTask(payload: TaskPayload) {
  return apiRequest<{ item: TaskItem }>('/tasks', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateTask(id: string, payload: TaskPayload) {
  return apiRequest<{ item: TaskItem }>(`/tasks/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function deleteTask(id: string) {
  return apiRequest<void>(`/tasks/${id}`, {
    method: 'DELETE',
    token: token()
  })
}

export const assignTask = (id: string, payload: { userId: string; role?: string }) =>
  apiRequest<{ item: TaskItem }>(`/tasks/${id}/assignments`, { method: 'POST', token: token(), body: JSON.stringify(payload) })
export const fetchTaskComments = (id: string) =>
  apiRequest<{ items: TaskComment[] }>(`/tasks/${id}/comments`, { token: token() })
export const addTaskComment = (id: string, content: string) =>
  apiRequest<{ item: TaskComment }>(`/tasks/${id}/comments`, { method: 'POST', token: token(), body: JSON.stringify({ content }) })
export const fetchTaskAttachments = (id: string) =>
  apiRequest<{ items: TaskAttachment[] }>(`/tasks/${id}/attachments`, { token: token() })
export const addTaskAttachment = (id: string, payload: { fileName: string; fileUrl: string }) =>
  apiRequest<{ item: TaskAttachment }>(`/tasks/${id}/attachments`, { method: 'POST', token: token(), body: JSON.stringify(payload) })
export const fetchTaskStatusHistory = (id: string) =>
  apiRequest<{ items: TaskStatusHistory[] }>(`/tasks/${id}/status-history`, { token: token() })
export const fetchTaskTimeEntries = (id: string) =>
  apiRequest<{ items: TaskTimeEntry[] }>(`/tasks/${id}/time-entries`, { token: token() })
export const addTaskTimeEntry = (id: string, payload: { hours: number; note?: string }) =>
  apiRequest<{ item: TaskTimeEntry }>(`/tasks/${id}/time-entries`, { method: 'POST', token: token(), body: JSON.stringify(payload) })
export const approveTask = (id: string) =>
  apiRequest<{ item: TaskItem }>(`/tasks/${id}/approve`, { method: 'POST', token: token() })
