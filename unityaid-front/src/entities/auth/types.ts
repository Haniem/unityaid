export type Membership = {
  organizationId: string
  organizationName: string
  role: 'super_admin' | 'org_admin' | 'coordinator' | 'volunteer'
  status: string
}

export type User = {
  id: string
  email: string
  firstName: string
  lastName: string
  patronymic?: string | null
  avatarUrl?: string | null
  locale: 'ru' | 'en'
  isActive: boolean
  primaryRole: Membership['role']
  organizationId?: string | null
  organizations: Membership[]
}

export type LoginResponse = {
  accessToken: string
  tokenType: string
  expiresIn: number
  user: User
}
