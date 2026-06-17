<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { Filter, Pencil, Plus, Search, Trash2, X } from 'lucide-vue-next'
import { deleteNews, fetchNewsCategories, fetchNewsList } from '../entities/news/api'
import type { NewsCategory, NewsItem } from '../entities/news/types'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

const items = ref<NewsItem[]>([])
const categories = ref<NewsCategory[]>([])
const isLoading = ref(true)
const isFiltersOpen = ref(false)
const errorMessage = ref('')
const search = ref('')
const status = ref('')
const categoryId = ref('')
const { page, perPage, pageItems } = useClientPagination(items, 12)

const statusOptions = [
  { id: '', name: 'Все статусы' },
  { id: 'draft', name: 'Черновики' },
  { id: 'scheduled', name: 'Запланировано' },
  { id: 'published', name: 'Опубликовано' }
]

const statusLabels: Record<string, string> = {
  draft: 'Черновик',
  scheduled: 'Запланировано',
  published: 'Опубликовано'
}

async function loadNews() {
  isLoading.value = true
  errorMessage.value = ''
  page.value = 1
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
    <div class="page-heading with-search">
      <div>
        <p class="eyebrow">Раздел</p>
        <h1>Новости</h1>
      </div>
      <label class="search-field heading-search">
        <Search :size="18" />
        <input v-model="search" type="search" placeholder="Поиск новостей" />
      </label>
      <button class="secondary-action" type="button" @click="isFiltersOpen = true">
        <Filter :size="18" />
        <span>Фильтры</span>
      </button>
      <RouterLink class="primary-action" to="/news/new">
        <Plus :size="18" />
        <span>Создать</span>
      </RouterLink>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-if="isLoading" class="empty-state">Загрузка новостей...</div>
    <div v-else-if="items.length === 0" class="empty-state">Новостей пока нет.</div>

    <div v-else class="news-grid">
      <article v-for="item in pageItems" :key="item.id" class="news-card">
        <RouterLink :to="`/news/${item.id}`" class="news-card-cover" aria-label="Открыть новость">
          <img v-if="item.coverImageUrl" :src="item.coverImageUrl" alt="" />
          <div v-else class="news-cover-placeholder">Пульс</div>
        </RouterLink>
        <div class="news-card-body">
          <span class="status-pill">{{ statusLabels[item.status] || item.status }} · {{ item.categoryName || 'без категории' }}</span>
          <RouterLink :to="`/news/${item.id}`" class="news-card-title"><h2>{{ item.title }}</h2></RouterLink>
          <p>{{ item.summary || 'Краткое описание пока не заполнено.' }}</p>
          <div class="news-card-footer">
            <small>{{ item.organizationName || 'Без организации' }} · {{ formatDate(item.scheduledAt || item.publishedAt || item.createdAt) }}</small>
            <div class="card-actions">
              <RouterLink class="icon-button" :to="`/news/${item.id}/edit`" aria-label="Редактировать"><Pencil :size="17" /></RouterLink>
              <button class="icon-button" type="button" aria-label="Удалить" @click="removeNews(item)"><Trash2 :size="17" /></button>
            </div>
          </div>
        </div>
      </article>
    </div>
    <PaginationBar v-if="!isLoading" v-model:page="page" :per-page="perPage" :total="items.length" />

    <div v-if="isFiltersOpen" class="drawer-backdrop" @click.self="isFiltersOpen = false">
      <aside class="filter-drawer">
        <div class="drawer-heading">
          <h2>Фильтры</h2>
          <button class="icon-button" type="button" aria-label="Закрыть" @click="isFiltersOpen = false"><X :size="18" /></button>
        </div>
        <label class="form-field">
          <span class="field-label">Статус</span>
          <CustomSelect v-model="status" :options="statusOptions" />
        </label>
        <label class="form-field">
          <span class="field-label">Категория</span>
          <CustomSelect
            v-model="categoryId"
            :options="[{ id: '', name: 'Все категории' }, ...categories.map((item) => ({ id: item.id, name: item.name }))]"
          />
        </label>
      </aside>
    </div>
  </section>
</template>
