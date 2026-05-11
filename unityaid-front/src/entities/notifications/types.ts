export type NotificationItem = {
  id: string
  userId: string
  type: string
  title: string
  body: string
  link: string
  entityType: string
  entityId: string
  isRead: boolean
  readAt: string | null
  createdAt: string
}

export type NotificationsResponse = {
  items: NotificationItem[]
  unreadCount: number
}
