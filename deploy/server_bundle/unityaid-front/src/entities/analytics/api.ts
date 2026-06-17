import { API_BASE_URL, ApiError, apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { AuditReport, EventsReport, GamificationReport, ManagementReport, OverviewReport, TasksReport, VolunteersReport } from './types'

const token = () => authState.token

function query(params: { from?: string; to?: string }) {
  const search = new URLSearchParams()
  if (params.from) search.set('from', params.from)
  if (params.to) search.set('to', params.to)
  return search.toString() ? `?${search}` : ''
}

export const fetchAnalyticsOverview = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: OverviewReport }>(`/analytics/overview${query(params)}`, { token: token() })

export const fetchAnalyticsVolunteers = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: VolunteersReport }>(`/analytics/volunteers${query(params)}`, { token: token() })

export const fetchAnalyticsEvents = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: EventsReport }>(`/analytics/events${query(params)}`, { token: token() })

export const fetchAnalyticsTasks = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: TasksReport }>(`/analytics/tasks${query(params)}`, { token: token() })

export const fetchAnalyticsGamification = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: GamificationReport }>(`/analytics/gamification${query(params)}`, { token: token() })

export const fetchAnalyticsAudit = (params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: AuditReport }>(`/analytics/audit${query(params)}`, { token: token() })

export const fetchManagementReport = (kind: string, params: { from?: string; to?: string } = {}) =>
  apiRequest<{ item: ManagementReport }>(`/analytics/management/${kind}${query(params)}`, { token: token() })

export async function downloadManagementReport(kind: string, format: 'xlsx' | 'pdf', params: { from?: string; to?: string } = {}) {
  const response = await fetch(`${API_BASE_URL}/analytics/management/${kind}/export/${format}${query(params)}`, {
    headers: {
      Authorization: `Bearer ${authState.token ?? ''}`
    }
  })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    throw new ApiError(payload?.message ?? 'Не удалось выгрузить отчет', response.status)
  }
  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `puls-${kind}-report.${format}`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
