import { createRouter, createWebHistory } from 'vue-router'
import { authState, fetchCurrentUser } from '../entities/auth/store'
import { hasSystemAdminRole } from '../shared/permissions'
import AppLayout from '../widgets/layout/AppLayout.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import InviteRegisterPage from '../pages/InviteRegisterPage.vue'
import NewsDetailPage from '../pages/NewsDetailPage.vue'
import NewsFormPage from '../pages/NewsFormPage.vue'
import NewsListPage from '../pages/NewsListPage.vue'
import NotificationsPage from '../pages/NotificationsPage.vue'
import OnboardingPage from '../pages/OnboardingPage.vue'
import OrganizationsPage from '../pages/OrganizationsPage.vue'
import ProfilePage from '../pages/ProfilePage.vue'
import EventsPage from '../pages/EventsPage.vue'
import KnowledgeArticlePage from '../pages/KnowledgeArticlePage.vue'
import KnowledgeBasePage from '../pages/KnowledgeBasePage.vue'
import KnowledgeFormPage from '../pages/KnowledgeFormPage.vue'
import TasksPage from '../pages/TasksPage.vue'
import EventDetailPage from '../pages/EventDetailPage.vue'
import FaqPage from '../pages/FaqPage.vue'
import AchievementsPage from '../pages/AchievementsPage.vue'
import AnalyticsPage from '../pages/AnalyticsPage.vue'
import AdminPanelPage from '../pages/AdminPanelPage.vue'
import CertificatesPage from '../pages/CertificatesPage.vue'
import CertificateVerifyPage from '../pages/CertificateVerifyPage.vue'
import OrganizationDetailPage from '../pages/OrganizationDetailPage.vue'
import TaskDetailPage from '../pages/TaskDetailPage.vue'
import TimeEntriesPage from '../pages/TimeEntriesPage.vue'
import VolunteersPage from '../pages/VolunteersPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'
import ProfileFieldsSettingsPage from '../pages/ProfileFieldsSettingsPage.vue'
import ShopPage from '../pages/ShopPage.vue'

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
      path: '/register',
      name: 'invite-register',
      component: InviteRegisterPage,
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
          path: 'profile/:userId',
          name: 'profile-user',
          component: ProfilePage
        },
        {
          path: 'faq',
          name: 'faq',
          component: FaqPage
        },
        {
          path: 'onboarding',
          name: 'onboarding',
          component: OnboardingPage,
          meta: { roles: ['super_admin', 'org_admin'] }
        },
        {
          path: 'settings',
          name: 'settings',
          component: SettingsPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'settings/:section',
          name: 'settings-section',
          component: SettingsPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'settings/profile-fields/manage',
          name: 'profile-fields-settings',
          component: ProfileFieldsSettingsPage,
          meta: { roles: ['super_admin', 'org_admin'] }
        },
        {
          path: 'admin',
          name: 'admin',
          component: AdminPanelPage,
          meta: { roles: ['super_admin'] }
        },
        {
          path: 'organizations',
          name: 'organizations',
          component: OrganizationsPage
        },
        {
          path: 'organizations/:id',
          name: 'organization-detail',
          component: OrganizationDetailPage
        },
        {
          path: 'volunteers',
          name: 'volunteers',
          component: VolunteersPage
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
          path: 'calendar/:id',
          name: 'event-detail',
          component: EventDetailPage
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: TasksPage
        },
        {
          path: 'tasks/:id',
          name: 'task-detail',
          component: TaskDetailPage
        },
        {
          path: 'time-entries',
          name: 'time-entries',
          component: TimeEntriesPage
        },
        {
          path: 'achievements',
          name: 'achievements',
          component: AchievementsPage
        },
        {
          path: 'shop',
          name: 'shop',
          component: ShopPage
        },
        {
          path: 'certificates',
          name: 'certificates',
          component: CertificatesPage
        },
        {
          path: 'certificates/verify/:code?',
          name: 'certificate-verify',
          component: CertificateVerifyPage,
          meta: { public: true }
        },
        {
          path: 'notifications',
          name: 'notifications',
          component: NotificationsPage
        },
        {
          path: 'knowledge-base',
          name: 'knowledge-base',
          component: KnowledgeBasePage
        },
        {
          path: 'knowledge-base/new',
          name: 'knowledge-base-new',
          component: KnowledgeFormPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'knowledge-base/:id/edit',
          name: 'knowledge-base-edit',
          component: KnowledgeFormPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'knowledge-base/:id',
          name: 'knowledge-base-detail',
          component: KnowledgeArticlePage
        },
        {
          path: 'analytics',
          name: 'analytics',
          component: AnalyticsPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'analytics/:report',
          name: 'analytics-report',
          component: AnalyticsPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'news',
          name: 'news',
          component: NewsListPage
        },
        {
          path: 'news/new',
          name: 'news-new',
          component: NewsFormPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'news/:id/edit',
          name: 'news-edit',
          component: NewsFormPage,
          meta: { roles: ['super_admin', 'org_admin', 'coordinator'] }
        },
        {
          path: 'news/:id',
          name: 'news-detail',
          component: NewsDetailPage
        }
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

  const roles = to.meta.roles as string[] | undefined
  if (roles?.length && !roles.includes(authState.user?.primaryRole ?? '') && !hasSystemAdminRole(authState.user)) {
    return { name: 'dashboard' }
  }

  if (to.name === 'login' && authState.user) {
    return { name: 'dashboard' }
  }

  return true
})
