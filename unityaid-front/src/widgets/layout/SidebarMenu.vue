<script setup lang="ts">
import { useRoute } from 'vue-router'
import { Building2, CalendarDays, Home, ListTodo, Network, Newspaper, UsersRound } from 'lucide-vue-next'

defineProps<{ collapsed: boolean }>()
defineEmits<{ toggle: [] }>()
const route = useRoute()

const items = [
  { label: 'Главная', icon: Home, to: '/' },
  { label: 'Задачи', icon: ListTodo, to: '/tasks' },
  { label: 'События', icon: CalendarDays, to: '/calendar' },
  { label: 'Волонтеры', icon: UsersRound, to: '/volunteers' },
  { label: 'Организации', icon: Network, to: '/organizations' },
  { label: 'Новости', icon: Newspaper, to: '/news' }
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-brand">
      <div class="logo-mark"><Building2 :size="22" /></div>
      <div v-if="!collapsed">
        <strong>UnityAid</strong>
        <span>volunteer hub</span>
      </div>
    </div>

    <nav class="sidebar-nav compact" aria-label="Основное меню">
      <RouterLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        :class="['nav-link', { 'is-active': isActive(item.to) }]"
        :title="item.label"
      >
        <component :is="item.icon" :size="19" />
        <span v-if="!collapsed">{{ item.label }}</span>
      </RouterLink>
    </nav>
  </aside>
</template>
