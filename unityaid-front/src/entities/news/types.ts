export type NewsItem = {
  id: string
  organizationId?: string | null
  organizationName?: string | null
  title: string
  slug: string
  summary: string
  contentHtml: string
  coverImageUrl?: string | null
  status: 'draft' | 'published'
  authorId?: string | null
  authorName?: string | null
  publishedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type NewsPayload = {
  title: string
  summary: string
  contentHtml: string
  coverImageUrl?: string | null
  status: 'draft' | 'published'
}
