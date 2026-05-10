<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Pencil, Trash2 } from 'lucide-vue-next'
import { deleteNews, fetchNewsItem } from '../entities/news/api'
import type { NewsItem } from '../entities/news/types'

const route = useRoute()
const router = useRouter()
const item = ref<NewsItem | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')

async function loadNews() {
  isLoading.value = true
  try {
    const response = await fetchNewsItem(String(route.params.id))
    item.value = response.item
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить новость'
  } finally {
    isLoading.value = false
  }
}

async function removeNews() {
  if (!item.value || !confirm(`Удалить новость "${item.value.title}"?`)) {
    return
  }
  await deleteNews(item.value.id)
  await router.push('/news')
}

function formatDate(value?: string | null) {
  if (!value) {
    return 'Не опубликовано'
  }
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'long', timeStyle: 'short' }).format(new Date(value))
}

onMounted(loadNews)
</script>

<template>
  <section class="page-section">
    <div v-if="isLoading" class="empty-state">Загрузка новости...</div>
    <p v-else-if="errorMessage" class="form-error">{{ errorMessage }}</p>

    <article v-else-if="item" class="news-detail">
      <div class="page-heading">
        <div>
          <p class="eyebrow">{{ item.organizationName || 'Новости' }}</p>
          <h1>{{ item.title }}</h1>
        </div>
        <div class="page-actions">
          <RouterLink class="secondary-action" :to="`/news/${item.id}/edit`">
            <Pencil :size="17" />
            <span>Редактировать</span>
          </RouterLink>
          <button class="secondary-action danger-action" type="button" @click="removeNews">
            <Trash2 :size="17" />
            <span>Удалить</span>
          </button>
        </div>
      </div>

      <img v-if="item.coverImageUrl" class="detail-cover" :src="item.coverImageUrl" alt="" />
      <p class="detail-meta">
        {{ item.authorName || 'Автор не указан' }} · {{ formatDate(item.publishedAt || item.createdAt) }}
      </p>
      <p v-if="item.summary" class="detail-summary">{{ item.summary }}</p>
      <div class="rich-content" v-html="item.contentHtml"></div>
    </article>
  </section>
</template>
