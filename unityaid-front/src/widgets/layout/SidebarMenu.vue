<script setup lang="ts">
import { useRoute } from 'vue-router'
import {
  BadgeHelp,
  Building2,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Home,
  ListTodo,
  Network,
  Newspaper
} from 'lucide-vue-next'

defineProps<{ collapsed: boolean }>()
defineEmits<{ toggle: [] }>()
const route = useRoute()

const sections = [
  {
    title: 'Основное',
    items: [{ label: 'Главная страница', icon: Home, to: '/' }]
  },
  {
    title: 'Продуктивность',
    items: [
      { label: 'Мои задачи', icon: ListTodo, to: '/tasks' },
      { label: 'Мероприятия', icon: CalendarDays, to: '/calendar' }
    ]
  },
  {
    title: 'Организации',
    items: [
      { label: 'Организации', icon: Network, to: '/organizations' },
      { label: 'Новости', icon: Newspaper, to: '/news' }
    ]
  }
]

function isActive(path: string) {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-brand">
      <button class="icon-button ghost" type="button" :aria-label="collapsed ? 'Развернуть меню' : 'Скрыть меню'" @click="$emit('toggle')">
        <ChevronRight v-if="collapsed" :size="20" />
        <ChevronLeft v-else :size="20" />
      </button>
      <div v-if="!collapsed" class="logo-mark"><Building2 :size="22" /></div>
      <strong v-if="!collapsed">UnityAid</strong>
    </div>

    <nav class="sidebar-nav" aria-label="Основное меню">
      <section v-for="section in sections" :key="section.title || 'home'" class="nav-section">
        <h2 v-if="section.title && !collapsed">{{ section.title }}</h2>
        <RouterLink
          v-for="item in section.items"
          :key="item.label"
          :to="item.to"
          :class="['nav-link', { 'is-active': isActive(item.to) }]"
        >
          <component :is="item.icon" :size="18" />
          <span v-if="!collapsed">{{ item.label }}</span>
        </RouterLink>
      </section>
    </nav>

    <div v-if="!collapsed" class="sidebar-faq">
      <BadgeHelp :size="18" />
      <span>FAQ и помощь</span>
    </div>
  </aside>
</template>
