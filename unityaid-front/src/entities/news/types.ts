export type NewsItem = {
  id: string
  organizationId?: string | null
  organizationName?: string | null
  title: string
  slug: string
  summary: string
  contentHtml: string
  coverImageUrl?: string | null
  categoryId?: string | null
  categoryName?: string | null
  status: 'draft' | 'published' | 'scheduled'
  authorId?: string | null
  authorName?: string | null
  scheduledAt?: string | null
  publishedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type NewsPayload = {
  title: string
  summary: string
  contentHtml: string
  coverImageUrl?: string | null
  categoryId?: string | null
  status: 'draft' | 'published' | 'scheduled'
  scheduledAt?: string | null
}

export type NewsCategory = { id: string; name: string; slug: string }
