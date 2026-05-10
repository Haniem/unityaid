import { authState } from '../auth/store'
import { apiRequest } from '../../shared/api'
import type { NewsItem, NewsPayload } from './types'

function token() {
  return authState.token
}

export function fetchNewsList() {
  return apiRequest<{ items: NewsItem[] }>('/news', { token: token() })
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
