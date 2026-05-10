<script setup lang="ts">
import { CalendarDays, CheckCircle2, Clock3, Trophy } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'

const stats = [
  { label: 'Мероприятий', value: '3', icon: CalendarDays },
  { label: 'Задач в работе', value: '2', icon: CheckCircle2 },
  { label: 'Волонтерских часов', value: '41.5', icon: Clock3 },
  { label: 'Баллов начислено', value: '540', icon: Trophy }
]
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Главная страница</p>
        <h1>Добро пожаловать, {{ authState.user?.firstName }}</h1>
      </div>
      <p>
        Здесь будет сводка по активности организации, ближайшим мероприятиям, задачам и геймификации.
      </p>
    </div>

    <div class="metric-grid">
      <article v-for="stat in stats" :key="stat.label" class="metric-card">
        <component :is="stat.icon" :size="22" />
        <span>{{ stat.label }}</span>
        <strong>{{ stat.value }}</strong>
      </article>
    </div>

    <div class="work-band">
      <h2>Ближайшие сценарии разработки</h2>
      <p>
        Следующий слой: CRUD мероприятий, заявки на участие, задачи, учет времени и реальные виджеты
        аналитики.
      </p>
    </div>
  </section>
</template>
