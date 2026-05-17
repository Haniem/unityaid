export type ImportRow = {
  row: number
  email: string
  firstName: string
  lastName: string
  city: string
  phone: string
  interests: string
  status: 'new' | 'active' | 'unavailable' | 'archived' | string
  errors: string[]
}

export type ImportPreview = {
  items: ImportRow[]
  validCount: number
  errorCount: number
}
