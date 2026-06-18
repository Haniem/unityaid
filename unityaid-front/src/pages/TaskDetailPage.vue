<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Clock3, Flag, Link as LinkIcon } from 'lucide-vue-next'
import {
  addTaskComment,
  addTaskTimeEntry,
  approveTask,
  assignTask,
  fetchTask,
  fetchTaskAttachments,
  fetchTaskComments,
  fetchTaskTimeEntries
} from '../entities/tasks/api'
import type { TaskAttachment, TaskComment, TaskItem, TaskTimeEntry } from '../entities/tasks/types'
import { fetchVolunteers } from '../entities/users/api'
import type { PossibleValue } from '../entities/forms/types'
import { formatDateTime } from '../shared/date'
import AsyncSelect from '../shared/ui/AsyncSelect.vue'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import { authState } from '../entities/auth/store'
import { canManageContent } from '../shared/permissions'

const route = useRoute()
const taskId = String(route.params.id)
const item = ref<TaskItem | null>(null)
const comments = ref<TaskComment[]>([])
const attachments = ref<TaskAttachment[]>([])
const timeEntries = ref<TaskTimeEntry[]>([])
const errorMessage = ref('')
const activeWorkTab = ref<'assignees' | 'comments' | 'time'>('assignees')

const assignment = reactive({ userId: '', role: 'assignee' })
const comment = ref('')
const timeEntry = reactive({ hours: '', note: '' })
const canManageTask = computed(() => canManageContent(authState.user))

const assignmentRoleOptions = [
  { id: 'assignee', name: 'Исполнитель' },
  { id: 'co_assignee', name: 'Соисполнитель' }
]

const statusLabels: Record<string, string> = {
  created: 'Создана',
  assigned: 'Назначена',
  in_progress: 'В работе',
  review: 'На проверке',
  completed: 'Выполнена'
}

const priorityLabels: Record<string, string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий'
}

const assigneeRoleLabels: Record<string, string> = {
  assignee: 'исполнитель',
  co_assignee: 'соисполнитель'
}

const timeStatusLabels: Record<string, string> = {
  pending: 'На проверке',
  approved: 'Подтверждено',
  rejected: 'Отклонено'
}

function initials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('') || 'П'
}

async function load() {
  try {
    const [taskResponse, commentsResponse, attachmentsResponse, timeResponse] = await Promise.all([
      fetchTask(taskId),
      fetchTaskComments(taskId),
      fetchTaskAttachments(taskId),
      fetchTaskTimeEntries(taskId)
    ])
    item.value = taskResponse.item
    comments.value = commentsResponse.items
    attachments.value = attachmentsResponse.items
    timeEntries.value = timeResponse.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить задачу'
  }
}

async function loadUserOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchVolunteers({ search: params.search })
  const items: PossibleValue[] = response.items
    .slice((params.page - 1) * params.perPage, params.page * params.perPage)
    .map((user) => ({ id: user.userId, name: `${user.lastName} ${user.firstName} · ${user.email}` }))
  return { items }
}

async function submitAssignment() {
  if (!canManageTask.value) return
  if (!assignment.userId) return
  item.value = (await assignTask(taskId, assignment)).item
  assignment.userId = ''
  assignment.role = 'assignee'
}

async function submitComment() {
  if (!comment.value.trim()) return
  await addTaskComment(taskId, comment.value)
  comment.value = ''
  comments.value = (await fetchTaskComments(taskId)).items
}

async function submitTime() {
  if (!canManageTask.value) return
  await addTaskTimeEntry(taskId, { hours: Number(timeEntry.hours), note: timeEntry.note })
  timeEntry.hours = ''
  timeEntry.note = ''
  timeEntries.value = (await fetchTaskTimeEntries(taskId)).items
}

async function confirmDone() {
  item.value = (await approveTask(taskId)).item
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <article v-else-if="item" class="detail-layout wide task-detail-page">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/tasks">К списку</RouterLink>
      </div>

      <div class="detail-panel task-hero-panel">
        <span class="status-pill">{{ statusLabels[item.status] || item.status }}</span>
        <p class="detail-summary">{{ item.description || 'Описание задачи пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><Flag :size="18" /><span>Приоритет: {{ priorityLabels[item.priority] || item.priority }}</span></div>
          <div><Clock3 :size="18" /><span>Срок: {{ formatDateTime(item.dueAt) || 'не указан' }}</span></div>
          <div>
            <LinkIcon :size="18" />
            <RouterLink v-if="item.eventId" :to="`/calendar/${item.eventId}`">{{ item.eventTitle || 'Мероприятие' }}</RouterLink>
            <span v-else>Без мероприятия</span>
          </div>
        </div>
        <div class="task-attachments-inline">
          <p class="eyebrow">Вложения</p>
          <div v-if="attachments.length" class="attachment-list">
            <a v-for="file in attachments" :key="file.id" :href="file.fileUrl" target="_blank">{{ file.fileName }}</a>
          </div>
          <p v-else class="muted-text">Вложения добавляются в форме редактирования задачи.</p>
        </div>
        <button class="primary-action task-confirm-action" type="button" @click="confirmDone">Подтвердить выполнение</button>
      </div>

      <div class="tabs task-work-tabs">
        <button :class="{ active: activeWorkTab === 'assignees' }" type="button" @click="activeWorkTab = 'assignees'">Исполнители</button>
        <button :class="{ active: activeWorkTab === 'comments' }" type="button" @click="activeWorkTab = 'comments'">Комментарии</button>
        <button :class="{ active: activeWorkTab === 'time' }" type="button" @click="activeWorkTab = 'time'">Учет времени</button>
      </div>

      <div class="task-workspace">
        <section v-if="activeWorkTab === 'assignees'" class="detail-panel">
          <p class="eyebrow">Исполнители</p>
          <form v-if="canManageTask" class="inline-member-form task-assignment-form" @submit.prevent="submitAssignment">
            <AsyncSelect v-model="assignment.userId" :load-options="loadUserOptions" placeholder="Выберите исполнителя" />
            <CustomSelect v-model="assignment.role" :options="assignmentRoleOptions" />
            <button class="primary-action" type="submit">Назначить</button>
          </form>
          <div class="assignee-list">
            <RouterLink v-for="assignee in item.assignees" :key="assignee.userId" class="assignee-chip" :to="`/profile/${assignee.userId}`">
              <span class="mini-avatar">{{ initials(assignee.name) }}</span>
              <span>
                <strong>{{ assignee.name }}</strong>
                <small>{{ assigneeRoleLabels[assignee.role] || assignee.role }}</small>
              </span>
            </RouterLink>
            <p v-if="!item.assignees.length" class="muted-text">Исполнители еще не назначены.</p>
          </div>
        </section>

        <section v-if="activeWorkTab === 'comments'" class="detail-panel">
          <p class="eyebrow">Комментарии</p>
          <form class="settings-form" @submit.prevent="submitComment">
            <textarea v-model="comment" rows="3" placeholder="Добавьте рабочий комментарий" required />
            <button class="primary-action" type="submit">Добавить</button>
          </form>
          <div class="activity-list">
            <p v-for="entry in comments" :key="entry.id"><strong>{{ entry.userName }}:</strong> {{ entry.content }}</p>
            <p v-if="!comments.length" class="muted-text">Комментариев пока нет.</p>
          </div>
        </section>

        <section v-if="activeWorkTab === 'time'" class="detail-panel">
          <p class="eyebrow">Учет времени</p>
          <form v-if="canManageTask" class="inline-member-form task-time-form" @submit.prevent="submitTime">
            <input v-model="timeEntry.hours" type="number" step="0.25" placeholder="часы" required />
            <input v-model="timeEntry.note" placeholder="комментарий" />
            <button class="primary-action" type="submit">Учесть</button>
          </form>
          <div class="activity-list">
            <p v-for="entry in timeEntries" :key="entry.id">{{ entry.userName }} · {{ entry.hours }} ч. · {{ timeStatusLabels[entry.status] || entry.status }}</p>
            <p v-if="!timeEntries.length" class="muted-text">Записей времени пока нет.</p>
          </div>
        </section>
      </div>
    </article>
  </section>
</template>
