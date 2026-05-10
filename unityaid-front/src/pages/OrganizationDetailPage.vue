<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Globe, Mail, MapPin, Phone, Trash2 } from 'lucide-vue-next'
import {
  addOrganizationMember,
  fetchOrganization,
  fetchOrganizationMembers,
  removeOrganizationMember,
  updateOrganization,
  updateOrganizationMember
} from '../entities/organizations/api'
import type {
  AddOrganizationMemberPayload,
  Organization,
  OrganizationMember,
  OrganizationPayload
} from '../entities/organizations/types'

const route = useRoute()
const organizationId = computed(() => String(route.params.id))
const item = ref<Organization | null>(null)
const members = ref<OrganizationMember[]>([])
const errorMessage = ref('')
const memberError = ref('')
const settingsMessage = ref('')
const settingsError = ref('')

const memberForm = reactive<AddOrganizationMemberPayload>({
  email: '',
  role: 'volunteer',
  status: 'active'
})

const settingsForm = reactive<OrganizationPayload>({
  name: '',
  slug: '',
  description: '',
  contactEmail: '',
  logoUrl: '',
  websiteUrl: '',
  phone: '',
  address: ''
})

const roleOptions = [
  { value: 'super_admin', label: 'Super admin' },
  { value: 'org_admin', label: 'Администратор' },
  { value: 'coordinator', label: 'Координатор' },
  { value: 'volunteer', label: 'Волонтер' }
] as const

const statusOptions = [
  { value: 'active', label: 'Активен' },
  { value: 'inactive', label: 'Неактивен' },
  { value: 'blocked', label: 'Заблокирован' }
] as const

function syncSettingsForm(organization: Organization) {
  settingsForm.name = organization.name
  settingsForm.slug = organization.slug
  settingsForm.description = organization.description
  settingsForm.contactEmail = organization.contactEmail ?? ''
  settingsForm.logoUrl = organization.logoUrl ?? ''
  settingsForm.websiteUrl = organization.websiteUrl ?? ''
  settingsForm.phone = organization.phone ?? ''
  settingsForm.address = organization.address ?? ''
}

async function load() {
  try {
    const [organizationResponse, membersResponse] = await Promise.all([
      fetchOrganization(organizationId.value),
      fetchOrganizationMembers(organizationId.value)
    ])
    item.value = organizationResponse.item
    members.value = membersResponse.items
    syncSettingsForm(organizationResponse.item)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить организацию'
  }
}

function nullable(value: string | null | undefined) {
  return value?.trim() || null
}

async function saveSettings() {
  if (!item.value) return
  settingsError.value = ''
  settingsMessage.value = ''
  try {
    const response = await updateOrganization(item.value.id, {
      ...settingsForm,
      contactEmail: nullable(settingsForm.contactEmail),
      logoUrl: nullable(settingsForm.logoUrl),
      websiteUrl: nullable(settingsForm.websiteUrl),
      phone: nullable(settingsForm.phone),
      address: nullable(settingsForm.address)
    })
    item.value = response.item
    syncSettingsForm(response.item)
    settingsMessage.value = 'Настройки сохранены'
  } catch (error) {
    settingsError.value = error instanceof Error ? error.message : 'Не удалось сохранить настройки'
  }
}

async function addMember() {
  memberError.value = ''
  try {
    await addOrganizationMember(organizationId.value, memberForm)
    memberForm.email = ''
    memberForm.role = 'volunteer'
    memberForm.status = 'active'
    members.value = (await fetchOrganizationMembers(organizationId.value)).items
  } catch (error) {
    memberError.value = error instanceof Error ? error.message : 'Не удалось добавить участника'
  }
}

async function updateMember(member: OrganizationMember) {
  memberError.value = ''
  try {
    const response = await updateOrganizationMember(organizationId.value, member.id, {
      role: member.role,
      status: member.status
    })
    const index = members.value.findIndex((item) => item.id === member.id)
    if (index >= 0) members.value[index] = response.item
  } catch (error) {
    memberError.value = error instanceof Error ? error.message : 'Не удалось обновить участника'
  }
}

async function removeMember(member: OrganizationMember) {
  if (!confirm(`Удалить ${member.email} из организации?`)) return
  await removeOrganizationMember(organizationId.value, member.id)
  members.value = members.value.filter((item) => item.id !== member.id)
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <article v-else-if="item" class="detail-layout wide">
      <div class="page-heading">
        <div class="organization-title">
          <img v-if="item.logoUrl" :src="item.logoUrl" alt="" />
          <span v-else class="organization-logo large">{{ item.name.slice(0, 1).toUpperCase() }}</span>
          <div>
            <p class="eyebrow">Организация</p>
            <h1>{{ item.name }}</h1>
          </div>
        </div>
        <RouterLink class="secondary-action" to="/organizations">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.isDeleted ? 'архив' : item.slug }}</span>
        <p class="detail-summary">{{ item.description || 'Описание организации пока не заполнено.' }}</p>
        <div class="meta-grid">
          <div class="meta-line">
            <Mail :size="16" />
            <span>{{ item.contactEmail || 'Контактный email не указан' }}</span>
          </div>
          <div class="meta-line">
            <Phone :size="16" />
            <span>{{ item.phone || 'Телефон не указан' }}</span>
          </div>
          <div class="meta-line">
            <Globe :size="16" />
            <span>{{ item.websiteUrl || 'Сайт не указан' }}</span>
          </div>
          <div class="meta-line">
            <MapPin :size="16" />
            <span>{{ item.address || 'Адрес не указан' }}</span>
          </div>
        </div>
      </div>

      <div class="organization-management">
        <section class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Участники</p>
              <h2>{{ members.length }} человек</h2>
            </div>
          </div>

          <form class="inline-member-form" @submit.prevent="addMember">
            <input v-model="memberForm.email" type="email" placeholder="email пользователя" required />
            <select v-model="memberForm.role">
              <option v-for="option in roleOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <select v-model="memberForm.status">
              <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <button class="primary-action" type="submit">Добавить</button>
          </form>
          <p v-if="memberError" class="form-error">{{ memberError }}</p>

          <div class="member-list">
            <article v-for="member in members" :key="member.id" class="member-row">
              <img :src="member.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
              <div>
                <strong>{{ member.lastName }} {{ member.firstName }}</strong>
                <small>{{ member.email }}</small>
              </div>
              <select v-model="member.role" @change="updateMember(member)">
                <option v-for="option in roleOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
              <select v-model="member.status" @change="updateMember(member)">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
              <button class="icon-button" type="button" aria-label="Удалить участника" @click="removeMember(member)">
                <Trash2 :size="17" />
              </button>
            </article>
          </div>
        </section>

        <form class="detail-panel settings-form" @submit.prevent="saveSettings">
          <div>
            <p class="eyebrow">Настройки</p>
            <h2>Профиль организации</h2>
          </div>

          <label><span>Название</span><input v-model="settingsForm.name" required /></label>
          <label><span>Slug</span><input v-model="settingsForm.slug" required /></label>
          <label><span>Логотип URL</span><input v-model="settingsForm.logoUrl" type="url" /></label>
          <label><span>Email</span><input v-model="settingsForm.contactEmail" type="email" /></label>
          <label><span>Телефон</span><input v-model="settingsForm.phone" type="tel" /></label>
          <label><span>Сайт</span><input v-model="settingsForm.websiteUrl" type="url" /></label>
          <label><span>Адрес</span><input v-model="settingsForm.address" /></label>
          <label><span>Описание</span><textarea v-model="settingsForm.description" rows="4" /></label>

          <p v-if="settingsError" class="form-error">{{ settingsError }}</p>
          <p v-if="settingsMessage" class="form-success">{{ settingsMessage }}</p>
          <button class="primary-action" type="submit">Сохранить настройки</button>
        </form>
      </div>
    </article>
  </section>
</template>
