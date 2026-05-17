import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { ImportPreview, ImportRow } from './types'

const token = () => authState.token

export function previewVolunteerImport(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return apiRequest<ImportPreview>('/volunteer-imports/preview', {
    method: 'POST',
    token: token(),
    body: formData
  })
}

export function commitVolunteerImport(items: ImportRow[]) {
  return apiRequest<{ imported: number }>('/volunteer-imports/commit', {
    method: 'POST',
    token: token(),
    body: JSON.stringify({ items })
  })
}
