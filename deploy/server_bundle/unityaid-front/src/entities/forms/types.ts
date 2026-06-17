export type FormValue = string | number | boolean | string[] | null

export type PossibleValue = {
  id: string
  name: string
  avatar?: string | null
}

export type FormField = {
  name?: string
  code: string
  type?: 'text' | 'email' | 'tel' | 'url' | 'number' | 'textarea' | 'select' | 'datetime' | 'file' | 'checkbox'
  require: boolean
  disabled?: boolean
  value: FormValue
  maxLength?: number
  min?: number
  rows?: number
  placeholder?: string
  help?: string
  endpoint?: string
  uploadEndpoint?: string
  accept?: string
  multi?: boolean
  possibleValues?: PossibleValue[]
}

export type BackendForm = {
  id?: string | null
  lang: string
  meta: {
    title: string
    description?: string
  }
  fields: FormField[]
}

export type FormModel = Record<string, FormValue>
