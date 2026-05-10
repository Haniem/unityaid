<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { CalendarDays, MapPin, QrCode, Users } from 'lucide-vue-next'
import {
  completeEvent,
  createEventApplication,
  createEventFeedback,
  createEventShift,
  fetchEvent,
  fetchEventApplications,
  fetchEventAttendance,
  fetchEventFeedback,
  fetchEventShifts,
  markEventAttendance,
  updateEventApplication
} from '../entities/events/api'
import type { EventApplication, EventAttendance, EventFeedback, EventItem, EventShift } from '../entities/events/types'
import { formatDateTime, fromDatetimeLocal } from '../shared/date'

const route = useRoute()
const eventId = String(route.params.id)
const item = ref<EventItem | null>(null)
const applications = ref<EventApplication[]>([])
const attendance = ref<EventAttendance[]>([])
const shifts = ref<EventShift[]>([])
const feedback = ref<EventFeedback[]>([])
const errorMessage = ref('')

const applicationMessage = ref('')
const attendanceForm = reactive({ userId: '', hours: '' })
const shiftForm = reactive({ title: '', startsAt: '', endsAt: '', capacity: '' })
const feedbackForm = reactive({ rating: 5, comment: '' })

async function load() {
  try {
    const [eventResponse, appsResponse, attendanceResponse, shiftsResponse, feedbackResponse] = await Promise.all([
      fetchEvent(eventId),
      fetchEventApplications(eventId),
      fetchEventAttendance(eventId),
      fetchEventShifts(eventId),
      fetchEventFeedback(eventId)
    ])
    item.value = eventResponse.item
    applications.value = appsResponse.items
    attendance.value = attendanceResponse.items
    shifts.value = shiftsResponse.items
    feedback.value = feedbackResponse.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить мероприятие'
  }
}

async function applyToEvent() {
  await createEventApplication(eventId, { message: applicationMessage.value })
  applicationMessage.value = ''
  applications.value = (await fetchEventApplications(eventId)).items
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
  attendance.value = (await fetchEventAttendance(eventId)).items
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
  feedback.value = (await fetchEventFeedback(eventId)).items
}

async function finishEvent() {
  await completeEvent(eventId)
  item.value = (await fetchEvent(eventId)).item
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
          <div><CalendarDays :size="18" /><span>{{ formatDateTime(item.startsAt) }} — {{ formatDateTime(item.endsAt) }}</span></div>
          <div><MapPin :size="18" /><span>{{ item.location || 'Место не указано' }}</span></div>
          <div><Users :size="18" /><span>Лимит: {{ item.maxParticipants || 'не указан' }}</span></div>
          <div><QrCode :size="18" /><span>QR-код отметки: {{ item.checkinCode }}</span></div>
        </div>
        <button class="primary-action" type="button" @click="finishEvent">Завершить и начислить часы</button>
      </div>

      <div class="management-grid">
        <section class="detail-panel">
          <p class="eyebrow">Заявки</p>
          <form class="inline-member-form" @submit.prevent="applyToEvent">
            <input v-model="applicationMessage" placeholder="Комментарий к заявке" />
            <button class="primary-action" type="submit">Подать заявку</button>
          </form>
          <article v-for="app in applications" :key="app.id" class="member-row">
            <div><strong>{{ app.userName }}</strong><small>{{ app.email }}</small></div>
            <span class="status-pill">{{ app.status }}</span>
            <select :value="app.status" @change="setApplicationStatus(app, ($event.target as HTMLSelectElement).value as EventApplication['status'])">
              <option value="pending">pending</option><option value="approved">approved</option><option value="waitlisted">waitlisted</option><option value="rejected">rejected</option>
            </select>
          </article>
        </section>

        <section class="detail-panel">
          <p class="eyebrow">Посещаемость</p>
          <form class="inline-member-form" @submit.prevent="markAttendance">
            <select v-model="attendanceForm.userId"><option value="">Участник</option><option v-for="app in applications" :key="app.id" :value="app.userId">{{ app.userName }}</option></select>
            <input v-model="attendanceForm.hours" type="number" step="0.25" placeholder="часы" />
            <button class="primary-action" type="submit">Отметить</button>
          </form>
          <p v-for="row in attendance" :key="row.id">{{ row.userName }} · {{ row.hours }} ч.</p>
        </section>

        <section class="detail-panel">
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

        <section class="detail-panel">
          <p class="eyebrow">Обратная связь</p>
          <form class="settings-form" @submit.prevent="sendFeedback">
            <input v-model.number="feedbackForm.rating" type="number" min="1" max="5" />
            <textarea v-model="feedbackForm.comment" rows="3" placeholder="Комментарий" />
            <button class="primary-action" type="submit">Отправить</button>
          </form>
          <p v-for="entry in feedback" :key="entry.id">{{ entry.userName }} · {{ entry.rating }}/5 · {{ entry.comment }}</p>
        </section>
      </div>
    </article>
  </section>
</template>
