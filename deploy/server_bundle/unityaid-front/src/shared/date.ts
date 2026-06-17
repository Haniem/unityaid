export function toDatetimeLocal(value?: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  const offset = date.getTimezoneOffset()
  const local = new Date(date.getTime() - offset * 60_000)
  return local.toISOString().slice(0, 16)
}

export function fromDatetimeLocal(value: string) {
  return new Date(value).toISOString()
}

export function formatDateTime(value?: string | null) {
  if (!value) {
    return 'Не указано'
  }
  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'medium',
    timeStyle: 'short'
  }).format(new Date(value))
}
