export type Organization = {
  id: string
  name: string
  slug: string
  description: string
  contactEmail?: string | null
  logoUrl?: string | null
  websiteUrl?: string | null
  phone?: string | null
  address?: string | null
  isDeleted: boolean
  createdAt: string
  updatedAt: string
}

export type OrganizationMember = {
  id: string
  userId: string
  email: string
  firstName: string
  lastName: string
  avatarUrl?: string | null
  role: 'super_admin' | 'org_admin' | 'coordinator' | 'volunteer'
  status: 'active' | 'inactive' | 'blocked'
  createdAt: string
  updatedAt: string
}

export type OrganizationPayload = {
  name: string
  slug: string
  description: string
  contactEmail?: string | null
  logoUrl?: string | null
  websiteUrl?: string | null
  phone?: string | null
  address?: string | null
}

export type AddOrganizationMemberPayload = {
  email: string
  role: OrganizationMember['role']
  status?: OrganizationMember['status']
}

export type UpdateOrganizationMemberPayload = {
  role: OrganizationMember['role']
  status: OrganizationMember['status']
}
