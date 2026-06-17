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
  assignees: TaskAssignee[]
  confirmedBy?: string | null
  confirmedAt?: string | null
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

export type TaskAssignee = { userId: string; name: string; email: string; role: 'assignee' | 'co_assignee' }
export type TaskComment = { id: string; userName: string; content: string; createdAt: string }
export type TaskAttachment = { id: string; fileName: string; fileUrl: string; createdAt: string }
export type TaskStatusHistory = { id: string; fromStatus?: string | null; toStatus: string; userName?: string | null; createdAt: string }
export type TaskTimeEntry = { id: string; userName: string; hours: number; note: string; status: string; createdAt: string }
