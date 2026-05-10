<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Clock3, Flag, Link as LinkIcon } from 'lucide-vue-next'
import { fetchTask } from '../entities/tasks/api'
import type { TaskItem } from '../entities/tasks/types'
import { formatDateTime } from '../shared/date'

const route = useRoute()
const item = ref<TaskItem | null>(null)
const errorMessage = ref('')

onMounted(async () => {
  try {
    item.value = (await fetchTask(String(route.params.id))).item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить задачу'
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
        <RouterLink class="secondary-action" to="/tasks">К списку</RouterLink>
      </div>

      <div class="detail-panel">
        <span class="status-pill">{{ item.status }}</span>
        <p class="detail-summary">{{ item.description || 'Описание задачи пока не заполнено.' }}</p>
        <div class="detail-metrics">
          <div><Flag :size="18" /><span>Приоритет: {{ item.priority }}</span></div>
          <div><Clock3 :size="18" /><span>Срок: {{ formatDateTime(item.dueAt) }}</span></div>
          <div><LinkIcon :size="18" /><span>{{ item.eventTitle || 'Без мероприятия' }}</span></div>
        </div>
      </div>
    </article>
  </section>
</template>
