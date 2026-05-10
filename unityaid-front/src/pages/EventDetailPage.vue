<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { CalendarDays, MapPin, Users } from 'lucide-vue-next'
import { fetchEvent } from '../entities/events/api'
import type { EventItem } from '../entities/events/types'
import { formatDateTime } from '../shared/date'

const route = useRoute()
const item = ref<EventItem | null>(null)
const errorMessage = ref('')

onMounted(async () => {
  try {
    item.value = (await fetchEvent(String(route.params.id))).item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить мероприятие'
  }
})
</script>

<template>
  <section class="page-section">
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <article v-else-if="item" class="detail-layout">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <RouterLink class="secondary-action" to="/calendar">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.status }} · {{ item.format }}</span>
        <p class="detail-summary">{{ item.description || 'Описание мероприятия пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><CalendarDays :size="18" /><span>{{ formatDateTime(item.startsAt) }} — {{ formatDateTime(item.endsAt) }}</span></div>
          <div><MapPin :size="18" /><span>{{ item.location || 'Место не указано' }}</span></div>
          <div><Users :size="18" /><span>Лимит: {{ item.maxParticipants || 'не указан' }}</span></div>
        </div>
      </div>
    </article>
  </section>
</template>
