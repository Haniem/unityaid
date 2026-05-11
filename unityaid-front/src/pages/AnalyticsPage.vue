<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Award, BarChart3, CalendarDays, CheckSquare, ClipboardList, History, UsersRound } from 'lucide-vue-next'
import {
  fetchAnalyticsAudit,
  fetchAnalyticsEvents,
  fetchAnalyticsGamification,
  fetchAnalyticsOverview,
  fetchAnalyticsTasks,
  fetchAnalyticsVolunteers
} from '../entities/analytics/api'
import type {
  AnalyticsReport,
  AuditEntry,
  ChartPoint,
  GamificationTransaction,
  Metric,
  TopVolunteer
} from '../entities/analytics/types'

const route = useRoute()
const report = ref<AnalyticsReport | null>(null)
const errorMessage = ref('')
const isLoading = ref(false)
const from = ref('')
const to = ref('')

const reports = [
  { code: 'overview', title: 'Обзор', description: 'Ключевые показатели, заявки, посещаемость и топ волонтеров.', icon: BarChart3, to: '/analytics/overview' },
  { code: 'volunteers', title: 'Волонтеры', description: 'Статусы, уровни, часы и лидеры волонтерской активности.', icon: UsersRound, to: '/analytics/volunteers' },
  { code: 'events', title: 'Мероприятия', description: 'Статусы мероприятий, заявки и фактическая посещаемость.', icon: CalendarDays, to: '/analytics/events' },
  { code: 'tasks', title: 'Задачи', description: 'Выполнение задач, статусы и распределение по приоритетам.', icon: CheckSquare, to: '/analytics/tasks' },
  { code: 'gamification', title: 'Начисления', description: 'Все начисления баллов: кому, сколько и за какое действие.', icon: Award, to: '/analytics/gamification' },
  { code: 'audit', title: 'Действия сотрудников', description: 'Журнал создания, изменения, удаления и служебных операций.', icon: History, to: '/analytics/audit' }
]

const currentReportCode = computed(() => String(route.params.report || ''))
const currentReport = computed(() => reports.find((item) => item.code === currentReportCode.value))
const isCatalog = computed(() => !currentReport.value)
const metrics = computed<Metric[]>(() => report.value?.metrics ?? [])
const topVolunteers = computed<TopVolunteer[]>(() => ('topVolunteers' in (report.value ?? {}) ? (report.value as any).topVolunteers : []))
const transactions = computed<GamificationTransaction[]>(() => ('transactions' in (report.value ?? {}) ? (report.value as any).transactions : []))
const auditEntries = computed<AuditEntry[]>(() => ('entries' in (report.value ?? {}) ? (report.value as any).entries : []))

const chartSections = computed(() => {
  if (!report.value) return []
  const item = report.value as any
  if (currentReportCode.value === 'overview') {
    return [
      { title: 'Заявки по дням', items: item.applications as ChartPoint[] },
      { title: 'Посещения по дням', items: item.attendance as ChartPoint[] }
    ]
  }
  if (currentReportCode.value === 'volunteers') {
    return [
      { title: 'Статусы волонтеров', items: item.status as ChartPoint[] },
      { title: 'Часы по уровням', items: item.hoursByLevel as ChartPoint[] }
    ]
  }
  if (currentReportCode.value === 'events') {
    return [
      { title: 'Статусы мероприятий', items: item.byStatus as ChartPoint[] },
      { title: 'Заявки по дням', items: item.applications as ChartPoint[] },
      { title: 'Посещения по дням', items: item.attendance as ChartPoint[] }
    ]
  }
  if (currentReportCode.value === 'tasks') {
    return [
      { title: 'Статусы задач', items: item.byStatus as ChartPoint[] },
      { title: 'Приоритеты задач', items: item.byPriority as ChartPoint[] },
      { title: 'Выполнено по дням', items: item.completed as ChartPoint[] }
    ]
  }
  if (currentReportCode.value === 'gamification') {
    return [
      { title: 'Баллы по основаниям', items: item.byReason as ChartPoint[] },
      { title: 'Начислено по дням', items: item.pointsByDay as ChartPoint[] }
    ]
  }
  return [
    { title: 'Действия', items: item.byAction as ChartPoint[] },
    { title: 'Разделы системы', items: item.byEntity as ChartPoint[] },
    { title: 'Активность по дням', items: item.activity as ChartPoint[] }
  ]
})

async function load() {
  if (isCatalog.value) return
  isLoading.value = true
  errorMessage.value = ''
  const params = { from: from.value ? new Date(from.value).toISOString() : '', to: to.value ? new Date(to.value).toISOString() : '' }
  try {
    if (currentReportCode.value === 'overview') report.value = (await fetchAnalyticsOverview(params)).item
    if (currentReportCode.value === 'volunteers') report.value = (await fetchAnalyticsVolunteers(params)).item
    if (currentReportCode.value === 'events') report.value = (await fetchAnalyticsEvents(params)).item
    if (currentReportCode.value === 'tasks') report.value = (await fetchAnalyticsTasks(params)).item
    if (currentReportCode.value === 'gamification') report.value = (await fetchAnalyticsGamification(params)).item
    if (currentReportCode.value === 'audit') report.value = (await fetchAnalyticsAudit(params)).item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить отчет'
  } finally {
    isLoading.value = false
  }
}

function barWidth(items: ChartPoint[], value: number) {
  const max = Math.max(1, ...items.map((item) => item.value))
  return `${Math.max(4, (value / max) * 100)}%`
}

function formatDateTime(value: string) {
  if (!value) return ''
  return new Date(value).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
}

function formatAction(entry: AuditEntry) {
  const id = entry.entityId ? ` #${entry.entityId}` : ''
  return `${entry.method} ${entry.action}: ${entry.entityType}${id} · ${entry.path}`
}

watch(() => route.params.report, () => {
  report.value = null
  load()
})
onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Аналитика</p>
        <h1>{{ currentReport?.title || 'Отчеты' }}</h1>
        <p>{{ currentReport?.description || 'Выберите отчет, чтобы посмотреть управленческие показатели UnityAid.' }}</p>
      </div>
    </div>

    <div v-if="isCatalog" class="analytics-report-grid">
      <RouterLink v-for="item in reports" :key="item.code" class="analytics-report-card" :to="item.to">
        <component :is="item.icon" :size="24" />
        <div>
          <strong>{{ item.title }}</strong>
          <p>{{ item.description }}</p>
        </div>
      </RouterLink>
    </div>

    <template v-else>
      <div class="filter-bar">
        <label class="filter-select">
          <span>С даты</span>
          <input v-model="from" type="date" />
        </label>
        <label class="filter-select">
          <span>По дату</span>
          <input v-model="to" type="date" />
        </label>
        <button class="primary-action" type="button" @click="load">Применить</button>
        <RouterLink class="secondary-action" to="/analytics">К отчетам</RouterLink>
      </div>

      <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
      <div v-else-if="isLoading" class="empty-state">Загрузка отчета...</div>

      <template v-else-if="report">
        <div class="metric-grid analytics-metrics">
          <article v-for="metric in metrics" :key="metric.code" class="metric-card">
            <ClipboardList :size="22" />
            <span>{{ metric.label }}</span>
            <strong>{{ Number.isInteger(metric.value) ? metric.value : metric.value.toFixed(1) }}</strong>
          </article>
        </div>

        <div class="analytics-chart-grid">
          <section v-for="section in chartSections" :key="section.title" class="detail-panel">
            <div class="section-heading">
              <div>
                <p class="eyebrow">График</p>
                <h2>{{ section.title }}</h2>
              </div>
            </div>
            <div class="bar-chart">
              <div v-for="point in section.items" :key="point.label" class="bar-row">
                <span>{{ point.label }}</span>
                <div><i :style="{ width: barWidth(section.items, point.value) }"></i></div>
                <strong>{{ point.value }}</strong>
              </div>
              <p v-if="section.items.length === 0" class="empty-state">Нет данных за период.</p>
            </div>
          </section>
        </div>

        <section v-if="topVolunteers.length" class="detail-panel analytics-top-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Рейтинг</p>
              <h2>Топ волонтеров</h2>
            </div>
          </div>
          <article v-for="volunteer in topVolunteers" :key="volunteer.userId" class="analytics-table-row">
            <strong>{{ volunteer.userName }}</strong>
            <span>{{ volunteer.email }}</span>
            <span>{{ volunteer.totalHours.toFixed(1) }} ч.</span>
            <span>{{ volunteer.points }} баллов</span>
          </article>
        </section>

        <section v-if="transactions.length" class="detail-panel analytics-top-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Геймификация</p>
              <h2>Последние начисления</h2>
            </div>
          </div>
          <article v-for="item in transactions" :key="item.id" class="analytics-table-row analytics-wide-row">
            <strong>{{ item.userName || item.email }}</strong>
            <span>{{ item.achievementName || item.reason || item.sourceType }}</span>
            <span>{{ item.points }} баллов</span>
            <span>{{ formatDateTime(item.createdAt) }}</span>
          </article>
        </section>

        <section v-if="auditEntries.length" class="detail-panel analytics-top-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Аудит</p>
              <h2>Последние действия</h2>
            </div>
          </div>
          <article v-for="entry in auditEntries" :key="entry.id" class="analytics-table-row analytics-wide-row">
            <strong>{{ entry.userName || entry.email }}</strong>
            <span>{{ formatAction(entry) }}</span>
            <span>HTTP {{ entry.statusCode }}</span>
            <span>{{ formatDateTime(entry.createdAt) }}</span>
          </article>
        </section>
      </template>
    </template>
  </section>
</template>
