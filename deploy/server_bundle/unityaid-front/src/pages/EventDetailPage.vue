<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { CalendarDays, CheckCircle2, Clock3, MapPin, QrCode, Star, Users } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import {
  bulkMarkEventAttendance,
  bulkUpdateEventApplications,
  completeEvent,
  createEventApplication,
  createEventFeedback,
  createEventShift,
  deleteEventApplication,
  fetchEvent,
  fetchEventApplications,
  fetchEventAttendance,
  fetchEventFeedback,
  fetchEventShifts,
  markEventAttendance,
  updateEventApplication,
  updateEventAttendance
} from '../entities/events/api'
import type { EventApplication, EventAttendance, EventFeedback, EventItem, EventShift } from '../entities/events/types'
import { formatDateTime, fromDatetimeLocal } from '../shared/date'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

const route = useRoute()
const eventId = String(route.params.id)
const item = ref<EventItem | null>(null)
const applications = ref<EventApplication[]>([])
const attendance = ref<EventAttendance[]>([])
const shifts = ref<EventShift[]>([])
const feedback = ref<EventFeedback[]>([])
const feedbackAverage = ref(0)
const feedbackCount = ref(0)
const attendanceHours = ref<Record<string, string>>({})
const selectedApplicationIds = ref<string[]>([])
const selectedAttendanceUserIds = ref<string[]>([])
const bulkRejectionReason = ref('')
const bulkAttendanceHours = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const activeTab = ref<'participants' | 'applications' | 'feedback'>('participants')

const applicationMessage = ref('')
const attendanceForm = reactive({ userId: '', hours: '' })
const shiftForm = reactive({ title: '', startsAt: '', endsAt: '', capacity: '' })
const feedbackForm = reactive({ rating: 5, comment: '' })
const { page: applicationsPage, perPage: applicationsPerPage, pageItems: applicationsPageItems } = useClientPagination(applications, 8)
const { page: attendancePage, perPage: attendancePerPage, pageItems: attendancePageItems } = useClientPagination(attendance, 8)

const applicationStatusOptions = [
  { id: 'pending', name: 'На рассмотрении' },
  { id: 'approved', name: 'Подтверждена' },
  { id: 'waitlisted', name: 'Лист ожидания' },
  { id: 'rejected', name: 'Отклонена' },
  { id: 'cancelled', name: 'Отменена' }
]

const currentUserId = computed(() => authState.user?.id ?? '')
const canManageEvent = computed(() => {
  if (!item.value || !authState.user) return false
  const systemAdmin = authState.user.systemRoles?.some((role) => role.code === 'system_admin') ?? false
  const organizationRole = authState.user.organizations.some(
    (membership) =>
      membership.organizationId === item.value?.organizationId &&
      ['super_admin', 'org_admin', 'coordinator'].includes(membership.role)
  )
  return systemAdmin || authState.user.primaryRole === 'super_admin' || organizationRole
})
const myApplication = computed(() => applications.value.find((app) => app.userId === currentUserId.value))
const approvedParticipants = computed(() => applications.value.filter((app) => app.status === 'approved'))
const pendingApplications = computed(() => applications.value.filter((app) => app.status === 'pending'))
const approvedParticipantOptions = computed(() => approvedParticipants.value.map((app) => ({ id: app.userId, name: app.userName })))
const myFeedback = computed(() => feedback.value.find((entry) => entry.userId === currentUserId.value))
const canLeaveFeedback = computed(() => item.value?.status === 'completed' && !canManageEvent.value)

const coordinatorStats = computed(() => [
  { label: 'Заявки', value: applications.value.length },
  { label: 'Ожидают', value: pendingApplications.value.length },
  { label: 'Участники', value: approvedParticipants.value.length },
  { label: 'Отмечены', value: attendance.value.length }
])

const applicationStatusLabels: Record<EventApplication['status'], string> = {
  pending: 'На рассмотрении',
  approved: 'Подтверждена',
  waitlisted: 'В листе ожидания',
  rejected: 'Отклонена',
  cancelled: 'Отменена'
}

const eventStatusLabels: Record<string, string> = {
  draft: 'Черновик',
  published: 'Опубликовано',
  completed: 'Завершено',
  cancelled: 'Отменено'
}

const eventFormatLabels: Record<string, string> = {
  online: 'Онлайн',
  offline: 'Очно',
  hybrid: 'Гибрид'
}

async function loadFeedback() {
  const response = await fetchEventFeedback(eventId)
  feedback.value = response.items
  feedbackAverage.value = response.averageRating
  feedbackCount.value = response.feedbackCount
}

async function load() {
  try {
    errorMessage.value = ''
    item.value = (await fetchEvent(eventId)).item
    const [appsResponse, shiftsResponse] = await Promise.all([fetchEventApplications(eventId), fetchEventShifts(eventId)])
    applications.value = appsResponse.items
    shifts.value = shiftsResponse.items
    await loadFeedback()
    if (canManageEvent.value) {
      const attendanceResponse = await fetchEventAttendance(eventId)
      attendance.value = attendanceResponse.items
      attendanceHours.value = Object.fromEntries(attendanceResponse.items.map((row) => [row.id, String(row.hours)]))
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить мероприятие'
  }
}

async function reloadApplications() {
  applications.value = (await fetchEventApplications(eventId)).items
}

async function applyToEvent() {
  successMessage.value = ''
  const response = await createEventApplication(eventId, { message: applicationMessage.value })
  applicationMessage.value = ''
  await reloadApplications()
  successMessage.value = `Заявка создана: ${applicationStatusLabels[response.item.status]}`
}

async function cancelApplication() {
  if (!myApplication.value) return
  await deleteEventApplication(eventId, myApplication.value.id)
  await reloadApplications()
  successMessage.value = 'Заявка отменена'
}

async function setApplicationStatus(app: EventApplication, status: EventApplication['status']) {
  const response = await updateEventApplication(eventId, app.id, status)
  const index = applications.value.findIndex((entry) => entry.id === app.id)
  if (index >= 0) applications.value[index] = response.item
}

async function bulkSetApplicationStatus(status: EventApplication['status']) {
  if (!selectedApplicationIds.value.length) return
  const response = await bulkUpdateEventApplications(eventId, {
    applicationIds: selectedApplicationIds.value,
    status,
    rejectionReason: status === 'rejected' ? bulkRejectionReason.value : ''
  })
  for (const updated of response.items) {
    const index = applications.value.findIndex((app) => app.id === updated.id)
    if (index >= 0) applications.value[index] = updated
  }
  selectedApplicationIds.value = []
  bulkRejectionReason.value = ''
  successMessage.value = status === 'approved' ? 'Выбранные заявки подтверждены' : 'Выбранные заявки отклонены'
}

async function markAttendance() {
  if (!attendanceForm.userId) return
  await markEventAttendance(eventId, {
    userId: attendanceForm.userId,
    checkinCode: item.value?.checkinCode,
    hours: attendanceForm.hours ? Number(attendanceForm.hours) : undefined
  })
  attendanceForm.userId = ''
  attendanceForm.hours = ''
  const response = await fetchEventAttendance(eventId)
  attendance.value = response.items
  attendanceHours.value = Object.fromEntries(response.items.map((row) => [row.id, String(row.hours)]))
}

async function saveAttendance(row: EventAttendance) {
  const hours = Number(attendanceHours.value[row.id] || row.hours)
  const response = await updateEventAttendance(eventId, row.id, { hours })
  const index = attendance.value.findIndex((entry) => entry.id === row.id)
  if (index >= 0) attendance.value[index] = response.item
}

async function markSelectedAttendance() {
  if (!selectedAttendanceUserIds.value.length) return
  const response = await bulkMarkEventAttendance(eventId, {
    userIds: selectedAttendanceUserIds.value,
    hours: bulkAttendanceHours.value ? Number(bulkAttendanceHours.value) : undefined
  })
  attendance.value = response.items
  attendanceHours.value = Object.fromEntries(response.items.map((row) => [row.id, String(row.hours)]))
  selectedAttendanceUserIds.value = []
  bulkAttendanceHours.value = ''
  successMessage.value = 'Посещаемость выбранных участников отмечена'
}

async function addShift() {
  await createEventShift(eventId, {
    title: shiftForm.title,
    startsAt: fromDatetimeLocal(shiftForm.startsAt),
    endsAt: fromDatetimeLocal(shiftForm.endsAt),
    capacity: shiftForm.capacity ? Number(shiftForm.capacity) : null
  })
  shiftForm.title = ''
  shiftForm.startsAt = ''
  shiftForm.endsAt = ''
  shiftForm.capacity = ''
  shifts.value = (await fetchEventShifts(eventId)).items
}

async function sendFeedback() {
  await createEventFeedback(eventId, feedbackForm)
  feedbackForm.rating = 5
  feedbackForm.comment = ''
  await loadFeedback()
  successMessage.value = 'Спасибо, отзыв сохранен'
}

async function finishEvent() {
  await completeEvent(eventId)
  item.value = (await fetchEvent(eventId)).item
  await loadFeedback()
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value))
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <article v-else-if="item" class="detail-layout wide event-detail-page">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/calendar">К списку</RouterLink>
      </div>

      <div class="detail-panel event-hero-panel">
        <span class="status-pill">{{ eventStatusLabels[item.status] || item.status }} · {{ eventFormatLabels[item.format] || item.format }}</span>
        <p class="detail-summary">{{ item.description || 'Описание мероприятия пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><CalendarDays :size="18" /><span>{{ formatDateTime(item.startsAt) }} - {{ formatDateTime(item.endsAt) }}</span></div>
          <div><MapPin :size="18" /><span>{{ item.location || 'Место не указано' }}</span></div>
          <div><Users :size="18" /><span>Лимит: {{ item.maxParticipants || 'не указан' }}</span></div>
          <div v-if="canManageEvent"><QrCode :size="18" /><span>Код отметки: {{ item.checkinCode }}</span></div>
          <div><Star :size="18" /><span>Оценка: {{ feedbackCount ? feedbackAverage.toFixed(1) : 'нет отзывов' }}</span></div>
        </div>
        <button v-if="canManageEvent && item.status !== 'completed'" class="primary-action event-complete-action" type="button" @click="finishEvent">
          <CheckCircle2 :size="17" /> Завершить и начислить часы
        </button>
      </div>

      <section v-if="canManageEvent" class="detail-panel coordinator-console">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Координатор</p>
            <h2>Оперативное управление</h2>
          </div>
        </div>
        <div class="detail-metrics">
          <div v-for="stat in coordinatorStats" :key="stat.label">
            <CheckCircle2 :size="18" />
            <span>{{ stat.label }}: {{ stat.value }}</span>
          </div>
        </div>
      </section>

      <section class="detail-panel event-workspace">
        <div class="tabs">
          <button :class="{ active: activeTab === 'participants' }" type="button" @click="activeTab = 'participants'">Участники</button>
          <button :class="{ active: activeTab === 'applications' }" type="button" @click="activeTab = 'applications'">Заявки</button>
          <button :class="{ active: activeTab === 'feedback' }" type="button" @click="activeTab = 'feedback'">Отзывы</button>
        </div>

        <div v-if="activeTab === 'participants'" class="tab-panel">
          <div v-if="canManageEvent" class="context-toolbar">
            <input v-model="bulkAttendanceHours" type="number" min="0" step="0.25" placeholder="Часы для выбранных" />
            <button class="primary-action" type="button" :disabled="selectedAttendanceUserIds.length === 0" @click="markSelectedAttendance">
              Отметить выбранных
            </button>
          </div>
          <article v-for="app in approvedParticipants" :key="app.id" class="member-row application-row">
            <input v-if="canManageEvent" v-model="selectedAttendanceUserIds" type="checkbox" :value="app.userId" aria-label="Выбрать участника" />
            <div>
              <strong>{{ app.userName }}</strong>
              <small>{{ app.email }}</small>
            </div>
            <span class="status-pill">Подтвержден</span>
          </article>
          <p v-if="approvedParticipants.length === 0" class="empty-state">Подтвержденных участников пока нет.</p>

          <template v-if="canManageEvent">
            <form class="inline-member-form attendance-form" @submit.prevent="markAttendance">
              <CustomSelect v-model="attendanceForm.userId" :options="approvedParticipantOptions" placeholder="Участник" />
              <input v-model="attendanceForm.hours" type="number" step="0.25" min="0" placeholder="часы" />
              <button class="primary-action" type="submit"><Clock3 :size="17" /> Отметить</button>
            </form>
            <article v-for="row in attendancePageItems" :key="row.id" class="member-row attendance-row">
              <div>
                <strong>{{ row.userName }}</strong>
                <small>{{ row.email }}</small>
              </div>
              <input v-model="attendanceHours[row.id]" type="number" min="0" step="0.25" />
              <button class="secondary-action" type="button" @click="saveAttendance(row)">Сохранить часы</button>
            </article>
            <PaginationBar v-model:page="attendancePage" :per-page="attendancePerPage" :total="attendance.length" />
          </template>
        </div>

        <div v-if="activeTab === 'applications'" class="tab-panel">
          <form v-if="!myApplication" class="settings-form application-form" @submit.prevent="applyToEvent">
            <textarea v-model="applicationMessage" rows="3" placeholder="Комментарий к заявке" />
            <button class="primary-action" type="submit">Подать заявку</button>
          </form>
          <div v-else class="member-row application-row">
            <div>
              <strong>Моя заявка</strong>
              <small>{{ applicationStatusLabels[myApplication.status] }}</small>
            </div>
            <button class="secondary-action danger-action" type="button" @click="cancelApplication">Отменить заявку</button>
          </div>
          <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

          <div v-if="canManageEvent" class="context-toolbar">
            <button class="secondary-action" type="button" :disabled="selectedApplicationIds.length === 0" @click="bulkSetApplicationStatus('approved')">
              Подтвердить выбранные заявки
            </button>
            <input v-model="bulkRejectionReason" placeholder="Причина отклонения" />
            <button class="secondary-action danger-action" type="button" :disabled="selectedApplicationIds.length === 0" @click="bulkSetApplicationStatus('rejected')">
              Отклонить выбранные заявки
            </button>
          </div>
          <article v-for="app in applicationsPageItems" :key="app.id" class="member-row application-row">
            <input v-if="canManageEvent" v-model="selectedApplicationIds" type="checkbox" :value="app.id" aria-label="Выбрать заявку" />
            <div>
              <strong>{{ app.userName }}</strong>
              <small>{{ app.email }} · {{ app.message || 'без комментария' }}</small>
              <small v-if="app.rejectionReason">Причина: {{ app.rejectionReason }}</small>
            </div>
            <span class="status-pill">{{ applicationStatusLabels[app.status] }}</span>
            <CustomSelect v-if="canManageEvent" :model-value="app.status" :options="applicationStatusOptions" @update:model-value="setApplicationStatus(app, $event as EventApplication['status'])" />
          </article>
          <PaginationBar v-model:page="applicationsPage" :per-page="applicationsPerPage" :total="applications.length" />
          <p v-if="applications.length === 0" class="empty-state">Заявок пока нет.</p>
        </div>

        <div v-if="activeTab === 'feedback'" class="tab-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Обратная связь</p>
              <h2>{{ feedbackCount ? `${feedbackAverage.toFixed(1)} / 5` : 'Пока нет оценок' }}</h2>
            </div>
          </div>
          <form v-if="canLeaveFeedback" class="settings-form" @submit.prevent="sendFeedback">
            <div class="rating-control" role="radiogroup" aria-label="Оценка мероприятия">
              <button v-for="value in 5" :key="value" type="button" :class="{ active: feedbackForm.rating >= value }" @click="feedbackForm.rating = value">
                <Star :size="22" />
              </button>
            </div>
            <textarea v-model="feedbackForm.comment" rows="3" placeholder="Комментарий" />
            <button class="primary-action" type="submit">{{ myFeedback ? 'Обновить отзыв' : 'Отправить отзыв' }}</button>
          </form>
          <p v-else-if="!canManageEvent && item.status !== 'completed'" class="empty-state">Форма отзыва откроется после завершения мероприятия.</p>
          <article v-for="entry in feedback" :key="entry.id" class="feedback-row">
            <div>
              <strong>{{ entry.userName }}</strong>
              <small>{{ formatDate(entry.createdAt) }}</small>
            </div>
            <span class="status-pill">{{ entry.rating }}/5</span>
            <p>{{ entry.comment || 'Без комментария' }}</p>
          </article>
          <p v-if="feedback.length === 0" class="empty-state">Отзывов пока нет.</p>
        </div>
      </section>

      <section v-if="canManageEvent" class="detail-panel">
        <p class="eyebrow">Смены</p>
        <form class="settings-form shift-form" @submit.prevent="addShift">
          <input v-model="shiftForm.title" placeholder="Название смены" required />
          <input v-model="shiftForm.startsAt" type="datetime-local" required />
          <input v-model="shiftForm.endsAt" type="datetime-local" required />
          <input v-model="shiftForm.capacity" type="number" min="1" placeholder="лимит" />
          <button class="primary-action" type="submit">Добавить смену</button>
        </form>
        <p v-for="shift in shifts" :key="shift.id">{{ shift.title }} · {{ formatDateTime(shift.startsAt) }}</p>
      </section>
    </article>
  </section>
</template>
