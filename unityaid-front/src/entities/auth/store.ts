import { reactive } from 'vue'
import { apiRequest } from '../../shared/api'
import type { DevTokenResponse, LoginResponse, User } from './types'

const TOKEN_KEY = 'unityaid.accessToken'
const REFRESH_TOKEN_KEY = 'unityaid.refreshToken'
const USER_KEY = 'unityaid.user'

type AuthState = {
  token: string | null
  refreshToken: string | null
  user: User | null
  isReady: boolean
}

export const authState = reactive<AuthState>({
  token: localStorage.getItem(TOKEN_KEY),
  refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY),
  user: readUser(),
  isReady: false
})

function readUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as User
  } catch {
    localStorage.removeItem(USER_KEY)
    return null
  }
}

function persistAuth(response: LoginResponse) {
  authState.token = response.accessToken
  authState.refreshToken = response.refreshToken
  authState.user = response.user
  localStorage.setItem(TOKEN_KEY, response.accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, response.refreshToken)
  localStorage.setItem(USER_KEY, JSON.stringify(response.user))
}

function persistUser(user: User) {
  authState.user = user
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export async function login(email: string, password: string) {
  const response = await apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  })

  persistAuth(response)
  return response.user
}

export async function register(payload: { email: string; password: string; firstName: string; lastName: string }) {
  return apiRequest<DevTokenResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function verifyEmail(token: string) {
  return apiRequest<{ status: string }>('/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify({ token })
  })
}

export async function requestPasswordReset(email: string) {
  return apiRequest<DevTokenResponse>('/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify({ email })
  })
}

export async function resetPassword(token: string, newPassword: string) {
  return apiRequest<{ status: string }>('/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({ token, newPassword })
  })
}

export async function changePassword(currentPassword: string, newPassword: string) {
  return apiRequest<{ status: string }>('/auth/change-password', {
    method: 'POST',
    token: authState.token,
    body: JSON.stringify({ currentPassword, newPassword })
  })
}

export async function fetchCurrentUser() {
  if (!authState.token) {
    if (authState.refreshToken) {
      const refreshed = await refreshSession()
      if (refreshed) {
        authState.isReady = true
        return authState.user
      }
    }
    authState.isReady = true
    return null
  }

  try {
    const response = await apiRequest<{ user: User }>('/auth/me', {
      token: authState.token
    })
    persistUser(response.user)
    return response.user
  } catch {
    const refreshed = await refreshSession()
    if (!refreshed) {
      logout()
    }
    return null
  } finally {
    authState.isReady = true
  }
}

async function refreshSession() {
  if (!authState.refreshToken) {
    return false
  }

  try {
    const response = await apiRequest<LoginResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken: authState.refreshToken })
    })
    persistAuth(response)
    return true
  } catch {
    return false
  }
}

export async function logoutRemote() {
  const refreshToken = authState.refreshToken
  try {
    if (authState.token) {
      await apiRequest<{ status: string }>('/auth/logout', {
        method: 'POST',
        token: authState.token,
        body: JSON.stringify({ refreshToken })
      })
    }
  } finally {
    logout()
  }
}

export function logout() {
  authState.token = null
  authState.refreshToken = null
  authState.user = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function setLocale(locale: 'ru' | 'en') {
  if (!authState.user) {
    return
  }

  authState.user.locale = locale
  localStorage.setItem(USER_KEY, JSON.stringify(authState.user))
}
