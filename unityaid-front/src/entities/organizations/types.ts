export type Organization = {
  id: string
  name: string
  slug: string
  description: string
  contactEmail?: string | null
  createdAt: string
  updatedAt: string
}

export type OrganizationPayload = {
  name: string
  slug: string
  description: string
  contactEmail?: string | null
}
