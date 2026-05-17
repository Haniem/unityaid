export type PendingInvite = {
  email: string
  role: 'org_admin' | 'coordinator' | 'volunteer'
}

export type TenantSettings = {
  id: string
  displayName: string
  description: string
  logoUrl?: string | null
  primaryColor: string
  accentColor: string
  timezone: string
  locale: 'ru' | 'en'
  contactEmail?: string | null
  contactPhone?: string | null
  defaultOrganizationId?: string | null
  defaultOrganizationName: string
  defaultOrganizationSlug: string
  onboardingCompleted: boolean
  pendingInvites: PendingInvite[]
  createdAt: string
  updatedAt: string
}

export type TenantSettingsPayload = {
  displayName: string
  description: string
  logoUrl?: string | null
  primaryColor: string
  accentColor: string
  timezone: string
  locale: 'ru' | 'en'
  contactEmail?: string | null
  contactPhone?: string | null
  defaultOrganizationId?: string | null
  defaultOrganizationName: string
  defaultOrganizationSlug: string
  onboardingCompleted: boolean
  pendingInvites: PendingInvite[]
}
