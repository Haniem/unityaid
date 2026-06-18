<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HeartHandshake, Link2 } from 'lucide-vue-next'
import { fetchInvitation, registerByInvitation } from '../entities/invitations/api'
import type { Invitation } from '../entities/invitations/types'

const route = useRoute()
const router = useRouter()
const invite = ref<Invitation | null>(null)
const isLoading = ref(true)
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const token = computed(() => {
  const value = route.query.invite
  return typeof value === 'string' ? value : ''
})

const form = reactive({
  lastName: '',
  firstName: '',
  patronymic: '',
  email: '',
  password: ''
})

async function load() {
  if (!token.value) {
    errorMessage.value = 'Ссылка приглашения не найдена.'
    isLoading.value = false
    return
  }
  try {
    invite.value = (await fetchInvitation(token.value)).item
    if (invite.value.email) form.email = invite.value.email
    if (invite.value.status !== 'pending') {
      errorMessage.value = 'Эта ссылка уже использована или недействительна.'
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить приглашение.'
  } finally {
    isLoading.value = false
  }
}

async function submit() {
  if (!token.value || !invite.value || invite.value.status !== 'pending') return
  errorMessage.value = ''
  successMessage.value = ''
  isSubmitting.value = true
  try {
    await registerByInvitation(token.value, form)
    successMessage.value = 'Аккаунт создан и привязан к организации. Сейчас откроется вход.'
    setTimeout(() => {
      router.push({ name: 'login', query: { email: form.email } })
    }, 900)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось зарегистрироваться по приглашению.'
  } finally {
    isSubmitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <main class="auth-page invite-register-page">
    <section class="auth-visual">
      <div class="brand-lockup">
        <span class="brand-mark"><HeartHandshake :size="28" /></span>
        <span>Пульс</span>
      </div>
      <h1>Регистрация волонтера по приглашению</h1>
      <p>Заполните профиль, и аккаунт сразу будет добавлен в организацию без подтверждения почты.</p>
    </section>

    <section class="auth-panel">
      <form class="auth-form invite-register-form" @submit.prevent="submit">
        <div>
          <p class="eyebrow">Приглашение</p>
          <h2>Создание аккаунта</h2>
        </div>

        <div v-if="invite" class="invite-summary">
          <Link2 :size="18" />
          <div>
            <strong>{{ invite.organizationName || 'Организация' }}</strong>
            <small>Роль: {{ invite.role === 'volunteer' ? 'Волонтер' : invite.role }}</small>
          </div>
        </div>

        <p v-if="isLoading" class="muted-text">Проверяем ссылку...</p>
        <template v-else>
          <label>
            <span>Фамилия</span>
            <input v-model="form.lastName" type="text" autocomplete="family-name" required :disabled="!!errorMessage && !invite" />
          </label>
          <label>
            <span>Имя</span>
            <input v-model="form.firstName" type="text" autocomplete="given-name" required :disabled="!!errorMessage && !invite" />
          </label>
          <label>
            <span>Отчество</span>
            <input v-model="form.patronymic" type="text" autocomplete="additional-name" />
          </label>
          <label>
            <span>Email</span>
            <input v-model="form.email" type="email" autocomplete="email" required />
          </label>
          <label>
            <span>Пароль</span>
            <input v-model="form.password" type="password" autocomplete="new-password" minlength="8" required />
          </label>
        </template>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

        <button class="primary-button" type="submit" :disabled="isSubmitting || isLoading || !!errorMessage || invite?.status !== 'pending'">
          {{ isSubmitting ? 'Создаем аккаунт...' : 'Зарегистрироваться' }}
        </button>

        <RouterLink class="secondary-action auth-link" to="/login">Уже есть аккаунт</RouterLink>
      </form>
    </section>
  </main>
</template>
