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
  CheckCircle2,
  Clock3,
  FileCheck2,
  FileText,
  Gift,
  HelpCircle,
  Home,
  ListTodo,
  LogOut,
  Menu,
  Network,
  Newspaper,
  Plus,
  Settings,
  UserRound,
  UsersRound,
  X
} from 'lucide-vue-next'
import { authState, logoutRemote } from '../../entities/auth/store'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '../../entities/notifications/api'
import type { NotificationItem } from '../../entities/notifications/types'
import { tenantTheme } from '../../entities/tenantSettings/theme'
import { canManageContent } from '../../shared/permissions'

const router = useRouter()
const route = useRoute()
const isUserMenuOpen = ref(false)
const isNavigationMenuOpen = ref(false)
const isNotificationsOpen = ref(false)
const isCreateMenuOpen = ref(false)
const notificationTab = ref<'all' | 'unread'>('all')
const notifications = ref<NotificationItem[]>([])
const unreadCount = ref(0)
const userPopoverRef = ref<HTMLElement | null>(null)
const navigationPopoverRef = ref<HTMLElement | null>(null)
const createPopoverRef = ref<HTMLElement | null>(null)
const avatarRef = ref<HTMLElement | null>(null)
const navigationButtonRef = ref<HTMLElement | null>(null)
const notificationsButtonRef = ref<HTMLElement | null>(null)
const createButtonRef = ref<HTMLElement | null>(null)

const navigationItems = computed(() => [
  { label: 'Главная', icon: Home, to: '/' },
  { label: 'Задачи', icon: ListTodo, to: '/tasks' },
  { label: 'События', icon: CalendarDays, to: '/calendar' },
  { label: 'Мои часы', icon: Clock3, to: '/time-entries' },
  { label: 'Магазин', icon: Gift, to: '/shop' },
  { label: 'Достижения', icon: Award, to: '/achievements' },
  { label: 'Сертификаты', icon: FileCheck2, to: '/certificates' },
  { label: 'База знаний', icon: BookOpen, to: '/knowledge-base' },
  ...(canUseCreateActions.value ? [{ label: 'Аналитика', icon: BarChart3, to: '/analytics' }] : []),
  { label: 'Волонтеры', icon: UsersRound, to: '/volunteers' },
  { label: 'Организации', icon: Network, to: '/organizations' },
  { label: 'Новости', icon: Newspaper, to: '/news' }
])

const createActions = [
  { label: 'Создать мероприятие', icon: CalendarDays, to: '/calendar' },
  { label: 'Поставить задачу', icon: CheckCircle2, to: '/tasks' },
  { label: 'Опубликовать новость', icon: FileText, to: '/news/new' }
]

const fullName = computed(() => {
  const user = authState.user
  if (!user) return ''
  return [user.lastName, user.firstName].filter(Boolean).join(' ')
})

const userInitials = computed(() => {
  const user = authState.user
  if (!user) return 'П'
  return `${user.firstName?.[0] ?? ''}${user.lastName?.[0] ?? ''}`.toUpperCase() || 'П'
})

const canOpenSettings = computed(() => ['super_admin', 'org_admin', 'coordinator'].includes(authState.user?.primaryRole ?? ''))
const canUseCreateActions = computed(() => canManageContent(authState.user))

const visibleNotifications = computed(() => {
  if (notificationTab.value === 'unread') {
    return notifications.value.filter((item) => !item.isRead)
  }
  return notifications.value
})

function closeOutside(event: MouseEvent) {
  const target = event.target as Node
  if (isCreateMenuOpen.value && !createPopoverRef.value?.contains(target) && !createButtonRef.value?.contains(target)) {
    isCreateMenuOpen.value = false
  }
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
}

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(`${path}/`)
}

function toggleNavigationMenu() {
  isNavigationMenuOpen.value = !isNavigationMenuOpen.value
  isUserMenuOpen.value = false
  isNotificationsOpen.value = false
  isCreateMenuOpen.value = false
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value
  isNavigationMenuOpen.value = false
  isNotificationsOpen.value = false
  isCreateMenuOpen.value = false
}

function toggleCreateMenu() {
  if (!canUseCreateActions.value) return
  isCreateMenuOpen.value = !isCreateMenuOpen.value
  isUserMenuOpen.value = false
  isNavigationMenuOpen.value = false
  isNotificationsOpen.value = false
}

async function loadNotifications() {
  if (!authState.token) return
  try {
    const response = await fetchNotifications()
    notifications.value = response.items
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
  isCreateMenuOpen.value = false
  if (isNotificationsOpen.value) {
    notificationTab.value = 'all'
    await loadNotifications()
  }
}

async function openNotificationsFromUserMenu() {
  isUserMenuOpen.value = false
  isNavigationMenuOpen.value = false
  isCreateMenuOpen.value = false
  isNotificationsOpen.value = true
  notificationTab.value = 'all'
  await loadNotifications()
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

function closeNotifications() {
  isNotificationsOpen.value = false
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    isNotificationsOpen.value = false
    isCreateMenuOpen.value = false
    isNavigationMenuOpen.value = false
    isUserMenuOpen.value = false
  }
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
  document.addEventListener('keydown', handleKeydown)
  loadNotifications()
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <RouterLink class="topbar-brand" to="/">
        <span class="logo-mark">
          <img v-if="tenantTheme.logoUrl" :src="tenantTheme.logoUrl" alt="" />
          <Building2 v-else :size="22" />
        </span>
        <span>
          <strong>{{ tenantTheme.displayName }}</strong>
          <small>volunteer hub</small>
        </span>
      </RouterLink>

    </div>

    <div class="topbar-actions">
      <button v-if="canUseCreateActions" ref="createButtonRef" class="icon-button accent topbar-secondary-action" type="button" aria-label="Создать" @click="toggleCreateMenu"><Plus :size="20" /></button>
      <button
        ref="notificationsButtonRef"
        :class="['icon-button topbar-secondary-action', { accent: isNotificationsOpen }]"
        type="button"
        aria-label="Уведомления"
        @click="toggleNotifications"
      >
        <Bell :size="19" />
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </button>
      <button class="icon-button topbar-secondary-action" type="button" aria-label="FAQ" @click="router.push('/faq')"><HelpCircle :size="19" /></button>
      <button v-if="canOpenSettings" class="icon-button topbar-secondary-action" type="button" aria-label="Настройки" @click="router.push('/settings')"><Settings :size="19" /></button>
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
        <img v-if="authState.user?.avatarUrl" :src="authState.user.avatarUrl" alt="" />
        <span v-else class="avatar-fallback compact">{{ userInitials }}</span>
      </button>
    </div>

    <Transition name="popover-scale">
      <div v-if="isCreateMenuOpen && canUseCreateActions" ref="createPopoverRef" class="create-popover">
        <RouterLink
          v-for="action in createActions"
          :key="action.to"
          class="create-popover-link"
          :to="action.to"
          @click="isCreateMenuOpen = false"
        >
          <component :is="action.icon" :size="18" />
          <span>{{ action.label }}</span>
        </RouterLink>
      </div>
    </Transition>

    <Transition name="popover-scale">
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
    </Transition>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="isNotificationsOpen" class="modal-backdrop notifications-backdrop" @click.self="closeNotifications">
          <section class="modal-panel notifications-modal" role="dialog" aria-modal="true" aria-label="Уведомления">
          <div class="modal-heading">
            <div>
              <p class="eyebrow">Центр уведомлений</p>
              <h2>Уведомления</h2>
            </div>
            <div class="modal-heading-actions">
              <button class="secondary-action" type="button" :disabled="unreadCount === 0" @click="readAllNotifications">
                Прочитать все
              </button>
              <button class="icon-button" type="button" aria-label="Закрыть уведомления" @click="closeNotifications">
                <X :size="18" />
              </button>
            </div>
          </div>

          <div class="segmented-control notification-tabs" aria-label="Фильтр уведомлений">
            <button :class="{ active: notificationTab === 'all' }" type="button" @click="notificationTab = 'all'">Все</button>
            <button :class="{ active: notificationTab === 'unread' }" type="button" @click="notificationTab = 'unread'">Не прочитанные</button>
          </div>

          <div v-if="visibleNotifications.length > 0" class="notification-modal-list">
            <button
              v-for="item in visibleNotifications"
              :key="item.id"
              :class="['notification-popover-item', { unread: !item.isRead }]"
              type="button"
              @click="readNotification(item)"
            >
              <span>{{ item.title }}</span>
              <small>{{ item.body }}</small>
              <time>{{ formatNotificationDate(item.createdAt) }}</time>
            </button>
          </div>
          <p v-else class="notification-empty">
            {{ notificationTab === 'unread' ? 'Непрочитанных уведомлений нет.' : 'Уведомлений пока нет.' }}
          </p>
          </section>
        </div>
      </Transition>
    </Teleport>

    <Transition name="popover-scale">
      <div v-if="isUserMenuOpen" ref="userPopoverRef" class="user-popover">
        <div class="popover-user">
          <img v-if="authState.user?.avatarUrl" :src="authState.user.avatarUrl" alt="" />
          <span v-else class="avatar-fallback compact">{{ userInitials }}</span>
          <strong>{{ fullName }}</strong>
        </div>

        <RouterLink class="popover-link" to="/profile" @click="isUserMenuOpen = false">
          <UserRound :size="17" />
          <span>Мой профиль</span>
        </RouterLink>

        <div class="mobile-user-actions">
          <p>Быстрые действия</p>
          <template v-if="canUseCreateActions">
            <RouterLink
              v-for="action in createActions"
              :key="action.to"
              class="popover-link"
              :to="action.to"
              @click="isUserMenuOpen = false"
            >
              <component :is="action.icon" :size="17" />
              <span>{{ action.label }}</span>
            </RouterLink>
          </template>
          <button class="popover-link" type="button" @click="openNotificationsFromUserMenu">
            <Bell :size="17" />
            <span>Уведомления</span>
            <span v-if="unreadCount > 0" class="badge inline-badge">{{ unreadCount }}</span>
          </button>
          <RouterLink class="popover-link" to="/faq" @click="isUserMenuOpen = false">
            <HelpCircle :size="17" />
            <span>FAQ</span>
          </RouterLink>
          <RouterLink v-if="canOpenSettings" class="popover-link" to="/settings" @click="isUserMenuOpen = false">
            <Settings :size="17" />
            <span>Настройки</span>
          </RouterLink>
        </div>

        <button class="popover-link danger" type="button" @click="signOut">
          <LogOut :size="17" />
          <span>Выйти из аккаунта</span>
        </button>
      </div>
    </Transition>
  </header>
</template>
