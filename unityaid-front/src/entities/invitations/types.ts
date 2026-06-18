export type Invitation = {
  id: string
  email?: string | null
  role: string
  token: string
  link: string
  status: 'pending' | 'accepted' | 'revoked' | 'expired'
  organizationId?: string | null
  organizationName?: string | null
  invitedBy?: string | null
  acceptedBy?: string | null
  expiresAt: string
  acceptedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type InvitationPayload = {
  email?: string
  role: string
  organizationId?: string
}

export type InvitationRegistrationPayload = {
  email: string
  password: string
  firstName: string
  lastName: string
  patronymic?: string
}
