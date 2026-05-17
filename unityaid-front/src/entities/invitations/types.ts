export type Invitation = {
  id: string
  email: string
  role: string
  token: string
  link: string
  status: 'pending' | 'accepted' | 'revoked' | 'expired'
  invitedBy?: string | null
  acceptedBy?: string | null
  expiresAt: string
  acceptedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type InvitationPayload = {
  email: string
  role: string
}
