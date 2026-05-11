<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Heading1, Heading2, ImagePlus, Link, List, Plus, Save } from 'lucide-vue-next'
import { createNews, createNewsCategory, fetchNewsItem, updateNews, uploadNewsImage } from '../entities/news/api'
import type { NewsPayload } from '../entities/news/types'
import { fetchCreateForm, fetchEditForm } from '../entities/forms/api'
import type { BackendForm, FormModel } from '../entities/forms/types'
import DynamicForm from '../shared/ui/DynamicForm.vue'
import { isoFromDatetimeLocal, modelFromForm, nullable, stringValue } from '../shared/forms'

const route = useRoute()
const router = useRouter()

const editorRef = ref<HTMLElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const formSchema = ref<BackendForm | null>(null)
const formModel = ref<FormModel>({})
const contentHtml = ref('')
const newCategoryName = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const categoryError = ref('')

const newsId = computed(() => (typeof route.params.id === 'string' ? route.params.id : null))
const isEdit = computed(() => Boolean(newsId.value))

async function loadNews() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    formSchema.value = newsId.value ? await fetchEditForm('news', newsId.value) : await fetchCreateForm('news')
    formModel.value = modelFromForm(formSchema.value)
    if (newsId.value) {
      const response = await fetchNewsItem(newsId.value)
      contentHtml.value = response.item.contentHtml
    }
    await nextTick()
    if (editorRef.value) editorRef.value.innerHTML = contentHtml.value
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить форму новости'
  } finally {
    isLoading.value = false
  }
}

function syncEditor() {
  contentHtml.value = editorRef.value?.innerHTML ?? ''
}

function format(command: string, value?: string) {
  document.execCommand(command, false, value)
  syncEditor()
}

function createLink() {
  const url = prompt('URL ссылки')
  if (url) format('createLink', url)
}

function openImageDialog() {
  fileInputRef.value?.click()
}

async function uploadImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const response = await uploadNewsImage(file)
    if (!formModel.value.coverImageUrl) formModel.value = { ...formModel.value, coverImageUrl: response.url }
    document.execCommand('insertImage', false, response.url)
    syncEditor()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить изображение'
  } finally {
    input.value = ''
  }
}

async function addCategory() {
  const name = newCategoryName.value.trim()
  if (!name) return
  categoryError.value = ''
  try {
    const response = await createNewsCategory(name)
    formSchema.value = newsId.value ? await fetchEditForm('news', newsId.value) : await fetchCreateForm('news')
    formModel.value = { ...formModel.value, categoryId: response.item.id }
    newCategoryName.value = ''
  } catch (error) {
    categoryError.value = error instanceof Error ? error.message : 'Не удалось добавить категорию'
  }
}

async function submit() {
  syncEditor()
  errorMessage.value = ''
  isSaving.value = true

  try {
    const payload: NewsPayload = {
      title: stringValue(formModel.value.title),
      summary: stringValue(formModel.value.summary),
      contentHtml: contentHtml.value,
      coverImageUrl: nullable(formModel.value.coverImageUrl) as string | null,
      categoryId: nullable(formModel.value.categoryId) as string | null,
      status: (stringValue(formModel.value.status) || 'published') as NewsPayload['status'],
      scheduledAt: formModel.value.status === 'scheduled' ? isoFromDatetimeLocal(formModel.value.scheduledAt) : null
    }
    const response = isEdit.value && newsId.value ? await updateNews(newsId.value, payload) : await createNews(payload)
    await router.push(`/news/${response.item.id}`)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить новость'
  } finally {
    isSaving.value = false
  }
}

onMounted(loadNews)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Раздел</p>
        <h1>{{ isEdit ? 'Редактирование новости' : 'Создание новости' }}</h1>
      </div>
    </div>

    <form class="news-form" @submit.prevent="submit">
      <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
      <div v-if="isLoading" class="empty-state">Загрузка формы...</div>

      <template v-else>
        <DynamicForm v-if="formSchema" v-model="formModel" :form="formSchema" />

        <label class="form-field category-create">
          <span class="field-label">
            <span>Новая категория</span>
            <span class="requirement-badge optional">необязательно</span>
          </span>
          <span class="inline-input">
            <input v-model="newCategoryName" placeholder="Например, обучение" />
            <button class="secondary-action" type="button" @click="addCategory"><Plus :size="16" />Добавить</button>
          </span>
          <small class="field-help">Категория появится в списке выбора сразу после создания.</small>
          <small v-if="categoryError" class="form-error">{{ categoryError }}</small>
        </label>

        <div class="editor-shell">
          <div class="editor-toolbar">
            <button type="button" title="Заголовок" @click="format('formatBlock', 'h2')"><Heading1 :size="17" /></button>
            <button type="button" title="Подзаголовок" @click="format('formatBlock', 'h3')"><Heading2 :size="17" /></button>
            <button type="button" title="Жирный" @click="format('bold')"><strong>B</strong></button>
            <button type="button" title="Курсив" @click="format('italic')"><em>I</em></button>
            <button type="button" title="Список" @click="format('insertUnorderedList')"><List :size="17" /></button>
            <button type="button" title="Ссылка" @click="createLink"><Link :size="17" /></button>
            <button type="button" title="Изображение" @click="openImageDialog"><ImagePlus :size="17" /></button>
            <input ref="fileInputRef" type="file" accept="image/*" hidden @change="uploadImage" />
          </div>
          <div ref="editorRef" class="wysiwyg-editor" contenteditable="true" data-placeholder="Введите текст новости" @input="syncEditor"></div>
        </div>

        <div class="form-actions">
          <RouterLink class="secondary-action" to="/news">Отмена</RouterLink>
          <button class="primary-action" type="submit" :disabled="isSaving">
            <Save :size="18" />
            <span>{{ isSaving ? 'Сохранение...' : 'Сохранить' }}</span>
          </button>
        </div>
      </template>
    </form>
  </section>
</template>
