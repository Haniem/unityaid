import { createRouter, createWebHistory } from 'vue-router'
import { authState, fetchCurrentUser } from '../entities/auth/store'
import AppLayout from '../widgets/layout/AppLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import NewsDetailPage from '../pages/NewsDetailPage.vue'
import NewsFormPage from '../pages/NewsFormPage.vue'
import NewsListPage from '../pages/NewsListPage.vue'
import OrganizationsPage from '../pages/OrganizationsPage.vue'
import PlaceholderPage from '../pages/PlaceholderPage.vue'
import ProfilePage from '../pages/ProfilePage.vue'
import EventsPage from '../pages/EventsPage.vue'
import TasksPage from '../pages/TasksPage.vue'

const placeholderRoutes = [
  ['team', 'Моя команда'],
  ['activities', 'Мои активности'],
  ['learning', 'Мое обучение'],
  ['growth', 'Мое развитие'],
  ['employees', 'Сотрудники'],
  ['knowledge-base', 'База знаний'],
  ['ideas', 'Есть идея'],
  ['store', 'Корпоративный магазин']
].map(([path, title]) => ({
  path,
  component: PlaceholderPage,
  meta: { title }
}))

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginPage,
      meta: { public: true }
    },
    {
      path: '/',
      component: AppLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: DashboardPage
        },
        {
          path: 'profile',
          name: 'profile',
          component: ProfilePage
        },
        {
          path: 'organizations',
          name: 'organizations',
          component: OrganizationsPage
        },
        {
          path: 'company-structure',
          redirect: '/organizations'
        },
        {
          path: 'company-pulse',
          redirect: '/organizations'
        },
        {
          path: 'requests',
          redirect: '/organizations'
        },
        {
          path: 'about-company',
          redirect: '/organizations'
        },
        {
          path: 'calendar',
          name: 'events',
          component: EventsPage
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: TasksPage
        },
        {
          path: 'news',
          name: 'news',
          component: NewsListPage
        },
        {
          path: 'news/new',
          name: 'news-new',
          component: NewsFormPage
        },
        {
          path: 'news/:id/edit',
          name: 'news-edit',
          component: NewsFormPage
        },
        {
          path: 'news/:id',
          name: 'news-detail',
          component: NewsDetailPage
        },
        ...placeholderRoutes
      ]
    }
  ]
})

router.beforeEach(async (to) => {
  if (!authState.isReady) {
    await fetchCurrentUser()
  }

  if (!to.meta.public && !authState.user) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.public && authState.user) {
    return { name: 'dashboard' }
  }

  return true
})
