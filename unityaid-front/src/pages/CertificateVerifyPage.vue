<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { BadgeCheck, SearchCheck } from 'lucide-vue-next'
import { verifyCertificate } from '../entities/certificates/api'
import type { Certificate } from '../entities/certificates/types'

const route = useRoute()
const code = ref('')
const certificate = ref<Certificate | null>(null)
const message = ref('')
const isLoading = ref(false)

async function verify() {
  const value = code.value.trim()
  if (!value) return
  isLoading.value = true
  message.value = ''
  certificate.value = null
  try {
    const response = await verifyCertificate(value)
    certificate.value = response.item
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Сертификат не найден'
  } finally {
    isLoading.value = false
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' })
}

onMounted(() => {
  const routeCode = route.params.code
  if (typeof routeCode === 'string') {
    code.value = routeCode
    verify()
  }
})
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Проверка</p>
        <h1>Проверка сертификата</h1>
        <p>Введите проверочный код документа, чтобы подтвердить его выпуск в системе «Пульс».</p>
      </div>
    </div>

    <form class="detail-panel certificate-verify-form" @submit.prevent="verify">
      <label class="search-field">
        <SearchCheck :size="18" />
        <input v-model="code" type="search" placeholder="Код сертификата" />
      </label>
      <button class="primary-action" type="submit" :disabled="isLoading || !code.trim()">
        <span>{{ isLoading ? 'Проверка...' : 'Проверить' }}</span>
      </button>
    </form>

    <p v-if="message" class="form-error">{{ message }}</p>

    <article v-if="certificate" class="detail-panel certificate-result">
      <BadgeCheck :size="28" />
      <div>
        <p class="eyebrow">Документ найден</p>
        <h2>{{ certificate.title }}</h2>
        <p>{{ certificate.description || 'Документ подтвержден.' }}</p>
        <div class="certificate-meta">
          <span>{{ certificate.totalHours }} ч.</span>
          <span>{{ certificate.verifyCode }}</span>
          <span>{{ formatDate(certificate.issuedAt) }}</span>
        </div>
      </div>
    </article>
  </section>
</template>
