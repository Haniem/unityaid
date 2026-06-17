import { authState } from '../auth/store'
import { apiRequest } from '../../shared/api'
import type { NewsCategory, NewsItem, NewsPayload } from './types'

function token() {
  return authState.token
}

export function fetchNewsList(params: { search?: string; status?: string; categoryId?: string; organizationId?: string } = {}) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.status) query.set('status', params.status)
  if (params.categoryId) query.set('categoryId', params.categoryId)
  if (params.organizationId) query.set('organizationId', params.organizationId)
  return apiRequest<{ items: NewsItem[] }>(`/news${query.toString() ? `?${query}` : ''}`, { token: token() })
}

export function fetchNewsItem(id: string) {
  return apiRequest<{ item: NewsItem }>(`/news/${id}`, { token: token() })
}

export function createNews(payload: NewsPayload) {
  return apiRequest<{ item: NewsItem }>('/news', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function updateNews(id: string, payload: NewsPayload) {
  return apiRequest<{ item: NewsItem }>(`/news/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function deleteNews(id: string) {
  return apiRequest<void>(`/news/${id}`, {
    method: 'DELETE',
    token: token()
  })
}

export function uploadNewsImage(file: File) {
  const body = new FormData()
  body.append('file', file)

  return apiRequest<{ url: string }>('/files/news-images', {
    method: 'POST',
    token: token(),
    body
  })
}

export const fetchNewsCategories = () =>
  apiRequest<{ items: NewsCategory[] }>('/news/categories', { token: token() })

export const createNewsCategory = (name: string) =>
  apiRequest<{ item: NewsCategory }>('/news/categories', { method: 'POST', token: token(), body: JSON.stringify({ name }) })

export const cleanupNewsFiles = () =>
  apiRequest<{ removed: number }>('/news/cleanup-files', { method: 'POST', token: token() })
