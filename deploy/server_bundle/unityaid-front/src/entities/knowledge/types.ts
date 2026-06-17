export type KnowledgeArticle = {
  id: string
  categoryId?: string | null
  categoryName?: string | null
  title: string
  slug: string
  summary: string
  contentHtml: string
  status: 'draft' | 'published' | 'archived'
  authorId?: string | null
  authorName?: string | null
  publishedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type KnowledgeCategory = {
  id: string
  name: string
  slug: string
  description: string
}

export type KnowledgePayload = {
  categoryId?: string | null
  title: string
  summary: string
  contentHtml: string
  status: 'draft' | 'published' | 'archived'
}
