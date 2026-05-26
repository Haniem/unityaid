<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { BookOpen, Filter, Pencil, Plus, Search, Trash2, X } from 'lucide-vue-next'
import {
  deleteKnowledgeArticle,
  fetchKnowledgeArticles,
  fetchKnowledgeCategories
} from '../entities/knowledge/api'
import type { KnowledgeArticle, KnowledgeCategory } from '../entities/knowledge/types'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

const items = ref<KnowledgeArticle[]>([])
const categories = ref<KnowledgeCategory[]>([])
const isLoading = ref(true)
const isFiltersOpen = ref(false)
const errorMessage = ref('')
const search = ref('')
const status = ref('')
const categoryId = ref('')
const { page, perPage, pageItems } = useClientPagination(items, 12)

const statusOptions = [
  { id: '', name: 'Все статусы' },
  { id: 'published', name: 'Опубликовано' },
  { id: 'draft', name: 'Черновики' },
  { id: 'archived', name: 'Архив' }
]

async function load() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchKnowledgeArticles({ search: search.value, status: status.value, categoryId: categoryId.value })
    items.value = response.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить базу знаний'
  } finally {
    isLoading.value = false
  }
}

async function removeArticle(item: KnowledgeArticle) {
  if (!confirm(`Удалить статью "${item.title}"?`)) return
  await deleteKnowledgeArticle(item.id)
  await load()
}

function formatDate(value?: string | null) {
  if (!value) return 'Не опубликовано'
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value))
}

let timer: number | undefined
watch([search, status, categoryId], () => {
  window.clearTimeout(timer)
  timer = window.setTimeout(load, 250)
})

onMounted(async () => {
  categories.value = (await fetchKnowledgeCategories()).items
  await load()
})
</script>

<template>
  <section class="page-section">
    <div class="page-heading with-search">
      <div>
        <p class="eyebrow">Обучение</p>
        <h1>База знаний</h1>
      </div>
      <label class="search-field heading-search">
        <Search :size="18" />
        <input v-model="search" type="search" placeholder="Поиск по статьям" />
      </label>
      <button class="secondary-action" type="button" @click="isFiltersOpen = true">
        <Filter :size="18" />
        <span>Фильтры</span>
      </button>
      <RouterLink class="primary-action" to="/knowledge-base/new">
        <Plus :size="18" />
        <span>Создать</span>
      </RouterLink>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-if="isLoading" class="empty-state">Загрузка статей...</div>
    <div v-else-if="items.length === 0" class="empty-state">Статей пока нет.</div>

    <div v-else class="knowledge-grid">
      <article v-for="item in pageItems" :key="item.id" class="knowledge-card">
        <RouterLink :to="`/knowledge-base/${item.id}`" class="knowledge-card-main">
          <span class="knowledge-card-icon"><BookOpen :size="22" /></span>
          <div>
            <span class="status-pill">{{ item.status }} · {{ item.categoryName || 'без категории' }}</span>
            <h2>{{ item.title }}</h2>
            <p>{{ item.summary || 'Краткое описание пока не заполнено.' }}</p>
            <small>{{ item.authorName || 'Пульс' }} · {{ formatDate(item.publishedAt || item.createdAt) }}</small>
          </div>
        </RouterLink>
        <div class="card-actions">
          <RouterLink class="icon-button" :to="`/knowledge-base/${item.id}/edit`" aria-label="Редактировать"><Pencil :size="17" /></RouterLink>
          <button class="icon-button" type="button" aria-label="Удалить" @click="removeArticle(item)"><Trash2 :size="17" /></button>
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
