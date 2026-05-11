import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { AuditReport, EventsReport, GamificationReport, OverviewReport, TasksReport, VolunteersReport } from './types'

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
