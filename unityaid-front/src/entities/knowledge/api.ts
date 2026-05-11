import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { KnowledgeArticle, KnowledgeCategory, KnowledgePayload } from './types'

const token = () => authState.token

function query(params: { search?: string; status?: string; categoryId?: string }) {
  const search = new URLSearchParams()
  if (params.search) search.set('search', params.search)
  if (params.status) search.set('status', params.status)
  if (params.categoryId) search.set('categoryId', params.categoryId)
  return search.toString() ? `?${search}` : ''
}

export const fetchKnowledgeArticles = (params: { search?: string; status?: string; categoryId?: string } = {}) =>
  apiRequest<{ items: KnowledgeArticle[] }>(`/knowledge-base${query(params)}`, { token: token() })

export const fetchKnowledgeArticle = (id: string) =>
  apiRequest<{ item: KnowledgeArticle }>(`/knowledge-base/${id}`, { token: token() })

export const createKnowledgeArticle = (payload: KnowledgePayload) =>
  apiRequest<{ item: KnowledgeArticle }>('/knowledge-base', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })

export const updateKnowledgeArticle = (id: string, payload: KnowledgePayload) =>
  apiRequest<{ item: KnowledgeArticle }>(`/knowledge-base/${id}`, {
    method: 'PUT',
    token: token(),
    body: JSON.stringify(payload)
  })

export const deleteKnowledgeArticle = (id: string) =>
  apiRequest<void>(`/knowledge-base/${id}`, {
    method: 'DELETE',
    token: token()
  })

export const fetchKnowledgeCategories = () =>
  apiRequest<{ items: KnowledgeCategory[] }>('/knowledge-base/categories', { token: token() })

export const createKnowledgeCategory = (name: string, description = '') =>
  apiRequest<{ item: KnowledgeCategory }>('/knowledge-base/categories', {
    method: 'POST',
    token: token(),
    body: JSON.stringify({ name, description })
  })
