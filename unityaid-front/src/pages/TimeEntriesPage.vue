<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { CheckCircle2, Clock3, XCircle } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import { fetchEvents } from '../entities/events/api'
import type { EventItem } from '../entities/events/types'
import { fetchTasks } from '../entities/tasks/api'
import type { TaskItem } from '../entities/tasks/types'
import { approveTimeEntry, bulkApproveTimeEntries, createTimeEntry, fetchTimeEntries, rejectTimeEntry } from '../entities/timeentries/api'
import type { TimeEntry, TimeEntryStatus } from '../entities/timeentries/types'
import { fetchVolunteers } from '../entities/users/api'
import type { PossibleValue } from '../entities/forms/types'
import { formatDateTime } from '../shared/date'
import AsyncSelect from '../shared/ui/AsyncSelect.vue'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

const items = ref<TimeEntry[]>([])
const events = ref<EventItem[]>([])
const tasks = ref<TaskItem[]>([])
const statusFilter = ref('')
const userFilter = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const form = reactive({ targetType: 'event' as 'event' | 'task', targetId: '', hours: '', description: '' })
const { page, perPage, pageItems } = useClientPagination(items, 12)
const pendingItems = computed(() => items.value.filter((item) => item.status === 'pending'))
const { page: pendingPage, perPage: pendingPerPage, pageItems: pendingPageItems } = useClientPagination(pendingItems, 8)
const targetTypeOptions = [{ id: 'event', name: 'Мероприятие' }, { id: 'task', name: 'Задача' }]
const timeStatusOptions = [{ id: '', name: 'Все' }, { id: 'pending', name: 'На проверке' }, { id: 'approved', name: 'Подтверждено' }, { id: 'rejected', name: 'Отклонено' }]

const canReview = computed(() => {
  const systemAdmin = authState.user?.systemRoles?.some((role) => role.code === 'system_admin') ?? false
  const organizationManager =
    authState.user?.organizations.some((membership) => ['super_admin', 'org_admin', 'coordinator'].includes(membership.role)) ?? false
  return systemAdmin || authState.user?.primaryRole === 'super_admin' || organizationManager
})

const approvedHours = computed(() =>
  items.value
    .filter((item) => item.status === 'approved' && item.userId === authState.user?.id)
    .reduce((sum, item) => sum + item.hours, 0)
)
const targetOptions = computed(() => form.targetType === 'event' ? events.value.map((event) => ({ id: event.id, name: event.title })) : tasks.value.map((task) => ({ id: task.id, name: task.title })))

const statusLabels: Record<TimeEntryStatus, string> = {
  pending: 'На проверке',
  approved: 'Подтверждено',
  rejected: 'Отклонено'
}

async function load() {
  errorMessage.value = ''
  try {
    const [entriesResponse, eventsResponse, tasksResponse] = await Promise.all([
      fetchTimeEntries({ status: statusFilter.value, userId: userFilter.value }),
      fetchEvents(),
      fetchTasks()
    ])
    items.value = entriesResponse.items
    events.value = eventsResponse.items
    tasks.value = tasksResponse.items
    if (!form.targetId) updateDefaultTarget()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить учет времени'
  }
}

async function loadUserOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchVolunteers({ search: params.search })
  const items: PossibleValue[] = response.items
    .slice((params.page - 1) * params.perPage, params.page * params.perPage)
    .map((user) => ({ id: user.userId, name: `${user.lastName} ${user.firstName} · ${user.email}` }))
  return { items }
}

function updateDefaultTarget() {
  form.targetId = form.targetType === 'event' ? events.value[0]?.id ?? '' : tasks.value[0]?.id ?? ''
}

async function submit() {
  errorMessage.value = ''
  successMessage.value = ''
  if (!form.targetId || !form.hours) return
  try {
    await createTimeEntry({
      eventId: form.targetType === 'event' ? form.targetId : null,
      taskId: form.targetType === 'task' ? form.targetId : null,
      hours: Number(form.hours),
      description: form.description
    })
    form.hours = ''
    form.description = ''
    successMessage.value = 'Запись отправлена на проверку'
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить запись'
  }
}

async function review(item: TimeEntry, action: 'approve' | 'reject') {
  const response = action === 'approve' ? await approveTimeEntry(item.id) : await rejectTimeEntry(item.id)
  const index = items.value.findIndex((entry) => entry.id === item.id)
  if (index >= 0) items.value[index] = response.item
}

function targetTitle(item: TimeEntry) {
  return item.taskTitle || item.eventTitle || 'Без привязки'
}

async function approveAllPending() {
  const ids = pendingItems.value.map((item) => item.id)
  if (!ids.length) return
  const response = await bulkApproveTimeEntries(ids)
  for (const entry of response.items) {
    const index = items.value.findIndex((item) => item.id === entry.id)
    if (index >= 0) items.value[index] = entry
  }
  successMessage.value = `Подтверждено записей: ${response.items.length}`
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Учет времени</p>
        <h1>Мои часы</h1>
      </div>
      <div class="time-total">
        <Clock3 :size="20" />
        <strong>{{ approvedHours.toFixed(2) }}</strong>
        <span>подтвержденных часов</span>
      </div>
    </div>

    <div class="time-layout">
      <form class="detail-panel settings-form" @submit.prevent="submit">
        <div>
          <p class="eyebrow">Новая запись</p>
          <h2>Добавить время</h2>
        </div>

        <label>
          <span>Тип работы</span>
          <CustomSelect v-model="form.targetType" :options="targetTypeOptions" @update:model-value="updateDefaultTarget" />
        </label>

        <label>
          <span>{{ form.targetType === 'event' ? 'Мероприятие' : 'Задача' }}</span>
          <CustomSelect v-model="form.targetId" :options="targetOptions" placeholder="Выберите запись" />
        </label>

        <label>
          <span>Количество часов</span>
          <input v-model="form.hours" type="number" min="0.25" step="0.25" required />
        </label>

        <label>
          <span>Описание</span>
          <textarea v-model="form.description" rows="4" placeholder="Что было сделано" />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

        <button class="primary-action" type="submit">Отправить на проверку</button>
      </form>

      <div class="time-content">
        <section v-if="canReview" class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Проверка</p>
              <h2>Ожидают подтверждения</h2>
            </div>
            <button class="secondary-action" type="button" :disabled="pendingItems.length === 0" @click="approveAllPending">
              <CheckCircle2 :size="17" />
              <span>Подтвердить все</span>
            </button>
          </div>
          <article v-for="item in pendingPageItems" :key="item.id" class="time-row">
            <div>
              <strong>{{ item.userName }}</strong>
              <small>{{ targetTitle(item) }} · {{ item.organizationName }} · {{ item.hours }} ч.</small>
              <p>{{ item.description || 'Описание не заполнено.' }}</p>
            </div>
            <div class="card-actions">
              <button class="icon-button accent" type="button" aria-label="Подтвердить" @click="review(item, 'approve')">
                <CheckCircle2 :size="18" />
              </button>
              <button class="icon-button" type="button" aria-label="Отклонить" @click="review(item, 'reject')">
                <XCircle :size="18" />
              </button>
            </div>
          </article>
          <PaginationBar v-model:page="pendingPage" :per-page="pendingPerPage" :total="pendingItems.length" />
          <p v-if="pendingItems.length === 0" class="empty-state">Нет записей на проверке.</p>
        </section>

        <section class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">История</p>
              <h2>Записи времени</h2>
            </div>
            <label class="filter-select compact-filter">
              <span>Статус</span>
              <CustomSelect v-model="statusFilter" :options="timeStatusOptions" @update:model-value="load" />
            </label>
            <label class="filter-select compact-filter time-user-filter">
              <span>Пользователь</span>
              <AsyncSelect v-model="userFilter" :load-options="loadUserOptions" placeholder="Все пользователи" @update:model-value="load" />
            </label>
          </div>

          <article v-for="item in pageItems" :key="item.id" class="time-row">
            <div>
              <strong>{{ targetTitle(item) }}</strong>
              <small>{{ item.userName }} · {{ item.organizationName }} · {{ formatDateTime(item.createdAt) }}</small>
              <p>{{ item.description || 'Описание не заполнено.' }}</p>
            </div>
            <div class="time-row-side">
              <strong>{{ item.hours }} ч.</strong>
              <span class="status-pill">{{ statusLabels[item.status] }}</span>
            </div>
          </article>
          <PaginationBar v-model:page="page" :per-page="perPage" :total="items.length" />
          <p v-if="items.length === 0" class="empty-state">История часов пока пуста.</p>
        </section>
      </div>
    </div>
  </section>
</template>
