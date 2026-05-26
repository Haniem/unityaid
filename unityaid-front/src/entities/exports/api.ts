import { API_BASE_URL, ApiError } from '../../shared/api'
import { authState } from '../auth/store'

export type ExportKind = 'volunteers' | 'events' | 'applications' | 'time-entries' | 'tasks' | 'certificates'

const filenames: Record<ExportKind, string> = {
  volunteers: 'puls-volunteers.csv',
  events: 'puls-events.csv',
  applications: 'puls-applications.csv',
  'time-entries': 'puls-time-entries.csv',
  tasks: 'puls-tasks.csv',
  certificates: 'puls-certificates.csv'
}

export async function downloadExport(kind: ExportKind) {
  const response = await fetch(`${API_BASE_URL}/exports/${kind}`, {
    headers: {
      Authorization: `Bearer ${authState.token ?? ''}`
    }
  })

  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    throw new ApiError(payload?.message ?? 'Не удалось выгрузить данные', response.status)
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filenames[kind]
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
