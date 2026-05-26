<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Building2, MailPlus, Palette, Save, Trash2 } from 'lucide-vue-next'
import { fetchTenantSettings, updateTenantSettings } from '../entities/tenantSettings/api'
import type { PendingInvite, TenantSettingsPayload } from '../entities/tenantSettings/types'

const router = useRouter()
const isLoading = ref(true)
const isSaving = ref(false)
const loadError = ref('')
const saveMessage = ref('')
const inviteEmail = ref('')
const inviteRole = ref<PendingInvite['role']>('coordinator')
const form = ref<TenantSettingsPayload>({
  displayName: 'Пульс',
  description: '',
  logoUrl: null,
  primaryColor: '#2f9f72',
  accentColor: '#22684e',
  timezone: 'Asia/Yekaterinburg',
  locale: 'ru',
  contactEmail: null,
  contactPhone: null,
  defaultOrganizationId: null,
  defaultOrganizationName: '',
  defaultOrganizationSlug: '',
  onboardingCompleted: false,
  pendingInvites: []
})

const canFinish = computed(() => {
  return form.value.displayName.trim().length >= 2 && form.value.defaultOrganizationName.trim().length >= 2
})

function makeSlug(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, '-')
    .replace(/^-+|-+$/g, '') || 'main'
}

function syncOrganizationSlug() {
  if (!form.value.defaultOrganizationSlug.trim()) {
    form.value.defaultOrganizationSlug = makeSlug(form.value.defaultOrganizationName)
  }
}

function addInvite() {
  const email = inviteEmail.value.trim().toLowerCase()
  if (!email) return
  if (!form.value.pendingInvites.some((item) => item.email === email)) {
    form.value.pendingInvites.push({ email, role: inviteRole.value })
  }
  inviteEmail.value = ''
}

function removeInvite(email: string) {
  form.value.pendingInvites = form.value.pendingInvites.filter((item) => item.email !== email)
}

async function load() {
  isLoading.value = true
  loadError.value = ''
  try {
    const response = await fetchTenantSettings()
    form.value = {
      displayName: response.item.displayName,
      description: response.item.description,
      logoUrl: response.item.logoUrl ?? null,
      primaryColor: response.item.primaryColor,
      accentColor: response.item.accentColor,
      timezone: response.item.timezone,
      locale: response.item.locale,
      contactEmail: response.item.contactEmail ?? null,
      contactPhone: response.item.contactPhone ?? null,
      defaultOrganizationId: response.item.defaultOrganizationId ?? null,
      defaultOrganizationName: response.item.defaultOrganizationName,
      defaultOrganizationSlug: response.item.defaultOrganizationSlug,
      onboardingCompleted: response.item.onboardingCompleted,
      pendingInvites: response.item.pendingInvites ?? []
    }
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : 'Не удалось загрузить настройки клиента'
  } finally {
    isLoading.value = false
  }
}

async function save(complete: boolean) {
  if (complete && !canFinish.value) {
    saveMessage.value = 'Заполните название пространства и первое подразделение'
    return
  }
  syncOrganizationSlug()
  isSaving.value = true
  saveMessage.value = ''
  try {
    const response = await updateTenantSettings({
      ...form.value,
      onboardingCompleted: complete || form.value.onboardingCompleted
    })
    form.value.defaultOrganizationId = response.item.defaultOrganizationId ?? null
    form.value.onboardingCompleted = response.item.onboardingCompleted
    saveMessage.value = complete ? 'Первичная настройка завершена' : 'Черновик настройки сохранен'
    if (complete) {
      await router.push('/')
    }
  } catch (error) {
    saveMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить настройки'
  } finally {
    isSaving.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Первичная настройка</p>
        <h1>Подготовка пространства клиента</h1>
        <p>Настройте бренд, первый филиал и сотрудников, чтобы клиент мог начать работу без системной админки.</p>
      </div>
    </div>

    <p v-if="loadError" class="form-error">{{ loadError }}</p>
    <p v-if="isLoading" class="empty-state">Загружаем настройки клиента...</p>

    <form v-else class="onboarding-grid" @submit.prevent="save(true)">
      <section class="detail-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Шаг 1</p>
            <h2>Пространство</h2>
          </div>
          <Building2 :size="22" />
        </div>
        <div class="settings-form settings-form-grid">
          <label>
            <span>Название клиента</span>
            <input v-model="form.displayName" type="text" />
          </label>
          <label>
            <span>Контактный email</span>
            <input v-model="form.contactEmail" type="email" />
          </label>
          <label class="settings-form-wide">
            <span>Описание</span>
            <textarea v-model="form.description" rows="3" />
          </label>
          <label>
            <span>Часовой пояс</span>
            <select v-model="form.timezone">
              <option value="Asia/Yekaterinburg">Asia/Yekaterinburg</option>
              <option value="Europe/Moscow">Europe/Moscow</option>
              <option value="UTC">UTC</option>
            </select>
          </label>
          <label>
            <span>Телефон</span>
            <input v-model="form.contactPhone" type="text" />
          </label>
        </div>
      </section>

      <section class="detail-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Шаг 2</p>
            <h2>Бренд</h2>
          </div>
          <Palette :size="22" />
        </div>
        <div class="settings-form settings-form-grid">
          <label>
            <span>Основной цвет</span>
            <input v-model="form.primaryColor" type="color" />
          </label>
          <label>
            <span>Акцентный цвет</span>
            <input v-model="form.accentColor" type="color" />
          </label>
          <label class="settings-form-wide">
            <span>URL логотипа</span>
            <input v-model="form.logoUrl" type="text" />
          </label>
        </div>
      </section>

      <section class="detail-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Шаг 3</p>
            <h2>Первое подразделение</h2>
          </div>
        </div>
        <div class="settings-form settings-form-grid">
          <label>
            <span>Название подразделения</span>
            <input v-model="form.defaultOrganizationName" type="text" @blur="syncOrganizationSlug" />
          </label>
          <label>
            <span>Slug</span>
            <input v-model="form.defaultOrganizationSlug" type="text" />
          </label>
        </div>
      </section>

      <section class="detail-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Шаг 4</p>
            <h2>Первые сотрудники</h2>
          </div>
          <MailPlus :size="22" />
        </div>
        <div class="invite-row">
          <input v-model="inviteEmail" type="email" placeholder="email сотрудника" />
          <select v-model="inviteRole">
            <option value="org_admin">Администратор</option>
            <option value="coordinator">Координатор</option>
            <option value="volunteer">Волонтер</option>
          </select>
          <button class="secondary-action" type="button" @click="addInvite">Добавить</button>
        </div>
        <div class="invite-list">
          <div v-for="invite in form.pendingInvites" :key="invite.email" class="invite-card">
            <span>
              <strong>{{ invite.email }}</strong>
              <small>{{ invite.role }}</small>
            </span>
            <button class="icon-button" type="button" :aria-label="`Удалить ${invite.email}`" @click="removeInvite(invite.email)">
              <Trash2 :size="16" />
            </button>
          </div>
        </div>
      </section>

      <div class="onboarding-actions">
        <button class="secondary-action" type="button" :disabled="isSaving" @click="save(false)">
          <Save :size="17" />
          <span>Сохранить черновик</span>
        </button>
        <button class="primary-action" type="submit" :disabled="isSaving || !canFinish">
          <Save :size="17" />
          <span>{{ isSaving ? 'Сохранение...' : 'Завершить настройку' }}</span>
        </button>
      </div>
    </form>
    <p v-if="saveMessage" class="form-success">{{ saveMessage }}</p>
  </section>
</template>
