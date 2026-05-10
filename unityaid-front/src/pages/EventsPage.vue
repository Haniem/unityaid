<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Eye, MapPin, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { createEvent, deleteEvent, fetchEvents, updateEvent } from '../entities/events/api'
import type { EventItem, EventPayload } from '../entities/events/types'
import { fetchOrganizations } from '../entities/organizations/api'
import type { Organization } from '../entities/organizations/types'
import { formatDateTime, fromDatetimeLocal, toDatetimeLocal } from '../shared/date'

const items = ref<EventItem[]>([])
const organizations = ref<Organization[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const form = reactive({
  organizationId: '',
  title: '',
  description: '',
  format: 'offline' as EventPayload['format'],
  status: 'published' as EventPayload['status'],
  startsAt: '',
  endsAt: '',
  location: '',
  maxParticipants: ''
})

const hasOrganizations = computed(() => organizations.value.length > 0)

function resetForm() {
  editingId.value = null
  form.organizationId = organizations.value[0]?.id ?? ''
  form.title = ''
  form.description = ''
  form.format = 'offline'
  form.status = 'published'
  form.startsAt = ''
  form.endsAt = ''
  form.location = ''
  form.maxParticipants = ''
  errorMessage.value = ''
}

function openCreateModal() {
  resetForm()
  isModalOpen.value = true
}

function openEditModal(item: EventItem) {
  editingId.value = item.id
  form.organizationId = item.organizationId
  form.title = item.title
  form.description = item.description
  form.format = item.format
  form.status = item.status
  form.startsAt = toDatetimeLocal(item.startsAt)
  form.endsAt = toDatetimeLocal(item.endsAt)
  form.location = item.location ?? ''
  form.maxParticipants = item.maxParticipants ? String(item.maxParticipants) : ''
  errorMessage.value = ''
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

function payload(): EventPayload {
  return {
    organizationId: form.organizationId,
    title: form.title,
    description: form.description,
    format: form.format,
    status: form.status,
    startsAt: fromDatetimeLocal(form.startsAt),
    endsAt: fromDatetimeLocal(form.endsAt),
    location: form.location || null,
    maxParticipants: form.maxParticipants ? Number(form.maxParticipants) : null
  }
}

async function load() {
  const [eventsResponse, organizationsResponse] = await Promise.all([fetchEvents(), fetchOrganizations()])
  items.value = eventsResponse.items
  organizations.value = organizationsResponse.items
  if (!form.organizationId) resetForm()
}

async function submit() {
  errorMessage.value = ''
  try {
    if (editingId.value) await updateEvent(editingId.value, payload())
    else await createEvent(payload())
    closeModal()
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить мероприятие'
  }
}

async function remove(item: EventItem) {
  if (!confirm(`Удалить мероприятие "${item.title}"?`)) return
  await deleteEvent(item.id)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Продуктивность</p>
        <h1>Мероприятия</h1>
      </div>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать мероприятие</span>
      </button>
    </div>

    <div class="event-list">
      <article v-for="item in items" :key="item.id" class="event-card">
        <RouterLink class="event-card-main" :to="`/calendar/${item.id}`">
          <span class="status-pill">{{ item.status }}</span>
          <h2>{{ item.title }}</h2>
          <p>{{ item.description || 'Описание пока не заполнено.' }}</p>
          <div class="meta-line">
            <MapPin :size="16" />
            <span>{{ item.location || item.format }} · {{ formatDateTime(item.startsAt) }}</span>
          </div>
          <small>{{ item.organizationName }} · лимит {{ item.maxParticipants || 'не указан' }}</small>
        </RouterLink>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/calendar/${item.id}`" aria-label="Открыть"><Eye :size="17" /></RouterLink>
          <button class="icon-button" type="button" aria-label="Редактировать" @click="openEditModal(item)"><Pencil :size="17" /></button>
          <button class="icon-button" type="button" aria-label="Удалить" @click="remove(item)"><Trash2 :size="17" /></button>
        </div>
      </article>
    </div>

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ editingId ? 'Редактирование мероприятия' : 'Новое мероприятие' }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <div v-if="!hasOrganizations" class="empty-state">Сначала создайте организацию.</div>
        <template v-else>
          <label><span>Организация</span><select v-model="form.organizationId"><option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option></select></label>
          <label><span>Название</span><input v-model="form.title" required /></label>
          <label><span>Описание</span><textarea v-model="form.description" rows="4" /></label>
          <div class="form-columns">
            <label><span>Формат</span><select v-model="form.format"><option value="offline">Офлайн</option><option value="online">Онлайн</option><option value="hybrid">Гибрид</option></select></label>
            <label><span>Статус</span><select v-model="form.status"><option value="draft">Черновик</option><option value="published">Опубликовано</option><option value="completed">Завершено</option><option value="cancelled">Отменено</option></select></label>
          </div>
          <div class="form-columns">
            <label><span>Начало</span><input v-model="form.startsAt" type="datetime-local" required /></label>
            <label><span>Окончание</span><input v-model="form.endsAt" type="datetime-local" required /></label>
          </div>
          <div class="form-columns">
            <label><span>Место</span><input v-model="form.location" /></label>
            <label><span>Лимит участников</span><input v-model="form.maxParticipants" type="number" min="1" /></label>
          </div>
          <div class="form-actions"><button class="secondary-action" type="button" @click="closeModal">Отмена</button><button class="primary-action" type="submit">Сохранить</button></div>
        </template>
      </form>
    </div>
  </section>
</template>
