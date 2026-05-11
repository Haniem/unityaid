<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Bell, CheckCheck } from 'lucide-vue-next'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '../entities/notifications/api'
import type { NotificationItem } from '../entities/notifications/types'

const items = ref<NotificationItem[]>([])
const unreadCount = ref(0)
const isLoading = ref(false)
const errorMessage = ref('')

async function load() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchNotifications()
    items.value = response.items
    unreadCount.value = response.unreadCount
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить уведомления'
  } finally {
    isLoading.value = false
  }
}

async function markRead(item: NotificationItem) {
  if (item.isRead) return
  await markNotificationRead(item.id)
  await load()
}

async function markAll() {
  await markAllNotificationsRead()
  await load()
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Центр уведомлений</p>
        <h1>Уведомления</h1>
        <p>Заявки, задачи, достижения и важные напоминания по мероприятиям.</p>
      </div>
      <button class="secondary-action" type="button" :disabled="unreadCount === 0" @click="markAll">
        <CheckCheck :size="18" />
        Прочитать все
      </button>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <div v-else-if="isLoading" class="empty-state">Загрузка уведомлений...</div>
    <div v-else-if="items.length === 0" class="empty-state">Пока нет уведомлений.</div>

    <div v-else class="notification-list">
      <article
        v-for="item in items"
        :key="item.id"
        :class="['notification-card', { unread: !item.isRead }]"
      >
        <span class="notification-icon"><Bell :size="18" /></span>
        <div>
          <RouterLink v-if="item.link" :to="item.link" @click="markRead(item)">
            <strong>{{ item.title }}</strong>
          </RouterLink>
          <strong v-else>{{ item.title }}</strong>
          <p>{{ item.body }}</p>
          <small>{{ formatDate(item.createdAt) }}</small>
        </div>
        <button v-if="!item.isRead" class="secondary-action" type="button" @click="markRead(item)">
          Прочитано
        </button>
      </article>
    </div>
  </section>
</template>
