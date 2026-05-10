<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { createOrganization, deleteOrganization, fetchOrganizations, updateOrganization } from '../entities/organizations/api'
import type { Organization, OrganizationPayload } from '../entities/organizations/types'

const items = ref<Organization[]>([])
const editingId = ref<string | null>(null)
const errorMessage = ref('')
const form = reactive<OrganizationPayload>({
  name: '',
  slug: '',
  description: '',
  contactEmail: ''
})

function resetForm() {
  editingId.value = null
  form.name = ''
  form.slug = ''
  form.description = ''
  form.contactEmail = ''
}

function edit(item: Organization) {
  editingId.value = item.id
  form.name = item.name
  form.slug = item.slug
  form.description = item.description
  form.contactEmail = item.contactEmail ?? ''
}

async function load() {
  const response = await fetchOrganizations()
  items.value = response.items
}

async function submit() {
  errorMessage.value = ''
  const payload = { ...form, contactEmail: form.contactEmail || null }
  try {
    if (editingId.value) {
      await updateOrganization(editingId.value, payload)
    } else {
      await createOrganization(payload)
    }
    resetForm()
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить организацию'
  }
}

async function remove(item: Organization) {
  if (!confirm(`Удалить организацию "${item.name}"? Связанные данные тоже могут быть удалены.`)) {
    return
  }
  await deleteOrganization(item.id)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Организации</p>
        <h1>Организации</h1>
      </div>
    </div>

    <div class="management-grid">
      <form class="entity-form" @submit.prevent="submit">
        <h2>{{ editingId ? 'Редактирование' : 'Новая организация' }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <label><span>Название</span><input v-model="form.name" required /></label>
        <label><span>Slug</span><input v-model="form.slug" placeholder="auto или dobrye-ruki" /></label>
        <label><span>Email</span><input v-model="form.contactEmail" type="email" /></label>
        <label><span>Описание</span><textarea v-model="form.description" rows="5" /></label>
        <div class="form-actions">
          <button class="secondary-action" type="button" @click="resetForm">Сбросить</button>
          <button class="primary-action" type="submit"><Plus :size="17" /> Сохранить</button>
        </div>
      </form>

      <div class="entity-list">
        <article v-for="item in items" :key="item.id" class="entity-row">
          <div>
            <h2>{{ item.name }}</h2>
            <p>{{ item.description || 'Описание пока не заполнено.' }}</p>
            <small>{{ item.contactEmail || 'Контактный email не указан' }}</small>
          </div>
          <div class="card-actions static-actions">
            <button class="icon-button" type="button" @click="edit(item)"><Pencil :size="17" /></button>
            <button class="icon-button" type="button" @click="remove(item)"><Trash2 :size="17" /></button>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
