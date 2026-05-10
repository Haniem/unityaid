<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Clock3, Flag, Link as LinkIcon } from 'lucide-vue-next'
import {
  addTaskAttachment,
  addTaskComment,
  addTaskTimeEntry,
  approveTask,
  assignTask,
  fetchTask,
  fetchTaskAttachments,
  fetchTaskComments,
  fetchTaskStatusHistory,
  fetchTaskTimeEntries
} from '../entities/tasks/api'
import type { TaskAttachment, TaskComment, TaskItem, TaskStatusHistory, TaskTimeEntry } from '../entities/tasks/types'
import { formatDateTime } from '../shared/date'

const route = useRoute()
const taskId = String(route.params.id)
const item = ref<TaskItem | null>(null)
const comments = ref<TaskComment[]>([])
const attachments = ref<TaskAttachment[]>([])
const history = ref<TaskStatusHistory[]>([])
const timeEntries = ref<TaskTimeEntry[]>([])
const errorMessage = ref('')

const assignment = reactive({ userId: '', role: 'assignee' })
const comment = ref('')
const attachment = reactive({ fileName: '', fileUrl: '' })
const timeEntry = reactive({ hours: '', note: '' })

async function load() {
  try {
    const [taskResponse, commentsResponse, attachmentsResponse, historyResponse, timeResponse] = await Promise.all([
      fetchTask(taskId),
      fetchTaskComments(taskId),
      fetchTaskAttachments(taskId),
      fetchTaskStatusHistory(taskId),
      fetchTaskTimeEntries(taskId)
    ])
    item.value = taskResponse.item
    comments.value = commentsResponse.items
    attachments.value = attachmentsResponse.items
    history.value = historyResponse.items
    timeEntries.value = timeResponse.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить задачу'
  }
}

async function submitAssignment() {
  if (!assignment.userId) return
  item.value = (await assignTask(taskId, assignment)).item
  assignment.userId = ''
  assignment.role = 'assignee'
}

async function submitComment() {
  await addTaskComment(taskId, comment.value)
  comment.value = ''
  comments.value = (await fetchTaskComments(taskId)).items
}

async function submitAttachment() {
  await addTaskAttachment(taskId, attachment)
  attachment.fileName = ''
  attachment.fileUrl = ''
  attachments.value = (await fetchTaskAttachments(taskId)).items
}

async function submitTime() {
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
    <article v-else-if="item" class="detail-layout wide">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/tasks">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.status }}</span>
        <p class="detail-summary">{{ item.description || 'Описание задачи пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><Flag :size="18" /><span>Приоритет: {{ item.priority }}</span></div>
          <div><Clock3 :size="18" /><span>Срок: {{ formatDateTime(item.dueAt) }}</span></div>
          <div><LinkIcon :size="18" /><span>{{ item.eventTitle || 'Без мероприятия' }}</span></div>
        </div>
        <button class="primary-action" type="button" @click="confirmDone">Подтвердить выполнение</button>
      </div>

      <div class="management-grid">
        <section class="detail-panel">
          <p class="eyebrow">Исполнители</p>
          <form class="inline-member-form" @submit.prevent="submitAssignment">
            <input v-model="assignment.userId" placeholder="ID пользователя" required />
            <select v-model="assignment.role"><option value="assignee">исполнитель</option><option value="co_assignee">соисполнитель</option></select>
            <button class="primary-action" type="submit">Назначить</button>
          </form>
          <p v-for="assignee in item.assignees" :key="assignee.userId">{{ assignee.name }} · {{ assignee.role }}</p>
        </section>

        <section class="detail-panel">
          <p class="eyebrow">Комментарии</p>
          <form class="settings-form" @submit.prevent="submitComment">
            <textarea v-model="comment" rows="3" required />
            <button class="primary-action" type="submit">Добавить</button>
          </form>
          <p v-for="entry in comments" :key="entry.id">{{ entry.userName }}: {{ entry.content }}</p>
        </section>

        <section class="detail-panel">
          <p class="eyebrow">Вложения</p>
          <form class="settings-form" @submit.prevent="submitAttachment">
            <input v-model="attachment.fileName" placeholder="Название файла" required />
            <input v-model="attachment.fileUrl" placeholder="URL файла" required />
            <button class="primary-action" type="submit">Добавить</button>
          </form>
          <p v-for="file in attachments" :key="file.id"><a :href="file.fileUrl" target="_blank">{{ file.fileName }}</a></p>
        </section>

        <section class="detail-panel">
          <p class="eyebrow">Время и история</p>
          <form class="inline-member-form" @submit.prevent="submitTime">
            <input v-model="timeEntry.hours" type="number" step="0.25" placeholder="часы" required />
            <input v-model="timeEntry.note" placeholder="комментарий" />
            <button class="primary-action" type="submit">Учесть</button>
          </form>
          <p v-for="entry in timeEntries" :key="entry.id">{{ entry.userName }} · {{ entry.hours }} ч. · {{ entry.status }}</p>
          <p v-for="entry in history" :key="entry.id">{{ entry.fromStatus || '—' }} → {{ entry.toStatus }} · {{ formatDateTime(entry.createdAt) }}</p>
        </section>
      </div>
    </article>
  </section>
</template>
