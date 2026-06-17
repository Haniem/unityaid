<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Database, Edit3, Plus, RefreshCcw, Save, Search, Trash2 } from 'lucide-vue-next'
import {
  createAdminRow,
  deleteAdminRow,
  fetchAdminEntities,
  fetchAdminRows,
  updateAdminRow
} from '../entities/admin/api'
import type { AdminEntityConfig, AdminEntityRow } from '../entities/admin/types'

const entities = ref<AdminEntityConfig[]>([])
const selectedEntityCode = ref('')
const rows = ref<AdminEntityRow[]>([])
const page = ref(1)
const perPage = 20
const total = ref(0)
const search = ref('')
const editorMode = ref<'create' | 'edit'>('create')
const editorId = ref('')
const editorText = ref('{}')
const message = ref('')
const isLoading = ref(false)
const isSaving = ref(false)

const selectedEntity = computed(() => entities.value.find((item) => item.code === selectedEntityCode.value) || null)
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / perPage)))

async function loadEntities() {
  const response = await fetchAdminEntities()
  entities.value = response.items
  if (!selectedEntityCode.value && entities.value.length) {
    selectedEntityCode.value = entities.value[0].code
  }
}

async function loadRows() {
  if (!selectedEntityCode.value) return
  isLoading.value = true
  message.value = ''
  try {
    const response = await fetchAdminRows(selectedEntityCode.value, {
      page: page.value,
      perPage,
      search: search.value.trim() || undefined
    })
    rows.value = response.items
    total.value = response.total
    if (editorMode.value === 'create') prepareCreate()
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось загрузить данные'
  } finally {
    isLoading.value = false
  }
}

function editableShape() {
  const shape: AdminEntityRow = {}
  selectedEntity.value?.editable
    .filter((column) => column !== 'updated_at')
    .forEach((column) => {
      shape[column] = ''
    })
  return shape
}

function prepareCreate() {
  editorMode.value = 'create'
  editorId.value = ''
  editorText.value = JSON.stringify(editableShape(), null, 2)
}

function editRow(row: AdminEntityRow) {
  if (!selectedEntity.value) return
  const id = String(row[selectedEntity.value.primaryKey] ?? '')
  const data: AdminEntityRow = {}
  selectedEntity.value.editable
    .filter((column) => column !== 'updated_at')
    .forEach((column) => {
      data[column] = row[column] ?? null
    })
  editorMode.value = 'edit'
  editorId.value = id
  editorText.value = JSON.stringify(data, null, 2)
}

async function save() {
  if (!selectedEntity.value) return
  isSaving.value = true
  message.value = ''
  try {
    const data = JSON.parse(editorText.value) as AdminEntityRow
    if (editorMode.value === 'edit') {
      await updateAdminRow(selectedEntity.value.code, editorId.value, data)
      message.value = 'Запись обновлена'
    } else {
      await createAdminRow(selectedEntity.value.code, data)
      message.value = 'Запись создана'
    }
    await loadRows()
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось сохранить запись'
  } finally {
    isSaving.value = false
  }
}

async function remove(row: AdminEntityRow) {
  if (!selectedEntity.value || !selectedEntity.value.canDelete) return
  const id = String(row[selectedEntity.value.primaryKey] ?? '')
  if (!id || !confirm('Удалить запись?')) return
  try {
    await deleteAdminRow(selectedEntity.value.code, id)
    message.value = 'Запись удалена'
    await loadRows()
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось удалить запись'
  }
}

function rowId(row: AdminEntityRow) {
  return selectedEntity.value ? String(row[selectedEntity.value.primaryKey] ?? '') : ''
}

function cellValue(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function columnLabel(column: string) {
  if (column === 'display_id') return '№'
  return column
}

watch(selectedEntityCode, () => {
  page.value = 1
  search.value = ''
  prepareCreate()
  loadRows()
})

onMounted(async () => {
  await loadEntities()
  await loadRows()
})
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Администрирование</p>
        <h1>Системная админ-панель</h1>
        <p>Постраничное управление сущностями приложения: просмотр, создание, редактирование и удаление.</p>
      </div>
      <button class="secondary-action" type="button" @click="loadRows">
        <RefreshCcw :size="17" />
        <span>Обновить</span>
      </button>
    </div>

    <p v-if="message" :class="message.includes('Не удалось') ? 'form-error' : 'form-success'">{{ message }}</p>

    <div class="admin-layout">
      <aside class="detail-panel admin-entities">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Сущности</p>
            <h2>Разделы</h2>
          </div>
          <Database :size="22" />
        </div>
        <button
          v-for="entity in entities"
          :key="entity.code"
          :class="['admin-entity-button', { active: entity.code === selectedEntityCode }]"
          type="button"
          @click="selectedEntityCode = entity.code"
        >
          <strong>{{ entity.label }}</strong>
          <small>{{ entity.code }}</small>
        </button>
      </aside>

      <section class="admin-content">
        <div class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">{{ selectedEntity?.code }}</p>
              <h2>{{ selectedEntity?.label }}</h2>
              <p v-if="selectedEntity?.description">{{ selectedEntity.description }}</p>
            </div>
          </div>

          <div class="filter-bar">
            <label class="search-field">
              <Search :size="17" />
              <input v-model="search" type="search" placeholder="Поиск" @keyup.enter="page = 1; loadRows()" />
            </label>
            <button class="secondary-action" type="button" @click="page = 1; loadRows()">Найти</button>
          </div>

          <div class="admin-table-wrap">
            <table class="admin-table">
              <thead>
                <tr>
                  <th v-for="column in selectedEntity?.columns" :key="column">{{ columnLabel(column) }}</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in rows" :key="rowId(row)">
                  <td v-for="column in selectedEntity?.columns" :key="column">{{ cellValue(row[column]) }}</td>
                  <td>
                    <button class="icon-button" type="button" aria-label="Редактировать" @click="editRow(row)">
                      <Edit3 :size="16" />
                    </button>
                    <button
                      v-if="selectedEntity?.canDelete"
                      class="icon-button danger-action"
                      type="button"
                      aria-label="Удалить"
                      @click="remove(row)"
                    >
                      <Trash2 :size="16" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <p v-if="!isLoading && rows.length === 0" class="empty-state">Записей не найдено.</p>

          <div class="admin-pagination">
            <button class="secondary-action" type="button" :disabled="page <= 1" @click="page--; loadRows()">Назад</button>
            <span>{{ page }} / {{ pageCount }} · {{ total }}</span>
            <button class="secondary-action" type="button" :disabled="page >= pageCount" @click="page++; loadRows()">Вперед</button>
          </div>
        </div>

        <form class="detail-panel settings-form admin-editor" @submit.prevent="save">
          <div class="section-heading">
            <div>
              <p class="eyebrow">{{ editorMode === 'edit' ? 'Редактирование' : 'Создание' }}</p>
              <h2>{{ editorMode === 'edit' ? editorId : 'Новая запись' }}</h2>
            </div>
            <button v-if="selectedEntity?.canCreate" class="secondary-action" type="button" @click="prepareCreate">
              <Plus :size="17" />
              <span>Новая</span>
            </button>
          </div>

          <textarea v-model="editorText" class="admin-json-editor" rows="18" spellcheck="false"></textarea>

          <button
            class="primary-action"
            type="submit"
            :disabled="isSaving || (editorMode === 'create' && !selectedEntity?.canCreate)"
          >
            <Save :size="17" />
            <span>{{ isSaving ? 'Сохранение...' : 'Сохранить JSON' }}</span>
          </button>
        </form>
      </section>
    </div>
  </section>
</template>
