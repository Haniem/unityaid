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
  isEmailVerified: boolean
  isActive: boolean
  primaryRole: Membership['role']
  organizationId?: string | null
  organizations: Membership[]
  systemRoles?: SystemRole[]
}

export type SystemRole = {
  id: string
  code: 'system_admin' | 'system_manager' | 'support' | 'user'
  name: string
  description: string
  createdAt?: string
}

export type LoginResponse = {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
  user: User
}

export type DevTokenResponse = {
  status: string
  token?: string
}
