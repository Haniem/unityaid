export type EventItem = {
  id: string
  organizationId: string
  organizationName: string
  title: string
  description: string
  format: 'online' | 'offline' | 'hybrid'
  status: 'draft' | 'published' | 'completed' | 'cancelled'
  startsAt: string
  endsAt: string
  location?: string | null
  maxParticipants?: number | null
  createdAt: string
  updatedAt: string
}

export type EventPayload = {
  organizationId: string
  title: string
  description: string
  format: EventItem['format']
  status: EventItem['status']
  startsAt: string
  endsAt: string
  location?: string | null
  maxParticipants?: number | null
}
