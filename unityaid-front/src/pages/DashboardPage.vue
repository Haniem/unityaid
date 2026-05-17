<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CalendarDays, CheckCircle2, Clock3, Trophy } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import { fetchTenantSettings } from '../entities/tenantSettings/api'

const stats = [
  { label: 'Мероприятий', value: '3', icon: CalendarDays },
  { label: 'Задач в работе', value: '2', icon: CheckCircle2 },
  { label: 'Волонтерских часов', value: '41.5', icon: Clock3 },
  { label: 'Баллов начислено', value: '540', icon: Trophy }
]

const onboardingCompleted = ref(true)

onMounted(async () => {
  if (!['super_admin', 'org_admin'].includes(authState.user?.primaryRole ?? '')) return
  try {
    const response = await fetchTenantSettings()
    onboardingCompleted.value = response.item.onboardingCompleted
  } catch {
    onboardingCompleted.value = true
  }
})
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Главная</p>
        <h1>Добро пожаловать, {{ authState.user?.firstName }}</h1>
      </div>
      <p>
        Здесь будет сводка по активности организации, ближайшим мероприятиям, задачам и геймификации.
      </p>
    </div>

    <div v-if="!onboardingCompleted" class="setup-banner">
      <div>
        <strong>Первичная настройка еще не завершена</strong>
        <span>Заполните пространство клиента, бренд, первое подразделение и приглашения сотрудников.</span>
      </div>
      <RouterLink class="primary-action" to="/onboarding">Продолжить настройку</RouterLink>
    </div>

    <div class="metric-grid">
      <article v-for="stat in stats" :key="stat.label" class="metric-card">
        <component :is="stat.icon" :size="22" />
        <span>{{ stat.label }}</span>
        <strong>{{ stat.value }}</strong>
      </article>
    </div>
  </section>
</template>
