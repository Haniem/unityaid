<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Eye, Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'
import { createOrganization, deleteOrganization, fetchOrganizations, updateOrganization } from '../entities/organizations/api'
import type { Organization, OrganizationPayload } from '../entities/organizations/types'

const items = ref<Organization[]>([])
const editingId = ref<string | null>(null)
const isModalOpen = ref(false)
const errorMessage = ref('')
const search = ref('')
const includeDeleted = ref(false)
const form = reactive<OrganizationPayload>({
  name: '',
  slug: '',
  description: '',
  contactEmail: '',
  logoUrl: '',
  websiteUrl: '',
  phone: '',
  address: ''
})

const activeCount = computed(() => items.value.filter((item) => !item.isDeleted).length)

function resetForm() {
  editingId.value = null
  form.name = ''
  form.slug = ''
  form.description = ''
  form.contactEmail = ''
  form.logoUrl = ''
  form.websiteUrl = ''
  form.phone = ''
  form.address = ''
  errorMessage.value = ''
}

function openCreateModal() {
  resetForm()
  isModalOpen.value = true
}

function openEditModal(item: Organization) {
  editingId.value = item.id
  form.name = item.name
  form.slug = item.slug
  form.description = item.description
  form.contactEmail = item.contactEmail ?? ''
  form.logoUrl = item.logoUrl ?? ''
  form.websiteUrl = item.websiteUrl ?? ''
  form.phone = item.phone ?? ''
  form.address = item.address ?? ''
  errorMessage.value = ''
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  resetForm()
}

async function load() {
  const response = await fetchOrganizations({
    search: search.value,
    includeDeleted: includeDeleted.value
  })
  items.value = response.items
}

function nullable(value: string | null | undefined) {
  return value?.trim() || null
}

async function submit() {
  errorMessage.value = ''
  const payload = {
    ...form,
    contactEmail: nullable(form.contactEmail),
    logoUrl: nullable(form.logoUrl),
    websiteUrl: nullable(form.websiteUrl),
    phone: nullable(form.phone),
    address: nullable(form.address)
  }
  try {
    if (editingId.value) await updateOrganization(editingId.value, payload)
    else await createOrganization(payload)
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
    <div class="page-heading">
      <div>
        <p class="eyebrow">Организации</p>
        <h1>Организации</h1>
        <p>{{ activeCount }} активных организаций</p>
      </div>
      <button class="primary-action" type="button" @click="openCreateModal">
        <Plus :size="18" />
        <span>Создать организацию</span>
      </button>
    </div>

    <div class="filter-bar">
      <label class="search-field">
        <Search :size="18" />
        <input v-model="search" type="search" placeholder="Поиск по названию, slug, описанию или email" />
      </label>
      <label class="toggle-field">
        <input v-model="includeDeleted" type="checkbox" />
        <span>Показывать архив</span>
      </label>
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
        <h2>{{ editingId ? 'Редактирование организации' : 'Новая организация' }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <div class="form-columns">
          <label><span>Название</span><input v-model="form.name" required /></label>
          <label><span>Slug</span><input v-model="form.slug" placeholder="auto или dobrye-ruki" /></label>
          <label><span>Email</span><input v-model="form.contactEmail" type="email" /></label>
          <label><span>Телефон</span><input v-model="form.phone" type="tel" /></label>
          <label><span>Сайт</span><input v-model="form.websiteUrl" type="url" /></label>
          <label><span>Логотип URL</span><input v-model="form.logoUrl" type="url" /></label>
        </div>
        <label><span>Адрес</span><input v-model="form.address" /></label>
        <label><span>Описание</span><textarea v-model="form.description" rows="5" /></label>
        <div class="form-actions">
          <button class="secondary-action" type="button" @click="closeModal">Отмена</button>
          <button class="primary-action" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  </section>
</template>
