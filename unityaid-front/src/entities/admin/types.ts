export type AdminEntityConfig = {
  code: string
  label: string
  primaryKey: string
  columns: string[]
  editable: string[]
  search: string[]
  canCreate: boolean
  canDelete: boolean
  description: string
}

export type AdminEntityRow = Record<string, unknown>

export type AdminListResponse = {
  entity: AdminEntityConfig
  items: AdminEntityRow[]
  page: number
  perPage: number
  total: number
}

export type AdminEntitiesResponse = {
  items: AdminEntityConfig[]
}
