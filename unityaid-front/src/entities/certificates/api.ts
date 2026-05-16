import { API_BASE_URL, apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { Certificate, CertificateGeneratePayload, CertificateListResponse } from './types'

const token = () => authState.token

export function fetchCertificates() {
  return apiRequest<CertificateListResponse>('/certificates', { token: token() })
}

export function generateCertificate(payload: CertificateGeneratePayload) {
  return apiRequest<{ item: Certificate }>('/certificates/generate', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function verifyCertificate(code: string) {
  return apiRequest<{ item: Certificate }>(`/certificates/verify/${encodeURIComponent(code)}`)
}

export async function downloadCertificate(id: string) {
  const response = await fetch(`${API_BASE_URL}/certificates/download/${id}`, {
    headers: {
      Authorization: `Bearer ${token() ?? ''}`
    }
  })
  if (!response.ok) {
    throw new Error('Не удалось скачать сертификат')
  }
  return response.blob()
}
