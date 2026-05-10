<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HeartHandshake } from 'lucide-vue-next'
import { login } from '../entities/auth/store'

const route = useRoute()
const router = useRouter()

const email = ref('admin@unityaid.test')
const password = ref('password')
const isSubmitting = ref(false)
const errorMessage = ref('')

const redirectTo = computed(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' ? redirect : '/'
})

async function submit() {
  errorMessage.value = ''
  isSubmitting.value = true

  try {
    await login(email.value, password.value)
    await router.push(redirectTo.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось войти'
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
        <span>UnityAid</span>
      </div>
      <h1>Управление волонтерами, задачами и вкладом в одном месте</h1>
      <p>
        Войдите в демо-аккаунт, чтобы открыть рабочую панель с меню, уведомлениями и профилем.
      </p>
    </section>

    <section class="auth-panel">
      <form class="auth-form" @submit.prevent="submit">
        <div>
          <p class="eyebrow">Авторизация</p>
          <h2>Вход в систему</h2>
        </div>

        <label>
          <span>Email</span>
          <input v-model="email" type="email" autocomplete="email" required />
        </label>

        <label>
          <span>Пароль</span>
          <input v-model="password" type="password" autocomplete="current-password" required />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

        <button class="primary-button" type="submit" :disabled="isSubmitting">
          {{ isSubmitting ? 'Входим...' : 'Войти' }}
        </button>

        <div class="demo-hint">
          <strong>Демо:</strong> admin@unityaid.test / password
        </div>
      </form>
    </section>
  </main>
</template>
