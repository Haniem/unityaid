<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HeartHandshake } from 'lucide-vue-next'
import {
  login,
  register,
  requestPasswordReset,
  resetPassword,
  verifyEmail
} from '../entities/auth/store'

type AuthMode = 'login' | 'register' | 'forgot' | 'reset' | 'verify'

const route = useRoute()
const router = useRouter()

const mode = ref<AuthMode>('login')
const email = ref(typeof route.query.email === 'string' ? route.query.email : 'admin@puls.test')
const password = ref('password')
const firstName = ref('')
const lastName = ref('')
const token = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const redirectTo = computed(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' ? redirect : '/'
})

const title = computed(() => {
  if (mode.value === 'register') return 'Регистрация'
  if (mode.value === 'forgot') return 'Восстановление пароля'
  if (mode.value === 'reset') return 'Новый пароль'
  if (mode.value === 'verify') return 'Подтверждение email'
  return 'Вход в систему'
})

function switchMode(nextMode: AuthMode) {
  mode.value = nextMode
  errorMessage.value = ''
  successMessage.value = ''
}

async function submit() {
  errorMessage.value = ''
  successMessage.value = ''
  isSubmitting.value = true

  try {
    if (mode.value === 'login') {
      await login(email.value, password.value)
      await router.push(redirectTo.value)
      return
    }

    if (mode.value === 'register') {
      const response = await register({
        email: email.value,
        password: password.value,
        firstName: firstName.value,
        lastName: lastName.value
      })
      token.value = response.token ?? ''
      successMessage.value = response.token
        ? 'Аккаунт создан. Dev-токен подтверждения подставлен ниже.'
        : 'Аккаунт создан. Проверьте письмо для подтверждения email.'
      mode.value = 'verify'
      return
    }

    if (mode.value === 'forgot') {
      const response = await requestPasswordReset(email.value)
      token.value = response.token ?? ''
      successMessage.value = response.token
        ? 'Dev-токен восстановления подставлен ниже.'
        : 'Если email существует, мы отправили инструкцию восстановления.'
      mode.value = 'reset'
      return
    }

    if (mode.value === 'reset') {
      await resetPassword(token.value, password.value)
      successMessage.value = 'Пароль обновлен. Теперь можно войти.'
      mode.value = 'login'
      return
    }

    if (mode.value === 'verify') {
      await verifyEmail(token.value)
      successMessage.value = 'Email подтвержден. Теперь можно войти.'
      mode.value = 'login'
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось выполнить действие'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-visual">
      <div class="brand-lockup">
        <span class="brand-mark"><HeartHandshake :size="28" /></span>
        <span>Пульс</span>
      </div>
      <h1>Управление волонтерами, задачами и вкладом в одном месте</h1>
      <p>
        Войдите в демо-аккаунт или создайте нового пользователя, чтобы проверить полный поток авторизации.
      </p>
    </section>

    <section class="auth-panel">
      <form class="auth-form" @submit.prevent="submit">
        <div>
          <p class="eyebrow">Авторизация</p>
          <h2>{{ title }}</h2>
        </div>

        <div class="auth-tabs" aria-label="Режим авторизации">
          <button type="button" :class="{ active: mode === 'login' }" @click="switchMode('login')">Вход</button>
          <button type="button" :class="{ active: mode === 'register' }" @click="switchMode('register')">Регистрация</button>
          <button type="button" :class="{ active: mode === 'forgot' }" @click="switchMode('forgot')">Сброс</button>
        </div>

        <label v-if="mode === 'register'">
          <span>Имя</span>
          <input v-model="firstName" type="text" autocomplete="given-name" required />
        </label>

        <label v-if="mode === 'register'">
          <span>Фамилия</span>
          <input v-model="lastName" type="text" autocomplete="family-name" required />
        </label>

        <label v-if="mode === 'login' || mode === 'register' || mode === 'forgot'">
          <span>Email</span>
          <input v-model="email" type="email" autocomplete="email" required />
        </label>

        <label v-if="mode === 'login' || mode === 'register' || mode === 'reset'">
          <span>{{ mode === 'reset' ? 'Новый пароль' : 'Пароль' }}</span>
          <input v-model="password" type="password" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" required />
        </label>

        <label v-if="mode === 'reset' || mode === 'verify'">
          <span>Токен</span>
          <input v-model="token" type="text" required />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

        <button class="primary-button" type="submit" :disabled="isSubmitting">
          {{ isSubmitting ? 'Отправляем...' : 'Продолжить' }}
        </button>

        <button v-if="mode === 'forgot'" class="secondary-action auth-link" type="button" @click="switchMode('reset')">
          У меня уже есть токен
        </button>
        <button v-if="mode === 'register'" class="secondary-action auth-link" type="button" @click="switchMode('verify')">
          Подтвердить email токеном
        </button>

        <div class="demo-hint">
          <strong>Демо:</strong> admin@puls.test / password
        </div>
      </form>
    </section>
  </main>
</template>
