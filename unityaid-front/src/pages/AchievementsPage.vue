<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Award, Medal, RefreshCw, Star, Trophy } from 'lucide-vue-next'
import {
  fetchAchievements,
  fetchGamificationProfile,
  fetchLeaderboard,
  recalculateAchievements
} from '../entities/gamification/api'
import type { Achievement, GamificationProfile, LeaderboardEntry } from '../entities/gamification/types'
import { authState } from '../entities/auth/store'

const profile = ref<GamificationProfile | null>(null)
const achievements = ref<Achievement[]>([])
const leaderboard = ref<LeaderboardEntry[]>([])
const errorMessage = ref('')
const isRecalculating = ref(false)

const earnedCodes = computed(() => new Set(profile.value?.achievements.map((item) => item.code) ?? []))
const canRecalculate = computed(() => {
  const systemAdmin = authState.user?.systemRoles?.some((role) => role.code === 'system_admin') ?? false
  const manager = authState.user?.organizations.some((item) => ['super_admin', 'org_admin', 'coordinator'].includes(item.role)) ?? false
  return systemAdmin || authState.user?.primaryRole === 'super_admin' || manager
})

async function load() {
  errorMessage.value = ''
  try {
    const [profileResponse, achievementsResponse, leaderboardResponse] = await Promise.all([
      fetchGamificationProfile(),
      fetchAchievements(),
      fetchLeaderboard()
    ])
    profile.value = profileResponse.item
    achievements.value = achievementsResponse.items
    leaderboard.value = leaderboardResponse.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить достижения'
  }
}

async function recalculate() {
  isRecalculating.value = true
  try {
    await recalculateAchievements()
    await load()
  } finally {
    isRecalculating.value = false
  }
}

function progressPercent() {
  return `${Math.min(100, Math.max(0, (profile.value?.levelProgress ?? 0) * 100)).toFixed(0)}%`
}

function initials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('') || 'П'
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Геймификация</p>
        <h1>Достижения</h1>
      </div>
      <button v-if="canRecalculate" class="secondary-action" type="button" :disabled="isRecalculating" @click="recalculate">
        <RefreshCw :size="17" />
        <span>{{ isRecalculating ? 'Пересчет...' : 'Пересчитать' }}</span>
      </button>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <template v-if="profile">
      <div class="metric-grid achievement-metrics">
        <article class="metric-card">
          <Trophy :size="22" />
          <span>Уровень</span>
          <strong>{{ profile.level }}</strong>
        </article>
        <article class="metric-card">
          <Star :size="22" />
          <span>Баллы</span>
          <strong>{{ profile.points }}</strong>
        </article>
        <article class="metric-card">
          <Award :size="22" />
          <span>Достижения</span>
          <strong>{{ profile.achievements.length }}</strong>
        </article>
        <article class="metric-card">
          <Medal :size="22" />
          <span>Часы</span>
          <strong>{{ profile.totalHours.toFixed(1) }}</strong>
        </article>
      </div>

      <div class="achievement-layout">
        <section class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Прогресс</p>
              <h2>До следующего уровня</h2>
            </div>
            <strong>{{ profile.points }} / {{ profile.nextLevelAt }}</strong>
          </div>
          <div class="level-progress">
            <span :style="{ width: progressPercent() }"></span>
          </div>
          <p class="achievement-hint">
            Следующая цель: {{ profile.nextAchievement?.name || 'все базовые достижения получены' }}
          </p>
        </section>

        <section class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Рейтинг</p>
              <h2>Волонтеры</h2>
            </div>
          </div>
          <article v-for="entry in leaderboard" :key="entry.userId" class="leaderboard-row">
            <strong>{{ entry.rank }}</strong>
            <img v-if="entry.avatarUrl" :src="entry.avatarUrl" alt="" />
            <span v-else class="mini-avatar">{{ initials(entry.userName) }}</span>
            <div>
              <span>{{ entry.userName }}</span>
              <small>{{ entry.totalHours.toFixed(1) }} ч. · {{ entry.level }} уровень</small>
            </div>
            <b>{{ entry.points }}</b>
          </article>
        </section>
      </div>

      <section class="detail-panel achievement-list-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Коллекция</p>
            <h2>Базовые достижения</h2>
          </div>
        </div>
        <div class="achievement-grid">
          <article v-for="achievement in achievements" :key="achievement.id" :class="['achievement-card', { earned: earnedCodes.has(achievement.code) }]">
            <Award :size="22" />
            <div>
              <strong>{{ achievement.name }}</strong>
              <p>{{ achievement.description }}</p>
              <small>{{ achievement.pointsReward }} баллов</small>
            </div>
          </article>
        </div>
      </section>
    </template>
  </section>
</template>
