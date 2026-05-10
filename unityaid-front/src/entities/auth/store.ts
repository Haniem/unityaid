import { reactive } from 'vue'
import { apiRequest } from '../../shared/api'
import type { LoginResponse, User } from './types'

const TOKEN_KEY = 'unityaid.accessToken'
const USER_KEY = 'unityaid.user'

type AuthState = {
  token: string | null
  user: User | null
  isReady: boolean
}

export const authState = reactive<AuthState>({
  token: localStorage.getItem(TOKEN_KEY),
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

function persistAuth(token: string, user: User) {
  authState.token = token
  authState.user = user
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export async function login(email: string, password: string) {
  const response = await apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  })

  persistAuth(response.accessToken, response.user)
  return response.user
}

export async function fetchCurrentUser() {
  if (!authState.token) {
    authState.isReady = true
    return null
  }

  try {
    const response = await apiRequest<{ user: User }>('/auth/me', {
      token: authState.token
    })
    authState.user = response.user
    localStorage.setItem(USER_KEY, JSON.stringify(response.user))
    return response.user
  } catch {
    logout()
    return null
  } finally {
    authState.isReady = true
  }
}

export function logout() {
  authState.token = null
  authState.user = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function setLocale(locale: 'ru' | 'en') {
  if (!authState.user) {
    return
  }

  authState.user.locale = locale
  localStorage.setItem(USER_KEY, JSON.stringify(authState.user))
}
