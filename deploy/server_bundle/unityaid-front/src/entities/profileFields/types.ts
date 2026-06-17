export type ProfileFieldOption = {
  id?: string
  value: string
  label: string
  sortOrder?: number
  isActive?: boolean
}

export type ProfileField = {
  id: string
  organizationId: string
  groupId: string
  code: string
  name: string
  type: 'text' | 'textarea' | 'number' | 'date' | 'datetime' | 'tel' | 'email' | 'url' | 'select' | 'multiselect' | 'checkbox' | 'file'
  required: boolean
  isSystem: boolean
  isActive: boolean
  editableByUser: boolean
  sortOrder: number
  placeholder: string
  help: string
  options: ProfileFieldOption[]
}

export type ProfileFieldGroup = {
  id: string
  organizationId: string
  code: string
  name: string
  description: string
  sortOrder: number
  isSystem: boolean
  isActive: boolean
  fields: ProfileField[]
}

export type ProfileGroupPayload = {
  organizationId: string
  name: string
  description: string
  sortOrder: number
  isActive: boolean
}

export type ProfileFieldPayload = {
  organizationId: string
  groupId: string
  name: string
  type: ProfileField['type']
  required: boolean
  editableByUser: boolean
  isActive: boolean
  sortOrder: number
  placeholder: string
  help: string
  options: ProfileFieldOption[]
}

export type ProfileValue = string | number | boolean | string[] | null
