<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Pencil } from 'lucide-vue-next'
import { fetchKnowledgeArticle } from '../entities/knowledge/api'
import type { KnowledgeArticle } from '../entities/knowledge/types'

const route = useRoute()
const item = ref<KnowledgeArticle | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')

function formatDate(value?: string | null) {
  if (!value) return ''
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'long' }).format(new Date(value))
}

onMounted(async () => {
  try {
    item.value = (await fetchKnowledgeArticle(String(route.params.id))).item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить статью'
  } finally {
    isLoading.value = false
  }
})
</script>

<template>
  <section class="page-section">
    <div v-if="isLoading" class="empty-state">Загрузка статьи...</div>
    <p v-else-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article v-else-if="item" class="knowledge-detail">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.categoryName || 'База знаний' }}</p>
          <h1>{{ item.title }}</h1>
          <p>{{ item.summary }}</p>
          <small>{{ item.authorName || 'UnityAid' }} · {{ formatDate(item.publishedAt || item.createdAt) }}</small>
        </div>
        <RouterLink class="secondary-action" :to="`/knowledge-base/${item.id}/edit`">
          <Pencil :size="18" />
          Редактировать
        </RouterLink>
      </div>
      <div class="rich-content knowledge-content" v-html="item.contentHtml"></div>
    </article>
  </section>
</template>
