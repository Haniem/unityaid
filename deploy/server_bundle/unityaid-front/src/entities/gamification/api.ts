import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { Achievement, GamificationProfile, LeaderboardEntry } from './types'

const token = () => authState.token

export const fetchGamificationProfile = () =>
  apiRequest<{ item: GamificationProfile }>('/gamification/me', { token: token() })

export const fetchUserGamificationProfile = (userId: string) =>
  apiRequest<{ item: GamificationProfile }>(`/gamification/users/${userId}`, { token: token() })

export const fetchLeaderboard = () =>
  apiRequest<{ items: LeaderboardEntry[] }>('/gamification/leaderboard', { token: token() })

export const fetchAchievements = () =>
  apiRequest<{ items: Achievement[] }>('/achievements', { token: token() })

export const recalculateAchievements = () =>
  apiRequest<{ status: string }>('/achievements/recalculate', { method: 'POST', token: token() })
