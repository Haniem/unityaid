<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  Award,
  BarChart3,
  Bell,
  BookOpen,
  Building2,
  CalendarDays,
  ChevronRight,
  Database,
  FileArchive,
  FileCheck2,
  FolderCog,
  ListTodo,
  Save,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  UsersRound
} from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import { fetchSystemRoles, fetchUsers, updateUserSystemRoles } from '../entities/users/api'
import type { SystemRole, User } from '../entities/users/types'

type SettingsSection = {
  id: string
  title: string
  description: string
  icon: object
  access: 'client-admin' | 'manager' | 'super-admin'
  items: string[]
}

const route = useRoute()
const users = ref<User[]>([])
const roles = ref<SystemRole[]>([])
const search = ref('')
const selectedUserId = ref('')
const selectedRoleIds = ref<string[]>([])
const loadError = ref('')
const saveMessage = ref('')
const isSaving = ref(false)

const isSuperAdmin = computed(() => authState.user?.primaryRole === 'super_admin')
const isClientAdmin = computed(() => ['super_admin', 'org_admin'].includes(authState.user?.primaryRole ?? ''))
const canManageSettings = computed(() => ['super_admin', 'org_admin', 'coordinator'].includes(authState.user?.primaryRole ?? ''))

const sections: SettingsSection[] = [
  {
    id: 'organizations',
    title: 'Организации',
    description: 'Подразделения, филиалы, публичные данные, контакты и правила членства.',
    icon: Building2,
    access: 'client-admin',
    items: ['Филиалы', 'Контакты', 'Логотипы', 'Правила членства']
  },
  {
    id: 'users-roles',
    title: 'Пользователи и роли',
    description: 'Приглашения, роли, политики регистрации и статусы волонтеров.',
    icon: UsersRound,
    access: 'client-admin',
    items: ['Приглашения', 'Роли', 'Права доступа', 'Статусы волонтеров']
  },
  {
    id: 'events',
    title: 'Мероприятия и заявки',
    description: 'Типы мероприятий, причины отказа, шаблоны и правила обработки заявок.',
    icon: CalendarDays,
    access: 'manager',
    items: ['Типы мероприятий', 'Поля заявки', 'Шаблоны', 'Автообработка']
  },
  {
    id: 'tasks',
    title: 'Задачи',
    description: 'Статусы, приоритеты, типовые задачи, просрочки и уведомления.',
    icon: ListTodo,
    access: 'manager',
    items: ['Статусы', 'Приоритеты', 'Типовые задачи', 'Просрочки']
  },
  {
    id: 'notifications',
    title: 'Уведомления',
    description: 'Типы уведомлений, каналы, шаблоны сообщений и напоминания.',
    icon: Bell,
    access: 'manager',
    items: ['In-app', 'Email', 'Шаблоны', 'Напоминания']
  },
  {
    id: 'certificates',
    title: 'Сертификаты и документы',
    description: 'PDF-шаблоны, подписи, нумерация, публичная проверка и согласия.',
    icon: FileCheck2,
    access: 'client-admin',
    items: ['Шаблоны PDF', 'Подписи', 'Нумерация', 'Согласия']
  },
  {
    id: 'content',
    title: 'База знаний и новости',
    description: 'Категории, права публикации, шаблоны материалов и архивирование.',
    icon: BookOpen,
    access: 'manager',
    items: ['Категории', 'Права публикации', 'Шаблоны', 'Архив']
  },
  {
    id: 'gamification',
    title: 'Геймификация',
    description: 'Правила начисления, достижения, уровни и лидерборд.',
    icon: Award,
    access: 'client-admin',
    items: ['Правила', 'Достижения', 'Уровни', 'Лидерборд']
  },
  {
    id: 'analytics',
    title: 'Аналитика и отчеты',
    description: 'KPI, сохраненные фильтры, отправка отчетов и доступ по ролям.',
    icon: BarChart3,
    access: 'client-admin',
    items: ['KPI', 'Фильтры', 'Рассылки', 'Доступы']
  },
  {
    id: 'files',
    title: 'Файлы',
    description: 'Ограничения загрузки, очистка неиспользуемых файлов и правила хранения.',
    icon: FileArchive,
    access: 'client-admin',
    items: ['Лимиты', 'Типы файлов', 'Очистка', 'Хранение']
  },
  {
    id: 'system-admin',
    title: 'Системная админ-панель',
    description: 'Технический CRUD, audit log и системные справочники для super admin.',
    icon: Database,
    access: 'super-admin',
    items: ['CRUD', 'Audit log', 'Справочники', 'Диагностика']
  }
]

const currentSectionId = computed(() => {
  const value = route.params.section
  return Array.isArray(value) ? value[0] : value
})

const availableSections = computed(() => sections.filter((section) => hasAccess(section)))
const selectedSection = computed(() => sections.find((section) => section.id === currentSectionId.value) || null)
const selectedSectionAvailable = computed(() => selectedSection.value ? hasAccess(selectedSection.value) : true)

const filteredUsers = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return users.value
  return users.value.filter((user) => {
    const name = `${user.lastName} ${user.firstName} ${user.email}`.toLowerCase()
    return name.includes(query)
  })
})

const selectedUser = computed(() => users.value.find((user) => user.id === selectedUserId.value) || null)

function hasAccess(section: SettingsSection) {
  if (section.access === 'super-admin') return isSuperAdmin.value
  if (section.access === 'client-admin') return isClientAdmin.value
  return canManageSettings.value
}

async function load() {
  if (!isSuperAdmin.value) return
  loadError.value = ''
  try {
    const [usersResponse, rolesResponse] = await Promise.all([fetchUsers(), fetchSystemRoles()])
    users.value = usersResponse.items
    roles.value = rolesResponse.items
    if (!selectedUserId.value && users.value.length) {
      selectUser(users.value[0])
    }
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : 'Не удалось загрузить настройки'
  }
}

function selectUser(user: User) {
  selectedUserId.value = user.id
  selectedRoleIds.value = user.systemRoles?.map((role) => role.id) ?? []
  saveMessage.value = ''
}

function toggleRole(roleId: string, checked: boolean) {
  selectedRoleIds.value = checked
    ? [...new Set([...selectedRoleIds.value, roleId])]
    : selectedRoleIds.value.filter((item) => item !== roleId)
}

async function saveRoles() {
  if (!selectedUser.value) return
  isSaving.value = true
  saveMessage.value = ''
  try {
    const response = await updateUserSystemRoles(selectedUser.value.id, selectedRoleIds.value)
    const index = users.value.findIndex((user) => user.id === selectedUser.value?.id)
    if (index >= 0) {
      users.value[index] = { ...users.value[index], systemRoles: response.items }
    }
    selectedRoleIds.value = response.items.map((role) => role.id)
    saveMessage.value = 'Системные роли обновлены'
  } catch (error) {
    saveMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить роли'
  } finally {
    isSaving.value = false
  }
}

function roleNames(user: User) {
  const items = user.systemRoles || []
  return items.length ? items.map((role) => role.name).join(', ') : 'Роли не назначены'
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Настройки</p>
        <h1>{{ selectedSection ? selectedSection.title : 'Администрирование клиента' }}</h1>
        <p>
          {{
            selectedSection
              ? selectedSection.description
              : 'Единая точка входа в настройки модулей, ролей, документов, отчетов и клиентского пространства.'
          }}
        </p>
      </div>
      <RouterLink v-if="selectedSection" class="secondary-action" to="/settings">
        <FolderCog :size="17" />
        <span>Все настройки</span>
      </RouterLink>
    </div>

    <div v-if="!selectedSection" class="settings-catalog">
      <RouterLink
        v-for="section in availableSections"
        :key="section.id"
        class="settings-card"
        :to="section.id === 'system-admin' ? '/admin' : `/settings/${section.id}`"
      >
        <span class="settings-card-icon">
          <component :is="section.icon" :size="24" />
        </span>
        <span class="settings-card-body">
          <strong>{{ section.title }}</strong>
          <small>{{ section.description }}</small>
          <span class="settings-chip-list">
            <span v-for="item in section.items.slice(0, 3)" :key="item">{{ item }}</span>
          </span>
        </span>
        <ChevronRight :size="20" />
      </RouterLink>
    </div>

    <div v-else-if="!selectedSectionAvailable" class="empty-state">
      У вас нет доступа к этому разделу настроек.
    </div>

    <template v-else-if="selectedSection.id === 'users-roles'">
      <div class="settings-module-shell">
        <div class="detail-panel">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Модуль</p>
              <h2>Пользователи и роли</h2>
            </div>
          </div>
          <p class="settings-module-description">
            Здесь будут настройки приглашений, клиентских ролей, политик регистрации и статусов волонтеров.
            Управление системными ролями доступно только super admin.
          </p>
        </div>

        <p v-if="loadError" class="form-error">{{ loadError }}</p>

        <div v-if="isSuperAdmin" class="settings-layout">
          <aside class="detail-panel settings-users-panel">
            <div class="section-heading">
              <div>
                <p class="eyebrow">Пользователи</p>
                <h2>Аккаунты</h2>
              </div>
            </div>
            <label class="search-field settings-search">
              <Search :size="17" />
              <input v-model="search" type="search" placeholder="Поиск по имени или email" />
            </label>
            <div class="settings-user-list">
              <button
                v-for="user in filteredUsers"
                :key="user.id"
                :class="['settings-user-row', { active: user.id === selectedUserId }]"
                type="button"
                @click="selectUser(user)"
              >
                <span>
                  <strong>{{ user.lastName }} {{ user.firstName }}</strong>
                  <small>{{ user.email }}</small>
                </span>
                <small>{{ roleNames(user) }}</small>
              </button>
            </div>
          </aside>

          <section class="detail-panel settings-roles-panel">
            <div class="section-heading">
              <div>
                <p class="eyebrow">Права в системе</p>
                <h2>{{ selectedUser ? `${selectedUser.lastName} ${selectedUser.firstName}` : 'Выберите пользователя' }}</h2>
              </div>
              <SlidersHorizontal :size="22" />
            </div>

            <div v-if="selectedUser" class="system-role-list">
              <label v-for="role in roles" :key="role.id" class="system-role-option">
                <input type="checkbox" :checked="selectedRoleIds.includes(role.id)" @change="toggleRole(role.id, ($event.target as HTMLInputElement).checked)" />
                <span>
                  <strong>{{ role.name }}</strong>
                  <small>{{ role.description }}</small>
                </span>
                <ShieldCheck :size="18" />
              </label>
            </div>

            <div class="form-actions">
              <button class="primary-action" type="button" :disabled="!selectedUser || isSaving" @click="saveRoles">
                <Save :size="17" />
                <span>{{ isSaving ? 'Сохранение...' : 'Сохранить' }}</span>
              </button>
            </div>
            <p v-if="saveMessage" class="form-success">{{ saveMessage }}</p>
          </section>
        </div>
      </div>
    </template>

    <div v-else class="settings-module-shell">
      <section class="detail-panel settings-module-panel">
        <span class="settings-card-icon">
          <component :is="selectedSection.icon" :size="26" />
        </span>
        <div>
          <p class="eyebrow">Раздел настроек</p>
          <h2>{{ selectedSection.title }}</h2>
          <p>{{ selectedSection.description }}</p>
        </div>
      </section>

      <section class="detail-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Будущие настройки</p>
            <h2>Что будет настраиваться</h2>
          </div>
        </div>
        <div class="settings-chip-list large">
          <span v-for="item in selectedSection.items" :key="item">{{ item }}</span>
        </div>
      </section>
    </div>
  </section>
</template>
