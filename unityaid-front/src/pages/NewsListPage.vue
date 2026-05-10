<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Plus, Pencil, Trash2 } from 'lucide-vue-next'
import { deleteNews, fetchNewsList } from '../entities/news/api'
import type { NewsItem } from '../entities/news/types'

const items = ref<NewsItem[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

async function loadNews() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchNewsList()
    items.value = response.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить новости'
  } finally {
    isLoading.value = false
  }
}

async function removeNews(item: NewsItem) {
  if (!confirm(`Удалить новость "${item.title}"?`)) {
    return
  }
  await deleteNews(item.id)
  await loadNews()
}

function formatDate(value?: string | null) {
  if (!value) {
    return 'Не опубликовано'
  }
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value))
}

onMounted(loadNews)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Моя компания</p>
        <h1>Новости</h1>
      </div>
      <RouterLink class="primary-action" to="/news/new">
        <Plus :size="18" />
        <span>Создать новость</span>
      </RouterLink>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-if="isLoading" class="empty-state">Загрузка новостей...</div>

    <div v-else-if="items.length === 0" class="empty-state">
      Новостей пока нет. Создайте первую публикацию для организации.
    </div>

    <div v-else class="news-grid">
      <article v-for="item in items" :key="item.id" class="news-card">
        <RouterLink :to="`/news/${item.id}`" class="news-card-main">
          <img v-if="item.coverImageUrl" :src="item.coverImageUrl" alt="" />
          <div v-else class="news-cover-placeholder">UnityAid</div>
          <div class="news-card-body">
            <span class="status-pill">{{ item.status === 'published' ? 'Опубликовано' : 'Черновик' }}</span>
            <h2>{{ item.title }}</h2>
            <p>{{ item.summary || 'Краткое описание пока не заполнено.' }}</p>
            <small>{{ item.organizationName }} · {{ formatDate(item.publishedAt || item.createdAt) }}</small>
          </div>
        </RouterLink>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/news/${item.id}/edit`" aria-label="Редактировать">
            <Pencil :size="17" />
          </RouterLink>
          <button class="icon-button" type="button" aria-label="Удалить" @click="removeNews(item)">
            <Trash2 :size="17" />
          </button>
        </div>
      </article>
    </div>
  </section>
</template>
