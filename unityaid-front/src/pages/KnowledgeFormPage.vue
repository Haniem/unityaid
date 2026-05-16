<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Save } from 'lucide-vue-next'
import {
  createKnowledgeArticle,
  createKnowledgeCategory,
  fetchKnowledgeArticle,
  fetchKnowledgeCategories,
  updateKnowledgeArticle
} from '../entities/knowledge/api'
import type { KnowledgeCategory, KnowledgePayload } from '../entities/knowledge/types'
import CustomSelect from '../shared/ui/CustomSelect.vue'

const route = useRoute()
const router = useRouter()
const categories = ref<KnowledgeCategory[]>([])
const isSaving = ref(false)
const errorMessage = ref('')
const categoryName = ref('')
const isEdit = computed(() => Boolean(route.params.id))
const statusOptions = [{ id: 'draft', name: 'Черновик' }, { id: 'published', name: 'Опубликовано' }, { id: 'archived', name: 'Архив' }]

const form = reactive<KnowledgePayload>({
  categoryId: null,
  title: '',
  summary: '',
  contentHtml: '',
  status: 'draft'
})

async function save() {
  isSaving.value = true
  errorMessage.value = ''
  try {
    const payload = { ...form, categoryId: form.categoryId || null }
    const response = isEdit.value
      ? await updateKnowledgeArticle(String(route.params.id), payload)
      : await createKnowledgeArticle(payload)
    await router.push(`/knowledge-base/${response.item.id}`)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить статью'
  } finally {
    isSaving.value = false
  }
}

async function addCategory() {
  if (!categoryName.value.trim()) return
  const response = await createKnowledgeCategory(categoryName.value.trim())
  categories.value.push(response.item)
  form.categoryId = response.item.id
  categoryName.value = ''
}

onMounted(async () => {
  categories.value = (await fetchKnowledgeCategories()).items
  if (isEdit.value) {
    const article = (await fetchKnowledgeArticle(String(route.params.id))).item
    form.categoryId = article.categoryId || null
    form.title = article.title
    form.summary = article.summary
    form.contentHtml = article.contentHtml
    form.status = article.status
  }
})
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">База знаний</p>
        <h1>{{ isEdit ? 'Редактирование статьи' : 'Новая статья' }}</h1>
      </div>
    </div>

    <form class="entity-form wide-form" @submit.prevent="save">
      <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

      <label class="form-field">
        <span class="field-label">Заголовок</span>
        <input v-model="form.title" required />
      </label>

      <label class="form-field">
        <span class="field-label">Краткое описание</span>
        <textarea v-model="form.summary" rows="3"></textarea>
      </label>

      <div class="form-columns">
        <label class="form-field">
          <span class="field-label">Категория</span>
          <CustomSelect v-model="form.categoryId" :options="[{ id: '', name: 'Без категории' }, ...categories.map((category) => ({ id: category.id, name: category.name }))]" />
        </label>

        <label class="form-field">
          <span class="field-label">Статус</span>
          <CustomSelect v-model="form.status" :options="statusOptions" />
        </label>
      </div>

      <div class="inline-member-form">
        <input v-model="categoryName" placeholder="Новая категория" />
        <button class="secondary-action" type="button" @click="addCategory">Добавить</button>
      </div>

      <label class="form-field">
        <span class="field-label">Содержимое</span>
        <textarea v-model="form.contentHtml" rows="16" placeholder="<p>Текст статьи...</p>"></textarea>
      </label>

      <button class="primary-action" type="submit" :disabled="isSaving">
        <Save :size="18" />
        Сохранить
      </button>
    </form>
  </section>
</template>
