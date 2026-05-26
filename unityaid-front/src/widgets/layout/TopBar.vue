<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Award,
  BarChart3,
  Bell,
  BookOpen,
  Building2,
  CalendarDays,
  Clock3,
  FileCheck2,
  HelpCircle,
  Home,
  ListTodo,
  LogOut,
  Menu,
  Network,
  Newspaper,
  Plus,
  Search,
  Settings,
  UserRound,
  UsersRound
} from 'lucide-vue-next'
import { authState, logoutRemote } from '../../entities/auth/store'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '../../entities/notifications/api'
import type { NotificationItem } from '../../entities/notifications/types'

const router = useRouter()
const route = useRoute()
const isUserMenuOpen = ref(false)
const isNavigationMenuOpen = ref(false)
const isNotificationsOpen = ref(false)
const notifications = ref<NotificationItem[]>([])
const unreadCount = ref(0)
const userPopoverRef = ref<HTMLElement | null>(null)
const navigationPopoverRef = ref<HTMLElement | null>(null)
const notificationsPopoverRef = ref<HTMLElement | null>(null)
const avatarRef = ref<HTMLElement | null>(null)
const navigationButtonRef = ref<HTMLElement | null>(null)
const notificationsButtonRef = ref<HTMLElement | null>(null)

const navigationItems = [
  { label: 'Главная', icon: Home, to: '/' },
  { label: 'Задачи', icon: ListTodo, to: '/tasks' },
  { label: 'События', icon: CalendarDays, to: '/calendar' },
  { label: 'Мои часы', icon: Clock3, to: '/time-entries' },
  { label: 'Достижения', icon: Award, to: '/achievements' },
  { label: 'Сертификаты', icon: FileCheck2, to: '/certificates' },
  { label: 'База знаний', icon: BookOpen, to: '/knowledge-base' },
  { label: 'Аналитика', icon: BarChart3, to: '/analytics' },
  { label: 'Волонтеры', icon: UsersRound, to: '/volunteers' },
  { label: 'Организации', icon: Network, to: '/organizations' },
  { label: 'Новости', icon: Newspaper, to: '/news' }
]

const fullName = computed(() => {
  const user = authState.user
  if (!user) return ''
  return [user.lastName, user.firstName, user.patronymic].filter(Boolean).join(' ')
})

function closeOutside(event: MouseEvent) {
  const target = event.target as Node
  if (isUserMenuOpen.value && !userPopoverRef.value?.contains(target) && !avatarRef.value?.contains(target)) {
    isUserMenuOpen.value = false
  }
  if (
    isNavigationMenuOpen.value &&
    !navigationPopoverRef.value?.contains(target) &&
    !navigationButtonRef.value?.contains(target)
  ) {
    isNavigationMenuOpen.value = false
  }
  if (
    isNotificationsOpen.value &&
    !notificationsPopoverRef.value?.contains(target) &&
    !notificationsButtonRef.value?.contains(target)
  ) {
    isNotificationsOpen.value = false
  }
}

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(`${path}/`)
}

function toggleNavigationMenu() {
  isNavigationMenuOpen.value = !isNavigationMenuOpen.value
  isUserMenuOpen.value = false
  isNotificationsOpen.value = false
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value
  isNavigationMenuOpen.value = false
  isNotificationsOpen.value = false
}

async function loadNotifications() {
  if (!authState.token) return
  try {
    const response = await fetchNotifications()
    notifications.value = response.items.slice(0, 5)
    unreadCount.value = response.unreadCount
  } catch {
    notifications.value = []
    unreadCount.value = 0
  }
}

async function toggleNotifications() {
  isNotificationsOpen.value = !isNotificationsOpen.value
  isUserMenuOpen.value = false
  isNavigationMenuOpen.value = false
  if (isNotificationsOpen.value) {
    await loadNotifications()
  }
}

async function readNotification(item: NotificationItem) {
  if (!item.isRead) {
    await markNotificationRead(item.id)
  }
  isNotificationsOpen.value = false
  await router.push(item.link || '/notifications')
  await loadNotifications()
}

async function readAllNotifications() {
  await markAllNotificationsRead()
  await loadNotifications()
}

function formatNotificationDate(value: string) {
  return new Date(value).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
}

async function signOut() {
  await logoutRemote()
  await router.push('/login')
}

onMounted(() => {
  document.addEventListener('mousedown', closeOutside)
  loadNotifications()
})
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <RouterLink class="topbar-brand" to="/">
        <span class="logo-mark"><Building2 :size="22" /></span>
        <span>
          <strong>Пульс</strong>
          <small>volunteer hub</small>
        </span>
      </RouterLink>

      <div class="topbar-search">
        <Search :size="18" />
        <input type="search" placeholder="Поиск" aria-label="Поиск" />
      </div>
    </div>

    <div class="topbar-actions">
      <button class="icon-button accent" type="button" aria-label="Создать"><Plus :size="20" /></button>
      <button
        ref="notificationsButtonRef"
        :class="['icon-button', { accent: isNotificationsOpen }]"
        type="button"
        aria-label="Уведомления"
        @click="toggleNotifications"
      >
        <Bell :size="19" />
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </button>
      <button class="icon-button" type="button" aria-label="FAQ"><HelpCircle :size="19" /></button>
      <button class="icon-button" type="button" aria-label="Настройки" @click="router.push('/settings')"><Settings :size="19" /></button>
      <button
        ref="navigationButtonRef"
        :class="['icon-button', { accent: isNavigationMenuOpen }]"
        type="button"
        aria-label="Основное меню"
        @click="toggleNavigationMenu"
      >
        <Menu :size="20" />
      </button>
      <button ref="avatarRef" class="avatar-button" type="button" aria-label="Меню пользователя" @click="toggleUserMenu">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
      </button>
    </div>

    <nav v-if="isNavigationMenuOpen" ref="navigationPopoverRef" class="navigation-popover" aria-label="Основное меню">
      <RouterLink
        v-for="item in navigationItems"
        :key="item.to"
        :to="item.to"
        :class="['navigation-popover-link', { active: isActive(item.to) }]"
        @click="isNavigationMenuOpen = false"
      >
        <component :is="item.icon" :size="18" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>

    <div v-if="isNotificationsOpen" ref="notificationsPopoverRef" class="notifications-popover">
      <div class="popover-header">
        <strong>Уведомления</strong>
        <button v-if="unreadCount > 0" type="button" @click="readAllNotifications">Прочитать все</button>
      </div>
      <button
        v-for="item in notifications"
        :key="item.id"
        :class="['notification-popover-item', { unread: !item.isRead }]"
        type="button"
        @click="readNotification(item)"
      >
        <span>{{ item.title }}</span>
        <small>{{ item.body }}</small>
        <time>{{ formatNotificationDate(item.createdAt) }}</time>
      </button>
      <RouterLink class="notification-popover-all" to="/notifications" @click="isNotificationsOpen = false">
        Все уведомления
      </RouterLink>
      <p v-if="notifications.length === 0" class="notification-empty">Нет новых уведомлений</p>
    </div>

    <div v-if="isUserMenuOpen" ref="userPopoverRef" class="user-popover">
      <div class="popover-user">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
        <strong>{{ fullName }}</strong>
      </div>

      <RouterLink class="popover-link" to="/profile" @click="isUserMenuOpen = false">
        <UserRound :size="17" />
        <span>Мой профиль</span>
      </RouterLink>

      <button class="popover-link danger" type="button" @click="signOut">
        <LogOut :size="17" />
        <span>Выйти из аккаунта</span>
      </button>
    </div>
  </header>
</template>
