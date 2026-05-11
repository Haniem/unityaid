<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from 'vue-router'
import {
  Bell,
  Building2,
  CalendarDays,
  Clock3,
  HelpCircle,
  Home,
  ListTodo,
  LogOut,
  Menu,
  MessageCircle,
  Network,
  Newspaper,
  Plus,
  Search,
  Settings,
  UserRound,
  UsersRound
} from 'lucide-vue-next'
import { authState, logoutRemote } from '../../entities/auth/store'

const router = useRouter()
const route = useRoute()
const isUserMenuOpen = ref(false)
const isNavigationMenuOpen = ref(false)
const userPopoverRef = ref<HTMLElement | null>(null)
const navigationPopoverRef = ref<HTMLElement | null>(null)
const avatarRef = ref<HTMLElement | null>(null)
const navigationButtonRef = ref<HTMLElement | null>(null)

const navigationItems = [
  { label: 'Главная', icon: Home, to: '/' },
  { label: 'Задачи', icon: ListTodo, to: '/tasks' },
  { label: 'События', icon: CalendarDays, to: '/calendar' },
  { label: 'Мои часы', icon: Clock3, to: '/time-entries' },
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
  if (
    isUserMenuOpen.value &&
    !userPopoverRef.value?.contains(target) &&
    !avatarRef.value?.contains(target)
  ) {
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
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value
  isNavigationMenuOpen.value = false
}

async function signOut() {
  await logoutRemote()
  await router.push('/login')
}

onMounted(() => document.addEventListener('mousedown', closeOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <RouterLink class="topbar-brand" to="/">
        <span class="logo-mark"><Building2 :size="22" /></span>
        <span>
          <strong>UnityAid</strong>
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
      <button class="icon-button" type="button" aria-label="Сообщения"><MessageCircle :size="19" /></button>
      <button class="icon-button" type="button" aria-label="Уведомления">
        <Bell :size="19" />
        <span class="badge">57</span>
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
