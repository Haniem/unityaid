<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Clock3, KeyRound } from 'lucide-vue-next'
import { authState, changePassword } from '../entities/auth/store'
import { fetchTimeEntries } from '../entities/timeentries/api'
import type { TimeEntry } from '../entities/timeentries/types'

const currentPassword = ref('')
const newPassword = ref('')
const isSubmitting = ref(false)
const isPasswordModalOpen = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const approvedEntries = ref<TimeEntry[]>([])

const totalApprovedHours = computed(() => approvedEntries.value.reduce((sum, item) => sum + item.hours, 0))

function openPasswordModal() {
  currentPassword.value = ''
  newPassword.value = ''
  errorMessage.value = ''
  successMessage.value = ''
  isPasswordModalOpen.value = true
}

async function submitPasswordChange() {
  errorMessage.value = ''
  successMessage.value = ''
  isSubmitting.value = true

  try {
    await changePassword(currentPassword.value, newPassword.value)
    currentPassword.value = ''
    newPassword.value = ''
    successMessage.value = 'Пароль обновлен'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сменить пароль'
  } finally {
    isSubmitting.value = false
  }
}

async function loadHours() {
  if (!authState.user) return
  const response = await fetchTimeEntries({ status: 'approved', userId: authState.user.id })
  approvedEntries.value = response.items.filter((item) => item.userId === authState.user?.id)
}

onMounted(loadHours)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Профиль</p>
        <h1>{{ authState.user?.lastName }} {{ authState.user?.firstName }}</h1>
      </div>
      <button class="secondary-action" type="button" @click="openPasswordModal">
        <KeyRound :size="18" />
        <span>Сброс пароля</span>
      </button>
    </div>

    <div class="profile-panel">
      <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
      <dl>
        <div>
          <dt>Роль</dt>
          <dd>{{ authState.user?.primaryRole }}</dd>
        </div>
        <div>
          <dt>Email</dt>
          <dd>{{ authState.user?.email }} · {{ authState.user?.isEmailVerified ? 'подтвержден' : 'не подтвержден' }}</dd>
        </div>
        <div>
          <dt>Организации</dt>
          <dd>{{ authState.user?.organizations.map((item) => item.organizationName).join(', ') || 'Нет организаций' }}</dd>
        </div>
        <div>
          <dt>Подтвержденные часы</dt>
          <dd class="profile-hours"><Clock3 :size="16" /> {{ totalApprovedHours.toFixed(2) }} ч.</dd>
        </div>
      </dl>
    </div>

    <div v-if="isPasswordModalOpen" class="modal-backdrop" @click.self="isPasswordModalOpen = false">
      <form class="modal-panel entity-form password-panel" @submit.prevent="submitPasswordChange">
        <div>
          <p class="eyebrow">Безопасность</p>
          <h2>Смена пароля</h2>
        </div>

        <label>
          <span>Текущий пароль</span>
          <input v-model="currentPassword" type="password" autocomplete="current-password" required />
        </label>

        <label>
          <span>Новый пароль</span>
          <input v-model="newPassword" type="password" autocomplete="new-password" minlength="8" required />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

        <div class="form-actions">
          <button class="secondary-action" type="button" @click="isPasswordModalOpen = false">Отмена</button>
          <button class="primary-action" type="submit" :disabled="isSubmitting">
            {{ isSubmitting ? 'Сохраняем...' : 'Обновить пароль' }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>
