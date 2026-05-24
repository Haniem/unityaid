<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bell, CalendarDays, CheckCircle2, Clock3, FileText, Newspaper, Plus, Settings, Trophy, UsersRound } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import { fetchEvents } from '../entities/events/api'
import type { EventItem } from '../entities/events/types'
import { fetchNewsList } from '../entities/news/api'
import type { NewsItem } from '../entities/news/types'
import { fetchNotifications } from '../entities/notifications/api'
import type { NotificationItem } from '../entities/notifications/types'
import { fetchTasks } from '../entities/tasks/api'
import type { TaskItem } from '../entities/tasks/types'
import { fetchTenantSettings } from '../entities/tenantSettings/api'
import { fetchTimeEntries } from '../entities/timeentries/api'
import type { TimeEntry } from '../entities/timeentries/types'
import { formatDateTime } from '../shared/date'

const events = ref<EventItem[]>([])
const tasks = ref<TaskItem[]>([])
const news = ref<NewsItem[]>([])
const notifications = ref<NotificationItem[]>([])
const timeEntries = ref<TimeEntry[]>([])
const onboardingCompleted = ref(true)
const isLoading = ref(true)
const loadError = ref('')

const canManage = computed(() => ['super_admin', 'org_admin', 'coordinator'].includes(authState.user?.primaryRole ?? ''))
const activeTasks = computed(() => tasks.value.filter((item) => !['completed', 'cancelled'].includes(item.status)))
const upcomingEvents = computed(() =>
  events.value
    .filter((item) => new Date(item.startsAt).getTime() >= Date.now() && item.status !== 'cancelled')
    .sort((a, b) => new Date(a.startsAt).getTime() - new Date(b.startsAt).getTime())
)
const urgentTasks = computed(() =>
  [...activeTasks.value]
    .sort((a, b) => new Date(a.dueAt || a.updatedAt).getTime() - new Date(b.dueAt || b.updatedAt).getTime())
    .slice(0, 5)
)
const approvedHours = computed(() =>
  timeEntries.value.filter((item) => item.status === 'approved').reduce((sum, item) => sum + item.hours, 0)
)
const unreadNotifications = computed(() => notifications.value.filter((item) => !item.isRead).length)

const stats = computed(() => [
  { label: 'Ближайшие мероприятия', value: upcomingEvents.value.length, icon: CalendarDays, to: '/calendar' },
  { label: 'Активные задачи', value: activeTasks.value.length, icon: CheckCircle2, to: '/tasks' },
  { label: 'Подтвержденные часы', value: approvedHours.value.toFixed(1), icon: Clock3, to: '/time-entries' },
  { label: 'Новые уведомления', value: unreadNotifications.value, icon: Bell, to: '/notifications' }
])

const quickActions = computed(() => {
  const base = [
    { label: 'Найти мероприятие', to: '/calendar', icon: CalendarDays },
    { label: 'Добавить часы', to: '/time-entries', icon: Clock3 },
    { label: 'Открыть новости', to: '/news', icon: Newspaper }
  ]
  if (!canManage.value) return base
  return [
    { label: 'Создать мероприятие', to: '/calendar', icon: Plus },
    { label: 'Поставить задачу', to: '/tasks', icon: CheckCircle2 },
    { label: 'Опубликовать новость', to: '/news/new', icon: FileText },
    { label: 'Настройки', to: '/settings', icon: Settings }
  ]
})

async function load() {
  isLoading.value = true
  loadError.value = ''
  try {
    const [eventsResponse, tasksResponse, newsResponse, notificationsResponse, timeResponse] = await Promise.all([
      fetchEvents(),
      fetchTasks(),
      fetchNewsList({ status: 'published' }),
      fetchNotifications(),
      fetchTimeEntries()
    ])
    events.value = eventsResponse.items
    tasks.value = tasksResponse.items
    news.value = newsResponse.items
    notifications.value = notificationsResponse.items
    timeEntries.value = timeResponse.items

    if (['super_admin', 'org_admin'].includes(authState.user?.primaryRole ?? '')) {
      const response = await fetchTenantSettings()
      onboardingCompleted.value = response.item.onboardingCompleted
    }
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : 'Не удалось загрузить главную страницу'
  } finally {
    isLoading.value = false
  }
}

function taskStatusLabel(status: TaskItem['status']) {
  const labels: Record<TaskItem['status'], string> = {
    created: 'Создана',
    assigned: 'Назначена',
    in_progress: 'В работе',
    review: 'На проверке',
    completed: 'Завершена',
    cancelled: 'Отменена'
  }
  return labels[status] || status
}

function priorityLabel(priority: TaskItem['priority']) {
  const labels: Record<TaskItem['priority'], string> = {
    low: 'Низкий',
    medium: 'Средний',
    high: 'Высокий'
  }
  return labels[priority] || priority
}

onMounted(load)
</script>

<template>
  <section class="page-section dashboard-page">
    <div class="page-heading dashboard-heading">
      <div>
        <p class="eyebrow">Главная</p>
        <h1>Добро пожаловать, {{ authState.user?.firstName }}</h1>
        <p>Рабочая сводка по задачам, мероприятиям, новостям и уведомлениям на сегодня.</p>
      </div>
      <button class="secondary-action" type="button" @click="load">Обновить</button>
    </div>

    <div v-if="!onboardingCompleted" class="setup-banner">
      <div>
        <strong>Первичная настройка еще не завершена</strong>
        <span>Заполните пространство клиента, бренд, первое подразделение и приглашения сотрудников.</span>
      </div>
      <RouterLink class="primary-action" to="/onboarding">Продолжить настройку</RouterLink>
    </div>

    <p v-if="loadError" class="form-error">{{ loadError }}</p>
    <div v-else-if="isLoading" class="empty-state">Загрузка сводки...</div>

    <template v-else>
      <div class="metric-grid">
        <RouterLink v-for="stat in stats" :key="stat.label" class="metric-card dashboard-metric" :to="stat.to">
          <component :is="stat.icon" :size="22" />
          <span>{{ stat.label }}</span>
          <strong>{{ stat.value }}</strong>
        </RouterLink>
      </div>

      <div class="dashboard-layout">
        <section class="detail-panel dashboard-widget dashboard-wide-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">План</p>
              <h2>Ближайшие мероприятия</h2>
            </div>
            <RouterLink class="secondary-action" to="/calendar">Все</RouterLink>
          </div>
          <RouterLink v-for="event in upcomingEvents.slice(0, 4)" :key="event.id" class="dashboard-row" :to="`/calendar/${event.id}`">
            <CalendarDays :size="18" />
            <div>
              <strong>{{ event.title }}</strong>
              <small>{{ formatDateTime(event.startsAt) }} · {{ event.location || event.format }} · {{ event.organizationName }}</small>
            </div>
            <span class="status-pill">{{ event.status }}</span>
          </RouterLink>
          <p v-if="upcomingEvents.length === 0" class="empty-state">Ближайших мероприятий пока нет.</p>
        </section>

        <section class="detail-panel dashboard-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Фокус</p>
              <h2>Задачи</h2>
            </div>
            <RouterLink class="secondary-action" to="/tasks">Все</RouterLink>
          </div>
          <RouterLink v-for="task in urgentTasks" :key="task.id" class="dashboard-row compact" :to="`/tasks/${task.id}`">
            <CheckCircle2 :size="18" />
            <div>
              <strong>{{ task.title }}</strong>
              <small>{{ taskStatusLabel(task.status) }} · {{ priorityLabel(task.priority) }} · {{ task.organizationName }}</small>
            </div>
          </RouterLink>
          <p v-if="urgentTasks.length === 0" class="empty-state">Активных задач нет.</p>
        </section>

        <section class="detail-panel dashboard-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Новости</p>
              <h2>Последние публикации</h2>
            </div>
            <RouterLink class="secondary-action" to="/news">Все</RouterLink>
          </div>
          <RouterLink v-for="item in news.slice(0, 3)" :key="item.id" class="dashboard-news-row" :to="`/news/${item.id}`">
            <img v-if="item.coverImageUrl" :src="item.coverImageUrl" alt="" />
            <span v-else class="dashboard-news-placeholder"><Newspaper :size="18" /></span>
            <div>
              <strong>{{ item.title }}</strong>
              <small>{{ item.categoryName || item.organizationName || 'Новость' }}</small>
            </div>
          </RouterLink>
          <p v-if="news.length === 0" class="empty-state">Опубликованных новостей пока нет.</p>
        </section>

        <section class="detail-panel dashboard-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Быстрый старт</p>
              <h2>Действия</h2>
            </div>
          </div>
          <div class="dashboard-action-grid">
            <RouterLink v-for="action in quickActions" :key="action.to" class="dashboard-action" :to="action.to">
              <component :is="action.icon" :size="18" />
              <span>{{ action.label }}</span>
            </RouterLink>
          </div>
        </section>

        <section class="detail-panel dashboard-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Лента</p>
              <h2>Уведомления</h2>
            </div>
            <RouterLink class="secondary-action" to="/notifications">Все</RouterLink>
          </div>
          <RouterLink v-for="item in notifications.slice(0, 4)" :key="item.id" :class="['dashboard-row', 'compact', { unread: !item.isRead }]" :to="item.link || '/notifications'">
            <Bell :size="18" />
            <div>
              <strong>{{ item.title }}</strong>
              <small>{{ item.body }}</small>
            </div>
          </RouterLink>
          <p v-if="notifications.length === 0" class="empty-state">Уведомлений пока нет.</p>
        </section>

        <section class="detail-panel dashboard-widget dashboard-wide-widget">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Прогресс</p>
              <h2>Мой вклад</h2>
            </div>
            <Trophy :size="22" />
          </div>
          <div class="dashboard-contribution">
            <div>
              <UsersRound :size="20" />
              <span>В работе задач</span>
              <strong>{{ activeTasks.length }}</strong>
            </div>
            <div>
              <Clock3 :size="20" />
              <span>Часов подтверждено</span>
              <strong>{{ approvedHours.toFixed(1) }}</strong>
            </div>
            <div>
              <Newspaper :size="20" />
              <span>Новостей в ленте</span>
              <strong>{{ news.length }}</strong>
            </div>
          </div>
        </section>
      </div>
    </template>
  </section>
</template>
