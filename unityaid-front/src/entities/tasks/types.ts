export type TaskItem = {
  id: string
  organizationId: string
  organizationName: string
  eventId?: string | null
  eventTitle?: string | null
  title: string
  description: string
  status: 'created' | 'assigned' | 'in_progress' | 'review' | 'completed' | 'cancelled'
  priority: 'low' | 'medium' | 'high'
  dueAt?: string | null
  createdAt: string
  updatedAt: string
}

export type TaskPayload = {
  organizationId: string
  eventId?: string | null
  title: string
  description: string
  status: TaskItem['status']
  priority: TaskItem['priority']
  dueAt?: string | null
}
