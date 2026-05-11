import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { BackendForm } from './types'

const token = () => authState.token

export function fetchCreateForm(entity: string) {
  return apiRequest<BackendForm>(`/forms/${entity}/create`, { token: token() })
}

export function fetchEditForm(entity: string, id: string) {
  return apiRequest<BackendForm>(`/forms/${entity}/${id}/edit`, { token: token() })
}

export function uploadFormFile(endpoint: string, file: File) {
  const body = new FormData()
  body.append('file', file)
  return apiRequest<{ url: string }>(endpoint, {
    method: 'POST',
    token: token(),
    body
  })
}
