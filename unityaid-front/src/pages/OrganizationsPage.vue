<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Eye, Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'
import { createOrganization, deleteOrganization, fetchOrganizations, updateOrganization } from '../entities/organizations/api'
import type { Organization, OrganizationPayload } from '../entities/organizations/types'
import { fetchCreateForm, fetchEditForm } from '../entities/forms/api'
import type { BackendForm, FormModel } from '../entities/forms/types'
import DynamicForm from '../shared/ui/DynamicForm.vue'
import { modelFromForm, nullable, stringValue } from '../shared/forms'

const items = ref<Organization[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const search = ref('')
const includeDeleted = ref(false)
const formSchema = ref<BackendForm | null>(null)
const formModel = ref<FormModel>({})

const activeCount = computed(() => items.value.filter((item) => !item.isDeleted).length)

function resetForm() {
  editingId.value = null
  formSchema.value = null
  formModel.value = {}
  errorMessage.value = ''
}

async function openCreateModal() {
  resetForm()
  formSchema.value = await fetchCreateForm('organizations')
  formModel.value = modelFromForm(formSchema.value)
  isModalOpen.value = true
}

async function openEditModal(item: Organization) {
  editingId.value = item.id
  formSchema.value = await fetchEditForm('organizations', item.id)
  formModel.value = modelFromForm(formSchema.value)
  errorMessage.value = ''
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

async function load() {
  const response = await fetchOrganizations({ search: search.value, includeDeleted: includeDeleted.value })
  items.value = response.items
}

function payload(): OrganizationPayload {
  return {
    name: stringValue(formModel.value.name),
    slug: stringValue(formModel.value.slug),
    description: stringValue(formModel.value.description),
    contactEmail: nullable(formModel.value.contactEmail) as string | null,
    logoUrl: nullable(formModel.value.logoUrl) as string | null,
    websiteUrl: nullable(formModel.value.websiteUrl) as string | null,
    phone: nullable(formModel.value.phone) as string | null,
    address: nullable(formModel.value.address) as string | null
  }
}

async function submit() {
  errorMessage.value = ''
  try {
    if (editingId.value) await updateOrganization(editingId.value, payload())
    else await createOrganization(payload())
    closeModal()
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить организацию'
  }
}

async function remove(item: Organization) {
  if (!confirm(`Архивировать организацию "${item.name}"? Связанные данные сохранятся.`)) return
  await deleteOrganization(item.id)
  await load()
}

let searchTimer: number | undefined
watch([search, includeDeleted], () => {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(load, 250)
})

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading with-search">
      <div>
        <p class="eyebrow">Раздел</p>
        <h1>Организации</h1>
        <p>{{ activeCount }} активных организаций</p>
      </div>
      <label class="search-field heading-search">
        <Search :size="18" />
        <input v-model="search" type="search" placeholder="Поиск по названию, slug, описанию или email" />
      </label>
      <label class="archive-toggle">
        <input v-model="includeDeleted" type="checkbox" />
        <span>Архив</span>
      </label>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать</span>
      </button>
    </div>

    <div v-if="items.length" class="directory-grid">
      <article v-for="item in items" :key="item.id" class="directory-card" :class="{ muted: item.isDeleted }">
        <div class="directory-card-main">
          <div class="organization-card-head">
            <img v-if="item.logoUrl" :src="item.logoUrl" alt="" />
            <span v-else class="organization-logo">{{ item.name.slice(0, 1).toUpperCase() }}</span>
            <span class="status-pill">{{ item.isDeleted ? 'архив' : item.slug }}</span>
          </div>
          <h2>{{ item.name }}</h2>
          <p>{{ item.description || 'Описание пока не заполнено.' }}</p>
          <small>{{ item.contactEmail || 'Контактный email не указан' }}</small>
        </div>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/organizations/${item.id}`" aria-label="Открыть"><Eye :size="17" /></RouterLink>
          <button class="icon-button" type="button" aria-label="Редактировать" :disabled="item.isDeleted" @click="openEditModal(item)">
            <Pencil :size="17" />
          </button>
          <button class="icon-button" type="button" aria-label="Архивировать" :disabled="item.isDeleted" @click="remove(item)">
            <Trash2 :size="17" />
          </button>
        </div>
      </article>
    </div>

    <div v-else class="empty-state">Организации не найдены.</div>

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ formSchema?.meta.title || (editingId ? 'Редактирование организации' : 'Новая организация') }}</h2>
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
