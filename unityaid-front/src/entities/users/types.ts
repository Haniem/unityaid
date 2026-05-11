export type Membership = {
  id: string
  organizationId: string
  organizationName: string
  role: 'super_admin' | 'org_admin' | 'coordinator' | 'volunteer'
  status: string
}

export type Skill = {
  id: string
  name: string
  createdAt: string
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
  lastLoginAt?: string | null
  organizations: Membership[]
  createdAt: string
  updatedAt: string
}

export type VolunteerProfile = {
  id: string
  userId: string
  email: string
  firstName: string
  lastName: string
  patronymic?: string | null
  avatarUrl?: string | null
  city?: string | null
  phone?: string | null
  bio: string
  totalHours: number
  points: number
  level: number
  skills: Skill[]
  organizations: Membership[]
  createdAt: string
  updatedAt: string
}

export type VolunteerPayload = {
  firstName: string
  lastName: string
  patronymic?: string | null
  avatarUrl?: string | null
  city?: string | null
  phone?: string | null
  bio: string
  skillIds: string[]
}

export type SkillPayload = {
  name: string
}
