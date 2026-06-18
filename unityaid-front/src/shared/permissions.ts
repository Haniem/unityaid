import type { User } from '../entities/auth/types'

const MANAGER_ROLES = new Set<string>(['super_admin', 'org_admin', 'coordinator'])
const CLIENT_ADMIN_ROLES = new Set<string>(['super_admin', 'org_admin'])

export function hasSystemAdminRole(user?: User | null) {
  return user?.systemRoles?.some((role) => role.code === 'system_admin') ?? false
}

export function canManageContent(user?: User | null) {
  if (!user) return false
  if (hasSystemAdminRole(user)) return true
  return MANAGER_ROLES.has(user.primaryRole)
}

export function canManageOrganizations(user?: User | null) {
  if (!user) return false
  if (hasSystemAdminRole(user)) return true
  return CLIENT_ADMIN_ROLES.has(user.primaryRole)
}
