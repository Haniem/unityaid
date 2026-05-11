export type Metric = {
  code: string
  label: string
  value: number
}

export type ChartPoint = {
  label: string
  value: number
}

export type TopVolunteer = {
  userId: string
  userName: string
  email: string
  totalHours: number
  points: number
  level: number
}

export type OverviewReport = {
  metrics: Metric[]
  applications: ChartPoint[]
  attendance: ChartPoint[]
  topVolunteers: TopVolunteer[]
}

export type VolunteersReport = {
  metrics: Metric[]
  status: ChartPoint[]
  hoursByLevel: ChartPoint[]
  topVolunteers: TopVolunteer[]
}

export type EventsReport = {
  metrics: Metric[]
  byStatus: ChartPoint[]
  applications: ChartPoint[]
  attendance: ChartPoint[]
}

export type TasksReport = {
  metrics: Metric[]
  byStatus: ChartPoint[]
  byPriority: ChartPoint[]
  completed: ChartPoint[]
}

export type GamificationTransaction = {
  id: string
  userId: string
  userName: string
  email: string
  achievementName: string
  sourceType: string
  sourceId: string
  points: number
  reason: string
  createdAt: string
}

export type GamificationReport = {
  metrics: Metric[]
  byReason: ChartPoint[]
  pointsByDay: ChartPoint[]
  transactions: GamificationTransaction[]
}

export type AuditEntry = {
  id: string
  userId: string
  userName: string
  email: string
  method: string
  action: string
  entityType: string
  entityId: string
  path: string
  statusCode: number
  createdAt: string
}

export type AuditReport = {
  metrics: Metric[]
  byAction: ChartPoint[]
  byEntity: ChartPoint[]
  activity: ChartPoint[]
  entries: AuditEntry[]
}

export type AnalyticsReport =
  | OverviewReport
  | VolunteersReport
  | EventsReport
  | TasksReport
  | GamificationReport
  | AuditReport
