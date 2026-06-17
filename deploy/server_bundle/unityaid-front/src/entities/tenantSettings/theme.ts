import { reactive } from 'vue'
import { fetchTenantSettings } from './api'
import type { TenantSettings } from './types'

const DEFAULT_PRIMARY = '#2f9f72'
const DEFAULT_ACCENT = '#227a58'

export const tenantTheme = reactive({
  displayName: 'Пульс',
  logoUrl: null as string | null,
  primaryColor: DEFAULT_PRIMARY,
  accentColor: DEFAULT_ACCENT,
  isLoaded: false
})

function normalizeHex(value: string | null | undefined, fallback: string) {
  if (!value) return fallback
  const trimmed = value.trim()
  return /^#[0-9a-f]{6}$/i.test(trimmed) ? trimmed : fallback
}

function hexToRgb(hex: string) {
  const value = hex.replace('#', '')
  return {
    r: Number.parseInt(value.slice(0, 2), 16),
    g: Number.parseInt(value.slice(2, 4), 16),
    b: Number.parseInt(value.slice(4, 6), 16)
  }
}

function mix(hex: string, target: '#ffffff' | '#000000', amount: number) {
  const source = hexToRgb(hex)
  const targetRgb = target === '#ffffff' ? { r: 255, g: 255, b: 255 } : { r: 0, g: 0, b: 0 }
  const channel = (from: number, to: number) => Math.round(from + (to - from) * amount).toString(16).padStart(2, '0')
  return `#${channel(source.r, targetRgb.r)}${channel(source.g, targetRgb.g)}${channel(source.b, targetRgb.b)}`
}

export function applyTenantTheme(settings: Pick<TenantSettings, 'displayName' | 'logoUrl' | 'primaryColor' | 'accentColor'>) {
  const primaryColor = normalizeHex(settings.primaryColor, DEFAULT_PRIMARY)
  const accentColor = normalizeHex(settings.accentColor, DEFAULT_ACCENT)

  tenantTheme.displayName = settings.displayName || 'Пульс'
  tenantTheme.logoUrl = settings.logoUrl || null
  tenantTheme.primaryColor = primaryColor
  tenantTheme.accentColor = accentColor
  tenantTheme.isLoaded = true

  const root = document.documentElement
  root.style.setProperty('--accent', primaryColor)
  root.style.setProperty('--accent-strong', accentColor)
  root.style.setProperty('--accent-soft', mix(primaryColor, '#ffffff', 0.88))
  root.style.setProperty('--accent-border', mix(primaryColor, '#ffffff', 0.68))
}

export async function loadTenantTheme() {
  const response = await fetchTenantSettings()
  applyTenantTheme(response.item)
  return response.item
}
