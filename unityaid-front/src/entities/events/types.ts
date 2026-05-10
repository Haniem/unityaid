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
  checkinCode: string
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

export type EventApplication = {
  id: string
  eventId: string
  userId: string
  userName: string
  email: string
  status: 'pending' | 'approved' | 'waitlisted' | 'rejected' | 'cancelled'
  message: string
  createdAt: string
  updatedAt: string
}

export type EventAttendance = {
  id: string
  eventId: string
  userId: string
  userName: string
  email: string
  checkInAt?: string | null
  checkOutAt?: string | null
  hours: number
}

export type EventShift = {
  id: string
  eventId: string
  title: string
  startsAt: string
  endsAt: string
  capacity?: number | null
}

export type EventFeedback = {
  id: string
  userName: string
  rating: number
  comment: string
  createdAt: string
}
