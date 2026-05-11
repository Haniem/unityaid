export type TimeEntryStatus = 'pending' | 'approved' | 'rejected'

export type TimeEntry = {
  id: string
  organizationId: string
  organizationName: string
  userId: string
  userName: string
  eventId?: string | null
  eventTitle?: string | null
  taskId?: string | null
  taskTitle?: string | null
  hours: number
  description: string
  status: TimeEntryStatus
  reviewedBy?: string | null
  reviewedByName?: string | null
  reviewedAt?: string | null
  createdAt: string
  updatedAt: string
}

export type TimeEntryPayload = {
  eventId?: string | null
  taskId?: string | null
  hours: number
  description?: string
}
