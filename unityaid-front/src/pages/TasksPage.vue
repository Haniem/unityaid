<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
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
const errorMessage = ref('')
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
}

function edit(item: TaskItem) {
  editingId.value = item.id
  form.organizationId = item.organizationId
  form.eventId = item.eventId ?? ''
  form.title = item.title
  form.description = item.description
  form.status = item.status
  form.priority = item.priority
  form.dueAt = toDatetimeLocal(item.dueAt)
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
    fetchTasks(),
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
    resetForm()
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
      <div><p class="eyebrow">Продуктивность</p><h1>Мои задачи</h1></div>
    </div>

    <div class="management-grid">
      <form class="entity-form" @submit.prevent="submit">
        <h2>{{ editingId ? 'Редактирование' : 'Новая задача' }}</h2>
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
        <div class="form-actions"><button class="secondary-action" type="button" @click="resetForm">Сбросить</button><button class="primary-action" type="submit"><Plus :size="17" /> Сохранить</button></div>
      </form>

      <div class="entity-list">
        <article v-for="item in items" :key="item.id" class="entity-row">
          <div>
            <h2>{{ item.title }}</h2>
            <p>{{ item.description || 'Описание пока не заполнено.' }}</p>
            <small>{{ item.organizationName }} · {{ item.eventTitle || 'без мероприятия' }} · {{ item.priority }} · {{ formatDateTime(item.dueAt) }}</small>
          </div>
          <div class="card-actions static-actions"><button class="icon-button" type="button" @click="edit(item)"><Pencil :size="17" /></button><button class="icon-button" type="button" @click="remove(item)"><Trash2 :size="17" /></button></div>
        </article>
      </div>
    </div>
  </section>
</template>
