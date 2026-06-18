<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Copy, Filter, Globe, Link2, Mail, MapPin, Phone, Trash2, X } from 'lucide-vue-next'
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
import { fetchEvents } from '../entities/events/api'
import type { EventItem } from '../entities/events/types'
import { fetchTasks } from '../entities/tasks/api'
import type { TaskItem } from '../entities/tasks/types'
import { fetchVolunteers } from '../entities/users/api'
import { createInvitation } from '../entities/invitations/api'
import type { Invitation } from '../entities/invitations/types'
import type { PossibleValue } from '../entities/forms/types'
import { authState } from '../entities/auth/store'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import AsyncSelect from '../shared/ui/AsyncSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'
import { formatDateTime } from '../shared/date'
import { canManageOrganizations } from '../shared/permissions'

const route = useRoute()
const organizationId = computed(() => String(route.params.id))
const item = ref<Organization | null>(null)
const members = ref<OrganizationMember[]>([])
const events = ref<EventItem[]>([])
const tasks = ref<TaskItem[]>([])
const errorMessage = ref('')
const memberError = ref('')
const settingsMessage = ref('')
const settingsError = ref('')
const inviteMessage = ref('')
const inviteError = ref('')
const generatedInvite = ref<Invitation | null>(null)
const activeOrgTab = ref<'members' | 'events'>('members')
const isSettingsModalOpen = ref(false)
const isEventFiltersOpen = ref(false)
const { page, perPage, pageItems } = useClientPagination(members, 10)
const eventFilters = reactive({ startsAt: '', endsAt: '', executorId: '' })
const canManageOrganizationDetails = computed(() => canManageOrganizations(authState.user))

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
  { id: 'super_admin', name: 'Super admin' },
  { id: 'org_admin', name: 'Администратор' },
  { id: 'coordinator', name: 'Координатор' },
  { id: 'volunteer', name: 'Волонтер' }
] as const

const statusOptions = [
  { id: 'active', name: 'Активен' },
  { id: 'inactive', name: 'Неактивен' },
  { id: 'blocked', name: 'Заблокирован' }
] as const

const eventStatusLabels: Record<string, string> = {
  draft: 'Черновик',
  published: 'Опубликовано',
  completed: 'Завершено',
  cancelled: 'Отменено'
}

const filteredEvents = computed(() => {
  const startsAt = eventFilters.startsAt ? new Date(eventFilters.startsAt) : null
  const endsAt = eventFilters.endsAt ? new Date(`${eventFilters.endsAt}T23:59:59`) : null

  return events.value.filter((event) => {
    if (event.organizationId !== organizationId.value) return false
    const eventStart = new Date(event.startsAt)
    const eventEnd = new Date(event.endsAt)
    if (startsAt && eventEnd < startsAt) return false
    if (endsAt && eventStart > endsAt) return false
    if (eventFilters.executorId) {
      const hasExecutorTask = tasks.value.some(
        (task) => task.eventId === event.id && task.assignees.some((assignee) => assignee.userId === eventFilters.executorId)
      )
      if (!hasExecutorTask) return false
    }
    return true
  })
})

const activeEventFiltersCount = computed(() =>
  [eventFilters.startsAt, eventFilters.endsAt, eventFilters.executorId].filter(Boolean).length
)

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
    const [organizationResponse, membersResponse, eventsResponse, tasksResponse] = await Promise.all([
      fetchOrganization(organizationId.value),
      fetchOrganizationMembers(organizationId.value),
      fetchEvents(),
      fetchTasks()
    ])
    item.value = organizationResponse.item
    members.value = membersResponse.items
    events.value = eventsResponse.items
    tasks.value = tasksResponse.items
    syncSettingsForm(organizationResponse.item)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить организацию'
  }
}

function openSettingsModal() {
  if (!canManageOrganizationDetails.value) return
  if (!item.value) return
  settingsMessage.value = ''
  settingsError.value = ''
  syncSettingsForm(item.value)
  isSettingsModalOpen.value = true
}

function nullable(value: string | null | undefined) {
  return value?.trim() || null
}

async function saveSettings() {
  if (!canManageOrganizationDetails.value) return
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
    isSettingsModalOpen.value = false
  } catch (error) {
    settingsError.value = error instanceof Error ? error.message : 'Не удалось сохранить настройки'
  }
}

async function addMember() {
  if (!canManageOrganizationDetails.value) return
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

function absoluteInviteLink(invite: Invitation) {
  return `${window.location.origin}${invite.link}`
}

function initials(firstName?: string, lastName?: string) {
  return `${firstName?.[0] ?? ''}${lastName?.[0] ?? ''}`.toUpperCase() || 'П'
}

async function loadExecutorOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchVolunteers({ search: params.search })
  const items: PossibleValue[] = response.items
    .slice((params.page - 1) * params.perPage, params.page * params.perPage)
    .map((user) => ({ id: user.userId, name: `${user.lastName} ${user.firstName} · ${user.email}` }))
  return { items }
}

function resetEventFilters() {
  eventFilters.startsAt = ''
  eventFilters.endsAt = ''
  eventFilters.executorId = ''
}

async function generateVolunteerInvite() {
  if (!canManageOrganizationDetails.value) return
  if (!item.value) return
  inviteMessage.value = ''
  inviteError.value = ''
  try {
    const response = await createInvitation({
      role: 'volunteer',
      organizationId: item.value.id
    })
    generatedInvite.value = response.item
    inviteMessage.value = 'Одноразовая ссылка создана'
  } catch (error) {
    inviteError.value = error instanceof Error ? error.message : 'Не удалось создать ссылку'
  }
}

async function copyInviteLink() {
  if (!generatedInvite.value) return
  await navigator.clipboard.writeText(absoluteInviteLink(generatedInvite.value))
  inviteMessage.value = 'Ссылка скопирована'
}

async function updateMember(member: OrganizationMember) {
  if (!canManageOrganizationDetails.value) return
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
  if (!canManageOrganizationDetails.value) return
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
        <div class="page-actions">
          <button v-if="canManageOrganizationDetails" class="secondary-action" type="button" @click="generateVolunteerInvite"><Link2 :size="17" /> Ссылка для волонтера</button>
          <button v-if="canManageOrganizationDetails" class="secondary-action" type="button" @click="openSettingsModal">Редактировать организацию</button>
          <RouterLink class="secondary-action" to="/organizations">К списку</RouterLink>
        </div>
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

      <section v-if="canManageOrganizationDetails && (generatedInvite || inviteMessage || inviteError)" class="detail-panel invite-link-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Регистрация волонтера</p>
            <h2>Одноразовая ссылка</h2>
          </div>
          <button v-if="generatedInvite" class="secondary-action" type="button" @click="copyInviteLink"><Copy :size="16" /> Скопировать</button>
        </div>
        <p v-if="generatedInvite" class="copy-line"><Link2 :size="14" /> {{ absoluteInviteLink(generatedInvite) }}</p>
        <p v-if="inviteMessage" class="form-success">{{ inviteMessage }}</p>
        <p v-if="inviteError" class="form-error">{{ inviteError }}</p>
      </section>

      <div class="tabs detail-tabs">
        <button :class="{ active: activeOrgTab === 'members' }" type="button" @click="activeOrgTab = 'members'">Сотрудники</button>
        <button :class="{ active: activeOrgTab === 'events' }" type="button" @click="activeOrgTab = 'events'">Мероприятия</button>
      </div>

      <div v-if="activeOrgTab === 'members'" class="organization-management">
        <section class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Участники</p>
              <h2>{{ members.length }} человек</h2>
            </div>
          </div>

          <form v-if="canManageOrganizationDetails" class="inline-member-form" @submit.prevent="addMember">
            <input v-model="memberForm.email" type="email" placeholder="email пользователя" required />
            <CustomSelect v-model="memberForm.role" :options="roleOptions" />
            <CustomSelect v-model="memberForm.status" :options="statusOptions" />
            <button class="primary-action" type="submit">Добавить</button>
          </form>
          <p v-if="memberError" class="form-error">{{ memberError }}</p>

          <div class="member-list">
            <article v-for="member in pageItems" :key="member.id" class="member-row">
              <img v-if="member.avatarUrl" :src="member.avatarUrl" alt="" />
              <span v-else class="mini-avatar">{{ initials(member.firstName, member.lastName) }}</span>
              <div>
                <strong>{{ member.lastName }} {{ member.firstName }}</strong>
                <small>{{ member.email }}</small>
              </div>
              <CustomSelect v-if="canManageOrganizationDetails" v-model="member.role" :options="roleOptions" @update:model-value="updateMember(member)" />
              <span v-else class="status-pill">{{ roleOptions.find((role) => role.id === member.role)?.name || member.role }}</span>
              <CustomSelect v-if="canManageOrganizationDetails" v-model="member.status" :options="statusOptions" @update:model-value="updateMember(member)" />
              <span v-else class="status-pill">{{ statusOptions.find((status) => status.id === member.status)?.name || member.status }}</span>
              <button v-if="canManageOrganizationDetails" class="icon-button" type="button" aria-label="Удалить участника" @click="removeMember(member)">
                <Trash2 :size="17" />
              </button>
            </article>
          </div>
          <PaginationBar v-model:page="page" :per-page="perPage" :total="members.length" />
        </section>
      </div>

      <section v-if="activeOrgTab === 'events'" class="detail-panel organization-events-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Мероприятия</p>
            <h2>{{ filteredEvents.length }} мероприятий</h2>
          </div>
          <button class="secondary-action" type="button" @click="isEventFiltersOpen = true">
            <Filter :size="16" />
            Фильтры
            <span v-if="activeEventFiltersCount" class="button-counter">{{ activeEventFiltersCount }}</span>
          </button>
        </div>
        <div class="member-list">
          <article v-for="event in filteredEvents" :key="event.id" class="member-row linked-row">
            <div>
              <RouterLink :to="`/calendar/${event.id}`"><strong>{{ event.title }}</strong></RouterLink>
              <small>{{ formatDateTime(event.startsAt) }} - {{ formatDateTime(event.endsAt) }} · {{ event.location || 'место не указано' }}</small>
            </div>
            <span class="status-pill">{{ eventStatusLabels[event.status] || event.status }}</span>
          </article>
          <p v-if="filteredEvents.length === 0" class="empty-state">Мероприятий по выбранным фильтрам нет.</p>
        </div>
      </section>

      <Teleport to="body">
        <div v-if="isEventFiltersOpen" class="modal-backdrop" @click.self="isEventFiltersOpen = false">
          <section class="modal-panel entity-form filters-modal">
            <div class="modal-heading">
              <div>
                <p class="eyebrow">Мероприятия</p>
                <h2>Фильтры</h2>
              </div>
              <button class="icon-button" type="button" aria-label="Закрыть" @click="isEventFiltersOpen = false">
                <X :size="18" />
              </button>
            </div>
            <div class="settings-form filters-form">
              <label><span>С даты</span><input v-model="eventFilters.startsAt" type="date" /></label>
              <label><span>По дату</span><input v-model="eventFilters.endsAt" type="date" /></label>
              <label class="settings-form-wide">
                <span>Исполнитель</span>
                <AsyncSelect v-model="eventFilters.executorId" :load-options="loadExecutorOptions" placeholder="Найти пользователя" />
              </label>
            </div>
            <div class="form-actions">
              <button class="secondary-action" type="button" @click="resetEventFilters">Сбросить</button>
              <button class="primary-action" type="button" @click="isEventFiltersOpen = false">Показать</button>
            </div>
          </section>
        </div>
      </Teleport>

      <Teleport to="body">
        <div v-if="isSettingsModalOpen" class="modal-backdrop" @click.self="isSettingsModalOpen = false">
          <form class="modal-panel entity-form settings-form organization-settings-modal" @submit.prevent="saveSettings">
            <div class="modal-heading">
              <div>
                <p class="eyebrow">Настройки</p>
                <h2>Профиль организации</h2>
              </div>
              <button class="icon-button" type="button" aria-label="Закрыть" @click="isSettingsModalOpen = false">×</button>
            </div>

            <div class="settings-form-grid">
              <label><span>Название</span><input v-model="settingsForm.name" required /></label>
              <label><span>Slug</span><input v-model="settingsForm.slug" required /></label>
              <label><span>Логотип URL</span><input v-model="settingsForm.logoUrl" type="url" /></label>
              <label><span>Email</span><input v-model="settingsForm.contactEmail" type="email" /></label>
              <label><span>Телефон</span><input v-model="settingsForm.phone" type="tel" /></label>
              <label><span>Сайт</span><input v-model="settingsForm.websiteUrl" type="url" /></label>
              <label class="settings-form-wide"><span>Адрес</span><input v-model="settingsForm.address" /></label>
              <label class="settings-form-wide"><span>Описание</span><textarea v-model="settingsForm.description" rows="4" /></label>
            </div>

            <p v-if="settingsError" class="form-error">{{ settingsError }}</p>
            <p v-if="settingsMessage" class="form-success">{{ settingsMessage }}</p>
            <div class="form-actions">
              <button class="secondary-action" type="button" @click="isSettingsModalOpen = false">Отмена</button>
              <button class="primary-action" type="submit">Сохранить настройки</button>
            </div>
          </form>
        </div>
      </Teleport>
    </article>
  </section>
</template>
