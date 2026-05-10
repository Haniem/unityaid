import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { TaskItem, TaskPayload } from './types'

const token = () => authState.token

export function fetchTasks() {
  return apiRequest<{ items: TaskItem[] }>('/tasks', { token: token() })
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
