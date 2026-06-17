import type { BackendForm, FormModel, FormValue } from '../entities/forms/types'

export function modelFromForm(form: BackendForm): FormModel {
  return Object.fromEntries(form.fields.map((field) => [field.code, normalizeInitialValue(field.value, field.multi)]))
}

export function nullable(value: FormValue) {
  if (typeof value !== 'string') return value
  return value.trim() || null
}

export function stringValue(value: FormValue) {
  return typeof value === 'string' ? value : ''
}

export function numberOrNull(value: FormValue) {
  if (value === null || value === '') return null
  return Number(value)
}

export function datetimeLocalValue(value: FormValue) {
  if (!value || typeof value !== 'string') return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const offset = date.getTimezoneOffset()
  const local = new Date(date.getTime() - offset * 60_000)
  return local.toISOString().slice(0, 16)
}

export function isoFromDatetimeLocal(value: FormValue) {
  if (!value || typeof value !== 'string') return null
  return new Date(value).toISOString()
}

function normalizeInitialValue(value: FormValue, multi?: boolean): FormValue {
  if (multi) return Array.isArray(value) ? value : []
  if (typeof value === 'string') return value
  return value ?? ''
}
