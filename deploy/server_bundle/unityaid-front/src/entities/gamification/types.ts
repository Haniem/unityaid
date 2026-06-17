export type Achievement = {
  id: string
  code: string
  name: string
  description: string
  icon: string
  pointsReward: number
  createdAt: string
}

export type UserAchievement = Achievement & {
  earnedAt: string
}

export type GamificationProfile = {
  userId: string
  userName: string
  email: string
  avatarUrl?: string | null
  totalHours: number
  points: number
  level: number
  nextLevelAt: number
  levelProgress: number
  achievements: UserAchievement[]
  nextAchievement?: Achievement | null
}

export type LeaderboardEntry = {
  rank: number
  userId: string
  userName: string
  email: string
  avatarUrl?: string | null
  totalHours: number
  points: number
  level: number
}
