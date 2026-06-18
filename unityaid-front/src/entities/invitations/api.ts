import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { Invitation, InvitationPayload, InvitationRegistrationPayload } from './types'

const token = () => authState.token

export function fetchInvitations() {
  return apiRequest<{ items: Invitation[] }>('/invitations', { token: token() })
}

export function createInvitation(payload: InvitationPayload) {
  return apiRequest<{ item: Invitation }>('/invitations', {
    method: 'POST',
    token: token(),
    body: JSON.stringify(payload)
  })
}

export function acceptInvitation(inviteToken: string) {
  return apiRequest<{ item: Invitation }>(`/invitations/${inviteToken}/accept`, {
    method: 'POST',
    token: token()
  })
}

export function fetchInvitation(inviteToken: string) {
  return apiRequest<{ item: Invitation }>(`/invitations/${inviteToken}`)
}

export function registerByInvitation(inviteToken: string, payload: InvitationRegistrationPayload) {
  return apiRequest<{ item: Invitation }>(`/invitations/${inviteToken}/register`, {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}
