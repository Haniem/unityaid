<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { CalendarDays, CheckCircle2, Clock3, MapPin, QrCode, Star, Users } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import {
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
const errorMessage = ref('')
const successMessage = ref('')

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
const approvedParticipantOptions = computed(() => approvedParticipants.value.map((app) => ({ id: app.userId, name: app.userName })))
const myFeedback = computed(() => feedback.value.find((entry) => entry.userId === currentUserId.value))
const canLeaveFeedback = computed(() => item.value?.status === 'completed' && !canManageEvent.value)

const applicationStatusLabels: Record<EventApplication['status'], string> = {
  pending: 'На рассмотрении',
  approved: 'Подтверждена',
  waitlisted: 'В листе ожидания',
  rejected: 'Отклонена',
  cancelled: 'Отменена'
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
    const [appsResponse, shiftsResponse] = await Promise.all([
      fetchEventApplications(eventId),
      fetchEventShifts(eventId)
    ])
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
  const index = applications.value.findIndex((item) => item.id === app.id)
  if (index >= 0) applications.value[index] = response.item
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
  const index = attendance.value.findIndex((item) => item.id === row.id)
  if (index >= 0) attendance.value[index] = response.item
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
    <article v-else-if="item" class="detail-layout wide">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/calendar">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.status }} · {{ item.format }}</span>
        <p class="detail-summary">{{ item.description || 'Описание мероприятия пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><CalendarDays :size="18" /><span>{{ formatDateTime(item.startsAt) }} - {{ formatDateTime(item.endsAt) }}</span></div>
          <div><MapPin :size="18" /><span>{{ item.location || 'Место не указано' }}</span></div>
          <div><Users :size="18" /><span>Лимит: {{ item.maxParticipants || 'не указан' }}</span></div>
          <div v-if="canManageEvent"><QrCode :size="18" /><span>Код отметки: {{ item.checkinCode }}</span></div>
          <div><Star :size="18" /><span>Средняя оценка: {{ feedbackCount ? feedbackAverage.toFixed(1) : 'нет отзывов' }}</span></div>
        </div>
        <button v-if="canManageEvent && item.status !== 'completed'" class="primary-action" type="button" @click="finishEvent">
          <CheckCircle2 :size="17" /> Завершить и начислить часы
        </button>
      </div>

      <div class="management-grid">
        <section class="detail-panel">
          <p class="eyebrow">Участие</p>
          <form v-if="!myApplication" class="settings-form" @submit.prevent="applyToEvent">
            <textarea v-model="applicationMessage" rows="3" placeholder="Комментарий к заявке" />
            <button class="primary-action" type="submit">Подать заявку</button>
          </form>
          <div v-else class="member-row application-row">
            <div>
              <strong>Моя заявка</strong>
              <small>{{ applicationStatusLabels[myApplication.status] }}</small>
            </div>
            <span class="status-pill">{{ myApplication.status }}</span>
            <button class="secondary-action danger-action" type="button" @click="cancelApplication">Отменить</button>
          </div>
          <p v-if="successMessage" class="form-success">{{ successMessage }}</p>
        </section>

        <section v-if="canManageEvent" class="detail-panel">
          <p class="eyebrow">Заявки координатора</p>
          <article v-for="app in applicationsPageItems" :key="app.id" class="member-row application-row">
            <div>
              <strong>{{ app.userName }}</strong>
              <small>{{ app.email }} · {{ app.message || 'без комментария' }}</small>
            </div>
            <span class="status-pill">{{ applicationStatusLabels[app.status] }}</span>
            <CustomSelect :model-value="app.status" :options="applicationStatusOptions" @update:model-value="setApplicationStatus(app, $event as EventApplication['status'])" />
          </article>
          <PaginationBar v-model:page="applicationsPage" :per-page="applicationsPerPage" :total="applications.length" />
          <p v-if="applications.length === 0" class="empty-state">Заявок пока нет.</p>
        </section>

        <section v-if="canManageEvent" class="detail-panel">
          <p class="eyebrow">Участники</p>
          <article v-for="app in approvedParticipants" :key="app.id" class="member-row application-row">
            <div>
              <strong>{{ app.userName }}</strong>
              <small>{{ app.email }}</small>
            </div>
            <span class="status-pill">подтвержден</span>
          </article>
          <p v-if="approvedParticipants.length === 0" class="empty-state">Подтвержденных участников пока нет.</p>
        </section>

        <section v-if="canManageEvent" class="detail-panel">
          <p class="eyebrow">Посещаемость</p>
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
            <button class="secondary-action" type="button" @click="saveAttendance(row)">Сохранить</button>
          </article>
          <PaginationBar v-model:page="attendancePage" :per-page="attendancePerPage" :total="attendance.length" />
        </section>

        <section v-if="canManageEvent" class="detail-panel">
          <p class="eyebrow">Смены</p>
          <form class="settings-form" @submit.prevent="addShift">
            <input v-model="shiftForm.title" placeholder="Название смены" required />
            <input v-model="shiftForm.startsAt" type="datetime-local" required />
            <input v-model="shiftForm.endsAt" type="datetime-local" required />
            <input v-model="shiftForm.capacity" type="number" min="1" placeholder="лимит" />
            <button class="primary-action" type="submit">Добавить смену</button>
          </form>
          <p v-for="shift in shifts" :key="shift.id">{{ shift.title }} · {{ formatDateTime(shift.startsAt) }}</p>
        </section>

        <section class="detail-panel feedback-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Обратная связь</p>
              <h2>{{ feedbackCount ? `${feedbackAverage.toFixed(1)} / 5` : 'Пока нет оценок' }}</h2>
            </div>
          </div>

          <form v-if="canLeaveFeedback" class="settings-form" @submit.prevent="sendFeedback">
            <div class="rating-control" role="radiogroup" aria-label="Оценка мероприятия">
              <button
                v-for="value in 5"
                :key="value"
                type="button"
                :class="{ active: feedbackForm.rating >= value }"
                @click="feedbackForm.rating = value"
              >
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
        </section>
      </div>
    </article>
  </section>
</template>
