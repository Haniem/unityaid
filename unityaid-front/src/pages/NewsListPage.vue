<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'
import { cleanupNewsFiles, deleteNews, fetchNewsCategories, fetchNewsList } from '../entities/news/api'
import type { NewsCategory, NewsItem } from '../entities/news/types'

const items = ref<NewsItem[]>([])
const categories = ref<NewsCategory[]>([])
const isLoading = ref(true)
const errorMessage = ref('')
const cleanupMessage = ref('')
const search = ref('')
const status = ref('')
const categoryId = ref('')

async function loadNews() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchNewsList({ search: search.value, status: status.value, categoryId: categoryId.value })
    items.value = response.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить новости'
  } finally {
    isLoading.value = false
  }
}

async function removeNews(item: NewsItem) {
  if (!confirm(`Удалить новость "${item.title}"?`)) return
  await deleteNews(item.id)
  await loadNews()
}

async function cleanupFiles() {
  const response = await cleanupNewsFiles()
  cleanupMessage.value = `Удалено файлов: ${response.removed}`
}

function formatDate(value?: string | null) {
  if (!value) return 'Не опубликовано'
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value))
}

let timer: number | undefined
watch([search, status, categoryId], () => {
  window.clearTimeout(timer)
  timer = window.setTimeout(loadNews, 250)
})

onMounted(async () => {
  categories.value = (await fetchNewsCategories()).items
  await loadNews()
})
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

    <div class="filter-bar">
      <label class="search-field"><Search :size="18" /><input v-model="search" type="search" placeholder="Поиск новостей" /></label>
      <select v-model="status"><option value="">Все статусы</option><option value="draft">Черновики</option><option value="scheduled">Запланировано</option><option value="published">Опубликовано</option></select>
      <select v-model="categoryId"><option value="">Все категории</option><option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option></select>
      <button class="secondary-action" type="button" @click="cleanupFiles">Очистить файлы</button>
    </div>

    <p v-if="cleanupMessage" class="form-success">{{ cleanupMessage }}</p>
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-if="isLoading" class="empty-state">Загрузка новостей...</div>
    <div v-else-if="items.length === 0" class="empty-state">Новостей пока нет.</div>

    <div v-else class="news-grid">
      <article v-for="item in items" :key="item.id" class="news-card">
        <RouterLink :to="`/news/${item.id}`" class="news-card-main">
          <img v-if="item.coverImageUrl" :src="item.coverImageUrl" alt="" />
          <div v-else class="news-cover-placeholder">UnityAid</div>
          <div class="news-card-body">
            <span class="status-pill">{{ item.status }} · {{ item.categoryName || 'без категории' }}</span>
            <h2>{{ item.title }}</h2>
            <p>{{ item.summary || 'Краткое описание пока не заполнено.' }}</p>
            <small>{{ item.organizationName }} · {{ formatDate(item.scheduledAt || item.publishedAt || item.createdAt) }}</small>
          </div>
        </RouterLink>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/news/${item.id}/edit`" aria-label="Редактировать"><Pencil :size="17" /></RouterLink>
          <button class="icon-button" type="button" aria-label="Удалить" @click="removeNews(item)"><Trash2 :size="17" /></button>
        </div>
      </article>
    </div>
  </section>
</template>
