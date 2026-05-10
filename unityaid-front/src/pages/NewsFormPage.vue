<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ImagePlus, Save } from 'lucide-vue-next'
import { createNews, fetchNewsItem, updateNews, uploadNewsImage } from '../entities/news/api'

const route = useRoute()
const router = useRouter()

const editorRef = ref<HTMLElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const title = ref('')
const summary = ref('')
const contentHtml = ref('')
const coverImageUrl = ref<string | null>(null)
const status = ref<'draft' | 'published'>('published')
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')

const newsId = computed(() => (typeof route.params.id === 'string' ? route.params.id : null))
const isEdit = computed(() => Boolean(newsId.value))

async function loadNews() {
  if (!newsId.value) {
    return
  }

  isLoading.value = true
  try {
    const response = await fetchNewsItem(newsId.value)
    title.value = response.item.title
    summary.value = response.item.summary
    contentHtml.value = response.item.contentHtml
    coverImageUrl.value = response.item.coverImageUrl ?? null
    status.value = response.item.status
    await nextTick()
    if (editorRef.value) {
      editorRef.value.innerHTML = contentHtml.value
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить новость'
  } finally {
    isLoading.value = false
  }
}

function syncEditor() {
  contentHtml.value = editorRef.value?.innerHTML ?? ''
}

function format(command: string) {
  document.execCommand(command)
  syncEditor()
}

function openImageDialog() {
  fileInputRef.value?.click()
}

async function uploadImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  try {
    const response = await uploadNewsImage(file)
    if (!coverImageUrl.value) {
      coverImageUrl.value = response.url
    }
    document.execCommand('insertImage', false, response.url)
    syncEditor()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить изображение'
  } finally {
    input.value = ''
  }
}

async function submit() {
  syncEditor()
  errorMessage.value = ''
  isSaving.value = true

  try {
    const payload = {
      title: title.value,
      summary: summary.value,
      contentHtml: contentHtml.value,
      coverImageUrl: coverImageUrl.value,
      status: status.value
    }

    const response = isEdit.value && newsId.value
      ? await updateNews(newsId.value, payload)
      : await createNews(payload)

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
        <p class="eyebrow">Новости</p>
        <h1>{{ isEdit ? 'Редактирование новости' : 'Создание новости' }}</h1>
      </div>
    </div>

    <form class="news-form" @submit.prevent="submit">
      <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
      <div v-if="isLoading" class="empty-state">Загрузка новости...</div>

      <template v-else>
        <label>
          <span>Заголовок</span>
          <input v-model="title" type="text" required />
        </label>

        <label>
          <span>Краткое описание</span>
          <textarea v-model="summary" rows="3" />
        </label>

        <label>
          <span>Статус</span>
          <select v-model="status">
            <option value="published">Опубликовано</option>
            <option value="draft">Черновик</option>
          </select>
        </label>

        <div class="editor-shell">
          <div class="editor-toolbar">
            <button type="button" @click="format('bold')"><strong>B</strong></button>
            <button type="button" @click="format('italic')"><em>I</em></button>
            <button type="button" @click="format('insertUnorderedList')">•</button>
            <button type="button" @click="openImageDialog">
              <ImagePlus :size="17" />
            </button>
            <input ref="fileInputRef" type="file" accept="image/*" hidden @change="uploadImage" />
          </div>
          <div
            ref="editorRef"
            class="wysiwyg-editor"
            contenteditable="true"
            data-placeholder="Введите текст новости. Можно добавить списки, выделение и изображения."
            @input="syncEditor"
          ></div>
        </div>

        <div v-if="coverImageUrl" class="cover-preview">
          <span>Обложка новости</span>
          <img :src="coverImageUrl" alt="" />
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
