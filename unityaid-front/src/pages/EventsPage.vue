<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Eye, MapPin, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { createEvent, createEventTemplate, createRecurringEvents, deleteEvent, fetchEventTemplates, fetchEvents, updateEvent } from '../entities/events/api'
import type { EventItem, EventPayload, EventTemplate } from '../entities/events/types'
import { fetchCreateForm, fetchEditForm } from '../entities/forms/api'
import type { BackendForm, FormModel } from '../entities/forms/types'
import DynamicForm from '../shared/ui/DynamicForm.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { isoFromDatetimeLocal, modelFromForm, nullable, numberOrNull, stringValue } from '../shared/forms'
import { formatDateTime } from '../shared/date'
import { useClientPagination } from '../shared/pagination'

const items = ref<EventItem[]>([])
const templates = ref<EventTemplate[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const formSchema = ref<BackendForm | null>(null)
const formModel = ref<FormModel>({})
const selectedTemplateId = ref('')
const saveAsTemplate = ref(false)
const templateName = ref('')
const isRecurring = ref(false)
const recurringFrequency = ref('weekly')
const recurringCount = ref(1)
const { page, perPage, pageItems } = useClientPagination(items, 10)

function resetForm() {
  editingId.value = null
  formSchema.value = null
  formModel.value = {}
  errorMessage.value = ''
  selectedTemplateId.value = ''
  saveAsTemplate.value = false
  templateName.value = ''
  isRecurring.value = false
  recurringFrequency.value = 'weekly'
  recurringCount.value = 1
}

async function openCreateModal() {
  resetForm()
  formSchema.value = await fetchCreateForm('events')
  formModel.value = modelFromForm(formSchema.value)
  isModalOpen.value = true
}

async function openEditModal(item: EventItem) {
  editingId.value = item.id
  formSchema.value = await fetchEditForm('events', item.id)
  formModel.value = modelFromForm(formSchema.value)
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

function payload(): EventPayload {
  return {
    organizationId: stringValue(formModel.value.organizationId),
    title: stringValue(formModel.value.title),
    description: stringValue(formModel.value.description),
    format: (stringValue(formModel.value.format) || 'offline') as EventPayload['format'],
    status: (stringValue(formModel.value.status) || 'published') as EventPayload['status'],
    startsAt: isoFromDatetimeLocal(formModel.value.startsAt) || '',
    endsAt: isoFromDatetimeLocal(formModel.value.endsAt) || '',
    location: nullable(formModel.value.location) as string | null,
    maxParticipants: numberOrNull(formModel.value.maxParticipants)
  }
}

async function load() {
  const [eventsResponse, templatesResponse] = await Promise.all([fetchEvents(), fetchEventTemplates().catch(() => ({ items: [] }))])
  items.value = eventsResponse.items
  templates.value = templatesResponse.items
}

async function submit() {
  errorMessage.value = ''
  try {
    const eventPayload = payload()
    if (editingId.value) await updateEvent(editingId.value, eventPayload)
    else if (isRecurring.value) {
      await createRecurringEvents({ event: eventPayload, frequency: recurringFrequency.value, count: recurringCount.value })
    } else {
      await createEvent(eventPayload)
    }
    if (saveAsTemplate.value && templateName.value.trim()) {
      await createEventTemplate({
        organizationId: eventPayload.organizationId,
        name: templateName.value.trim(),
        title: eventPayload.title,
        description: eventPayload.description,
        format: eventPayload.format,
        location: eventPayload.location,
        maxParticipants: eventPayload.maxParticipants,
        defaultDurationMinutes: 120
      })
    }
    closeModal()
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить мероприятие'
  }
}

function applyTemplate() {
  const template = templates.value.find((item) => item.id === selectedTemplateId.value)
  if (!template) return
  formModel.value.organizationId = template.organizationId
  formModel.value.title = template.title
  formModel.value.description = template.description
  formModel.value.format = template.format
  formModel.value.location = template.location ?? ''
  formModel.value.maxParticipants = template.maxParticipants ?? ''
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
        <p class="eyebrow">Раздел</p>
        <h1>Мероприятия</h1>
      </div>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать</span>
      </button>
    </div>

    <div class="event-list">
      <article v-for="item in pageItems" :key="item.id" class="event-card">
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
    <PaginationBar v-model:page="page" :per-page="perPage" :total="items.length" />

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ formSchema?.meta.title || (editingId ? 'Редактирование мероприятия' : 'Новое мероприятие') }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <DynamicForm v-if="formSchema" v-model="formModel" :form="formSchema" />
        <div v-if="!editingId" class="event-planning-tools">
          <label v-if="templates.length">
            <span>Шаблон</span>
            <select v-model="selectedTemplateId" @change="applyTemplate">
              <option value="">Без шаблона</option>
              <option v-for="template in templates" :key="template.id" :value="template.id">{{ template.name }}</option>
            </select>
          </label>
          <label class="toggle-field inline-toggle">
            <input v-model="isRecurring" type="checkbox" />
            <span>Повторяющееся мероприятие</span>
          </label>
          <div v-if="isRecurring" class="recurring-grid">
            <label>
              <span>Периодичность</span>
              <select v-model="recurringFrequency">
                <option value="weekly">Еженедельно</option>
                <option value="monthly">Ежемесячно</option>
                <option value="daily">Ежедневно</option>
              </select>
            </label>
            <label>
              <span>Количество</span>
              <input v-model.number="recurringCount" type="number" min="1" max="24" />
            </label>
          </div>
          <label class="toggle-field inline-toggle">
            <input v-model="saveAsTemplate" type="checkbox" />
            <span>Сохранить как шаблон</span>
          </label>
          <input v-if="saveAsTemplate" v-model="templateName" placeholder="Название шаблона" />
        </div>
        <div class="form-actions">
          <button class="secondary-action" type="button" @click="closeModal">Отмена</button>
          <button class="primary-action" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  </section>
</template>
