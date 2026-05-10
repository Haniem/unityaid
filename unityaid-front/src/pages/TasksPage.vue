<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Eye, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { fetchOrganizations } from '../entities/organizations/api'
import type { Organization } from '../entities/organizations/types'
import { fetchEvents } from '../entities/events/api'
import type { EventItem } from '../entities/events/types'
import { createTask, deleteTask, fetchTasks, updateTask } from '../entities/tasks/api'
import type { TaskItem, TaskPayload } from '../entities/tasks/types'
import { formatDateTime, fromDatetimeLocal, toDatetimeLocal } from '../shared/date'

const items = ref<TaskItem[]>([])
const organizations = ref<Organization[]>([])
const events = ref<EventItem[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const statusFilter = ref('')
const priorityFilter = ref('')
const form = reactive({
  organizationId: '',
  eventId: '',
  title: '',
  description: '',
  status: 'created' as TaskPayload['status'],
  priority: 'medium' as TaskPayload['priority'],
  dueAt: ''
})

const filteredEvents = computed(() => events.value.filter((event) => event.organizationId === form.organizationId))

function resetForm() {
  editingId.value = null
  form.organizationId = organizations.value[0]?.id ?? ''
  form.eventId = ''
  form.title = ''
  form.description = ''
  form.status = 'created'
  form.priority = 'medium'
  form.dueAt = ''
  errorMessage.value = ''
}

function openCreateModal() {
  resetForm()
  isModalOpen.value = true
}

function openEditModal(item: TaskItem) {
  editingId.value = item.id
  form.organizationId = item.organizationId
  form.eventId = item.eventId ?? ''
  form.title = item.title
  form.description = item.description
  form.status = item.status
  form.priority = item.priority
  form.dueAt = toDatetimeLocal(item.dueAt)
  errorMessage.value = ''
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

function payload(): TaskPayload {
  return {
    organizationId: form.organizationId,
    eventId: form.eventId || null,
    title: form.title,
    description: form.description,
    status: form.status,
    priority: form.priority,
    dueAt: form.dueAt ? fromDatetimeLocal(form.dueAt) : null
  }
}

async function load() {
  const [tasksResponse, organizationsResponse, eventsResponse] = await Promise.all([
    fetchTasks({ status: statusFilter.value, priority: priorityFilter.value }),
    fetchOrganizations(),
    fetchEvents()
  ])
  items.value = tasksResponse.items
  organizations.value = organizationsResponse.items
  events.value = eventsResponse.items
  if (!form.organizationId) resetForm()
}

async function submit() {
  errorMessage.value = ''
  try {
    if (editingId.value) await updateTask(editingId.value, payload())
    else await createTask(payload())
    closeModal()
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить задачу'
  }
}

async function remove(item: TaskItem) {
  if (!confirm(`Удалить задачу "${item.title}"?`)) return
  await deleteTask(item.id)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Продуктивность</p>
        <h1>Мои задачи</h1>
      </div>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать задачу</span>
      </button>
    </div>

    <div class="filter-bar">
      <select v-model="statusFilter" @change="load">
        <option value="">Все статусы</option>
        <option value="created">Создана</option>
        <option value="assigned">Назначена</option>
        <option value="in_progress">В работе</option>
        <option value="review">На проверке</option>
        <option value="completed">Выполнена</option>
      </select>
      <select v-model="priorityFilter" @change="load">
        <option value="">Все приоритеты</option>
        <option value="low">Низкий</option>
        <option value="medium">Средний</option>
        <option value="high">Высокий</option>
      </select>
    </div>

    <div class="task-board">
      <article v-for="item in items" :key="item.id" class="task-card">
        <RouterLink class="task-card-main" :to="`/tasks/${item.id}`">
          <span :class="['priority-dot', item.priority]"></span>
          <h2>{{ item.title }}</h2>
          <p>{{ item.description || 'Описание пока не заполнено.' }}</p>
          <small>{{ item.organizationName }} · {{ item.eventTitle || 'без мероприятия' }} · {{ formatDateTime(item.dueAt) }}</small>
        </RouterLink>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/tasks/${item.id}`" aria-label="Открыть"><Eye :size="17" /></RouterLink>
          <button class="icon-button" type="button" aria-label="Редактировать" @click="openEditModal(item)"><Pencil :size="17" /></button>
          <button class="icon-button" type="button" aria-label="Удалить" @click="remove(item)"><Trash2 :size="17" /></button>
        </div>
      </article>
    </div>

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ editingId ? 'Редактирование задачи' : 'Новая задача' }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <label><span>Организация</span><select v-model="form.organizationId"><option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option></select></label>
        <label><span>Мероприятие</span><select v-model="form.eventId"><option value="">Без мероприятия</option><option v-for="event in filteredEvents" :key="event.id" :value="event.id">{{ event.title }}</option></select></label>
        <label><span>Название</span><input v-model="form.title" required /></label>
        <label><span>Описание</span><textarea v-model="form.description" rows="4" /></label>
        <div class="form-columns">
          <label><span>Статус</span><select v-model="form.status"><option value="created">Создана</option><option value="assigned">Назначена</option><option value="in_progress">В работе</option><option value="review">На проверке</option><option value="completed">Выполнена</option><option value="cancelled">Отменена</option></select></label>
          <label><span>Приоритет</span><select v-model="form.priority"><option value="low">Низкий</option><option value="medium">Средний</option><option value="high">Высокий</option></select></label>
        </div>
        <label><span>Срок</span><input v-model="form.dueAt" type="datetime-local" /></label>
        <div class="form-actions"><button class="secondary-action" type="button" @click="closeModal">Отмена</button><button class="primary-action" type="submit">Сохранить</button></div>
      </form>
    </div>
  </section>
</template>
