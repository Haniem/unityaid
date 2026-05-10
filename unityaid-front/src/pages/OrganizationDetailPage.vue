<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Mail } from 'lucide-vue-next'
import { fetchOrganization } from '../entities/organizations/api'
import type { Organization } from '../entities/organizations/types'

const route = useRoute()
const item = ref<Organization | null>(null)
const errorMessage = ref('')

onMounted(async () => {
  try {
    item.value = (await fetchOrganization(String(route.params.id))).item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить организацию'
  }
})
</script>

<template>
  <section class="page-section">
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <article v-else-if="item" class="detail-layout">
      <div class="page-heading">
        <div>
          <p class="eyebrow">Организация</p>
          <h1>{{ item.name }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/organizations">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.slug }}</span>
        <p class="detail-summary">{{ item.description || 'Описание организации пока не заполнено.' }}</p>
        <div class="meta-line">
          <Mail :size="16" />
          <span>{{ item.contactEmail || 'Контактный email не указан' }}</span>
        </div>
      </div>
    </article>
  </section>
</template>
