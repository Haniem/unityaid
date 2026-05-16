<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Eye, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { createTask, deleteTask, fetchTasks, updateTask } from '../entities/tasks/api'
import type { TaskItem, TaskPayload } from '../entities/tasks/types'
import { fetchCreateForm, fetchEditForm } from '../entities/forms/api'
import type { BackendForm, FormModel } from '../entities/forms/types'
import DynamicForm from '../shared/ui/DynamicForm.vue'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { isoFromDatetimeLocal, modelFromForm, nullable, stringValue } from '../shared/forms'
import { formatDateTime } from '../shared/date'
import { useClientPagination } from '../shared/pagination'

const items = ref<TaskItem[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const statusFilter = ref('')
const priorityFilter = ref('')
const formSchema = ref<BackendForm | null>(null)
const formModel = ref<FormModel>({})
const { page, perPage, pageItems } = useClientPagination(items, 12)

const statusOptions = [
  { id: '', name: 'Все статусы' },
  { id: 'created', name: 'Создана' },
  { id: 'assigned', name: 'Назначена' },
  { id: 'in_progress', name: 'В работе' },
  { id: 'review', name: 'На проверке' },
  { id: 'completed', name: 'Выполнена' }
]
const priorityOptions = [
  { id: '', name: 'Все приоритеты' },
  { id: 'low', name: 'Низкий' },
  { id: 'medium', name: 'Средний' },
  { id: 'high', name: 'Высокий' }
]

function resetForm() {
  editingId.value = null
  formSchema.value = null
  formModel.value = {}
  errorMessage.value = ''
}

async function openCreateModal() {
  resetForm()
  formSchema.value = await fetchCreateForm('tasks')
  formModel.value = modelFromForm(formSchema.value)
  isModalOpen.value = true
}

async function openEditModal(item: TaskItem) {
  editingId.value = item.id
  formSchema.value = await fetchEditForm('tasks', item.id)
  formModel.value = modelFromForm(formSchema.value)
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

function payload(): TaskPayload {
  return {
    organizationId: stringValue(formModel.value.organizationId),
    eventId: nullable(formModel.value.eventId) as string | null,
    title: stringValue(formModel.value.title),
    description: stringValue(formModel.value.description),
    status: (stringValue(formModel.value.status) || 'created') as TaskPayload['status'],
    priority: (stringValue(formModel.value.priority) || 'medium') as TaskPayload['priority'],
    dueAt: isoFromDatetimeLocal(formModel.value.dueAt)
  }
}

async function load() {
  items.value = (await fetchTasks({ status: statusFilter.value, priority: priorityFilter.value })).items
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
        <p class="eyebrow">Раздел</p>
        <h1>Мои задачи</h1>
      </div>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать</span>
      </button>
    </div>

    <div class="filter-bar">
      <label class="filter-select">
        <span>Статус</span>
        <CustomSelect v-model="statusFilter" :options="statusOptions" @update:model-value="load" />
      </label>
      <label class="filter-select">
        <span>Приоритет</span>
        <CustomSelect v-model="priorityFilter" :options="priorityOptions" @update:model-value="load" />
      </label>
    </div>

    <div class="task-board">
      <article v-for="item in pageItems" :key="item.id" class="task-card">
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
    <PaginationBar v-model:page="page" :per-page="perPage" :total="items.length" />

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ formSchema?.meta.title || (editingId ? 'Редактирование задачи' : 'Новая задача') }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <DynamicForm v-if="formSchema" v-model="formModel" :form="formSchema" />
        <div class="form-actions">
          <button class="secondary-action" type="button" @click="closeModal">Отмена</button>
          <button class="primary-action" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  </section>
</template>
