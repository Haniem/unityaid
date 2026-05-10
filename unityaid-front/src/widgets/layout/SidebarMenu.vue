<script setup lang="ts">
import { useRoute } from 'vue-router'
import {
  BadgeHelp,
  BarChart3,
  BookOpen,
  Building2,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  GraduationCap,
  Home,
  Lightbulb,
  ListTodo,
  Network,
  Newspaper,
  Send,
  Store,
  Trophy,
  Users,
  WalletCards
} from 'lucide-vue-next'

defineProps<{ collapsed: boolean }>()
defineEmits<{ toggle: [] }>()
const route = useRoute()

const sections = [
  {
    title: '',
    items: [{ label: 'Главная страница', icon: Home, to: '/' }]
  },
  {
    title: 'Избранное',
    items: [{ label: 'Моя команда', icon: Trophy, to: '/team' }]
  },
  {
    title: 'Продуктивность',
    items: [
      { label: 'Мои активности', icon: BarChart3, to: '/activities' },
      { label: 'Мои задачи', icon: ListTodo, to: '/tasks' },
      { label: 'Мероприятия', icon: CalendarDays, to: '/calendar' },
      { label: 'Мое обучение', icon: GraduationCap, to: '/learning' },
      { label: 'Мое развитие', icon: WalletCards, to: '/growth' }
    ]
  },
  {
    title: 'Организации',
    items: [
      { label: 'Организации', icon: Network, to: '/organizations' },
      { label: 'Сотрудники', icon: Users, to: '/employees' },
      { label: 'Новости', icon: Newspaper, to: '/news' },
      { label: 'База знаний', icon: BookOpen, to: '/knowledge-base' },
      { label: 'Есть идея', icon: Lightbulb, to: '/ideas' },
      { label: 'Корпоративный магазин', icon: Store, to: '/store' }
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
          <Send v-if="!collapsed && item.label.includes('активности')" class="nav-tail" :size="14" />
        </RouterLink>
      </section>
    </nav>

    <div v-if="!collapsed" class="sidebar-faq">
      <BadgeHelp :size="18" />
      <span>FAQ и помощь</span>
    </div>
  </aside>
</template>
