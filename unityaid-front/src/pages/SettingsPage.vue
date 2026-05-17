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
const userSearch = ref('')
const settingsSearch = ref('')
const selectedUserId = ref('')
const selectedRoleIds = ref<string[]>([])
const loadError = ref('')
const saveMessage = ref('')
const moduleSaveMessage = ref('')
const isSaving = ref(false)
const organizationSettings = ref({
  displayName: 'UnityAid',
  contactEmail: 'hello@unityaid.test',
  timezone: 'Asia/Yekaterinburg',
  defaultBranch: 'Главное подразделение',
  membershipPolicy: 'manual'
})
const userRoleSettings = ref({
  inviteMode: 'link',
  defaultRole: 'volunteer',
  registrationPolicy: 'invite_only',
  defaultVolunteerStatus: 'new',
  allowSelfRegistration: false
})
const eventSettings = ref({
  defaultFormat: 'offline',
  approvalMode: 'manual',
  waitlistEnabled: true,
  defaultCapacity: 25,
  cancellationReason: 'Нет свободных мест'
})
const taskSettings = ref({
  defaultStatus: 'created',
  defaultPriority: 'medium',
  reviewRequired: true,
  overdueHours: 24,
  overdueNotifyRole: 'coordinator'
})
const notificationSettings = ref({
  inAppEnabled: true,
  emailEnabled: false,
  reminderHours: 24,
  digestFrequency: 'weekly',
  escalationEnabled: true
})
const certificateSettings = ref({
  templateName: 'Базовый сертификат',
  numberingPrefix: 'UA',
  publicVerification: true,
  signatureRole: 'Руководитель программы',
  consentRequired: true
})
const contentSettings = ref({
  defaultNewsStatus: 'draft',
  defaultArticleStatus: 'draft',
  moderationRequired: true,
  archiveAfterDays: 365,
  allowCoordinatorsPublish: true
})
const gamificationSettings = ref({
  enabled: true,
  pointsPerHour: 10,
  pointsPerTask: 25,
  leaderboardEnabled: true,
  recalculationMode: 'manual'
})
const analyticsSettings = ref({
  activeVolunteerDays: 30,
  targetHours: 1000,
  targetEvents: 20,
  reportFrequency: 'monthly',
  exportAuditEnabled: true
})
const fileSettings = ref({
  maxImageMb: 5,
  maxAttachmentMb: 20,
  allowedTypes: 'jpg, png, pdf, docx, xlsx',
  cleanupEnabled: true,
  retentionDays: 365
})

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
const filteredSettingsSections = computed(() => {
  const query = settingsSearch.value.trim().toLowerCase()
  if (!query) return availableSections.value

  return availableSections.value.filter((section) => {
    const searchableText = [section.title, section.description, ...section.items].join(' ').toLowerCase()
    return searchableText.includes(query)
  })
})

const filteredUsers = computed(() => {
  const query = userSearch.value.trim().toLowerCase()
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

function saveModuleSettings(sectionTitle: string) {
  moduleSaveMessage.value = `Настройки раздела "${sectionTitle}" сохранены в черновике интерфейса`
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <nav v-if="selectedSection" class="settings-breadcrumbs" aria-label="Навигация настроек">
          <RouterLink to="/settings">Настройки</RouterLink>
          <ChevronRight :size="14" />
          <span>{{ selectedSection.title }}</span>
        </nav>
        <p v-else class="eyebrow">Настройки</p>
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

    <template v-if="!selectedSection">
      <div class="settings-toolbar">
        <label class="search-field">
          <Search :size="17" />
          <input v-model="settingsSearch" type="search" placeholder="Найти настройки модуля" />
        </label>
      </div>

      <div v-if="filteredSettingsSections.length > 0" class="settings-catalog">
      <RouterLink
        v-for="section in filteredSettingsSections"
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

      <p v-else class="empty-state">По этому запросу настроек не найдено.</p>
    </template>

    <div v-else-if="!selectedSectionAvailable" class="empty-state">
      У вас нет доступа к этому разделу настроек.
    </div>

    <template v-else-if="selectedSection.id === 'organizations'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <Building2 :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Организации</h2>
            <p>Базовые параметры клиентского пространства, филиалов и правил членства.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Организации')">
            <label>
              <span>Название пространства</span>
              <input v-model="organizationSettings.displayName" type="text" />
            </label>
            <label>
              <span>Контактный email</span>
              <input v-model="organizationSettings.contactEmail" type="email" />
            </label>
            <label>
              <span>Часовой пояс</span>
              <select v-model="organizationSettings.timezone">
                <option value="Asia/Yekaterinburg">Asia/Yekaterinburg</option>
                <option value="Europe/Moscow">Europe/Moscow</option>
                <option value="UTC">UTC</option>
              </select>
            </label>
            <label>
              <span>Подразделение по умолчанию</span>
              <input v-model="organizationSettings.defaultBranch" type="text" />
            </label>
            <label class="settings-form-wide">
              <span>Правило добавления участников</span>
              <select v-model="organizationSettings.membershipPolicy">
                <option value="manual">Только вручную администратором</option>
                <option value="invite">По приглашению</option>
                <option value="open">Открытая заявка с подтверждением</option>
              </select>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

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

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Пользователи и роли')">
            <label>
              <span>Режим приглашений</span>
              <select v-model="userRoleSettings.inviteMode">
                <option value="link">Одноразовая ссылка</option>
                <option value="email">Email-приглашение</option>
                <option value="manual">Ручное создание</option>
              </select>
            </label>
            <label>
              <span>Роль по умолчанию</span>
              <select v-model="userRoleSettings.defaultRole">
                <option value="volunteer">Волонтер</option>
                <option value="coordinator">Координатор</option>
                <option value="org_admin">Администратор организации</option>
              </select>
            </label>
            <label>
              <span>Политика регистрации</span>
              <select v-model="userRoleSettings.registrationPolicy">
                <option value="invite_only">Только по приглашению</option>
                <option value="moderated">Саморегистрация с модерацией</option>
                <option value="open">Открытая регистрация</option>
              </select>
            </label>
            <label>
              <span>Статус нового волонтера</span>
              <select v-model="userRoleSettings.defaultVolunteerStatus">
                <option value="new">Новый</option>
                <option value="active">Активный</option>
                <option value="unavailable">Недоступен</option>
              </select>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="userRoleSettings.allowSelfRegistration" type="checkbox" />
              <span>Разрешить самостоятельную регистрацию волонтеров</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>

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
              <input v-model="userSearch" type="search" placeholder="Поиск по имени или email" />
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

    <template v-else-if="selectedSection.id === 'events'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <CalendarDays :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Мероприятия и заявки</h2>
            <p>Правила создания мероприятий, обработки заявок, лимитов и листа ожидания.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Мероприятия и заявки')">
            <label>
              <span>Формат мероприятия по умолчанию</span>
              <select v-model="eventSettings.defaultFormat">
                <option value="offline">Офлайн</option>
                <option value="online">Онлайн</option>
                <option value="hybrid">Гибрид</option>
              </select>
            </label>
            <label>
              <span>Обработка заявок</span>
              <select v-model="eventSettings.approvalMode">
                <option value="manual">Ручное подтверждение</option>
                <option value="auto">Автоподтверждение при свободных местах</option>
                <option value="rules">По правилам клиента</option>
              </select>
            </label>
            <label>
              <span>Лимит участников по умолчанию</span>
              <input v-model.number="eventSettings.defaultCapacity" type="number" min="1" />
            </label>
            <label>
              <span>Причина отказа по умолчанию</span>
              <input v-model="eventSettings.cancellationReason" type="text" />
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="eventSettings.waitlistEnabled" type="checkbox" />
              <span>Включить лист ожидания при заполненном лимите</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'tasks'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <ListTodo :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Задачи</h2>
            <p>Статусы, приоритеты, подтверждение выполнения, просрочки и уведомления исполнителей.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Задачи')">
            <label>
              <span>Статус новой задачи</span>
              <select v-model="taskSettings.defaultStatus">
                <option value="created">Создана</option>
                <option value="assigned">Назначена</option>
                <option value="in_progress">В работе</option>
              </select>
            </label>
            <label>
              <span>Приоритет по умолчанию</span>
              <select v-model="taskSettings.defaultPriority">
                <option value="low">Низкий</option>
                <option value="medium">Средний</option>
                <option value="high">Высокий</option>
              </select>
            </label>
            <label>
              <span>Просрочка через, часов</span>
              <input v-model.number="taskSettings.overdueHours" type="number" min="1" />
            </label>
            <label>
              <span>Кого уведомлять о просрочке</span>
              <select v-model="taskSettings.overdueNotifyRole">
                <option value="coordinator">Координатора</option>
                <option value="org_admin">Администратора организации</option>
                <option value="assignee">Исполнителя</option>
              </select>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="taskSettings.reviewRequired" type="checkbox" />
              <span>Требовать подтверждение выполнения координатором</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'notifications'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <Bell :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Уведомления</h2>
            <p>Каналы доставки, напоминания, digest-рассылки и эскалации важных событий.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Уведомления')">
            <label>
              <span>Напоминать за, часов</span>
              <input v-model.number="notificationSettings.reminderHours" type="number" min="1" />
            </label>
            <label>
              <span>Сводка уведомлений</span>
              <select v-model="notificationSettings.digestFrequency">
                <option value="never">Не отправлять</option>
                <option value="daily">Ежедневно</option>
                <option value="weekly">Еженедельно</option>
              </select>
            </label>
            <label class="toggle-field">
              <input v-model="notificationSettings.inAppEnabled" type="checkbox" />
              <span>Включить in-app уведомления</span>
            </label>
            <label class="toggle-field">
              <input v-model="notificationSettings.emailEnabled" type="checkbox" />
              <span>Включить email-уведомления</span>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="notificationSettings.escalationEnabled" type="checkbox" />
              <span>Эскалировать просрочки задач и неподтвержденные часы</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'certificates'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <FileCheck2 :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Сертификаты и документы</h2>
            <p>Шаблоны PDF, подписи, нумерация, согласия и публичная проверка документов.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Сертификаты и документы')">
            <label>
              <span>Шаблон по умолчанию</span>
              <input v-model="certificateSettings.templateName" type="text" />
            </label>
            <label>
              <span>Префикс нумерации</span>
              <input v-model="certificateSettings.numberingPrefix" type="text" />
            </label>
            <label class="settings-form-wide">
              <span>Подписант</span>
              <input v-model="certificateSettings.signatureRole" type="text" />
            </label>
            <label class="toggle-field">
              <input v-model="certificateSettings.publicVerification" type="checkbox" />
              <span>Включить публичную проверку по коду</span>
            </label>
            <label class="toggle-field">
              <input v-model="certificateSettings.consentRequired" type="checkbox" />
              <span>Требовать согласие перед выдачей документа</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'content'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <BookOpen :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>База знаний и новости</h2>
            <p>Публикация материалов, модерация, статусы по умолчанию, категории и архивирование.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('База знаний и новости')">
            <label>
              <span>Статус новой новости</span>
              <select v-model="contentSettings.defaultNewsStatus">
                <option value="draft">Черновик</option>
                <option value="published">Опубликована</option>
                <option value="scheduled">Запланирована</option>
              </select>
            </label>
            <label>
              <span>Статус новой статьи</span>
              <select v-model="contentSettings.defaultArticleStatus">
                <option value="draft">Черновик</option>
                <option value="published">Опубликована</option>
                <option value="archived">Архив</option>
              </select>
            </label>
            <label>
              <span>Архивировать через, дней</span>
              <input v-model.number="contentSettings.archiveAfterDays" type="number" min="1" />
            </label>
            <label class="toggle-field">
              <input v-model="contentSettings.moderationRequired" type="checkbox" />
              <span>Требовать модерацию материалов</span>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="contentSettings.allowCoordinatorsPublish" type="checkbox" />
              <span>Разрешить координаторам публикацию без администратора</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'gamification'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <Award :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Геймификация</h2>
            <p>Правила начисления баллов, достижения, уровни и лидерборд волонтеров.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Геймификация')">
            <label>
              <span>Баллов за час</span>
              <input v-model.number="gamificationSettings.pointsPerHour" type="number" min="0" />
            </label>
            <label>
              <span>Баллов за задачу</span>
              <input v-model.number="gamificationSettings.pointsPerTask" type="number" min="0" />
            </label>
            <label>
              <span>Режим пересчета</span>
              <select v-model="gamificationSettings.recalculationMode">
                <option value="manual">Ручной</option>
                <option value="scheduled">По расписанию</option>
                <option value="automatic">Автоматический</option>
              </select>
            </label>
            <label class="toggle-field">
              <input v-model="gamificationSettings.enabled" type="checkbox" />
              <span>Включить геймификацию</span>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="gamificationSettings.leaderboardEnabled" type="checkbox" />
              <span>Показывать лидерборд волонтеров</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'analytics'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <BarChart3 :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Аналитика и отчеты</h2>
            <p>KPI клиента, период активности волонтера, плановые отчеты и аудит выгрузок.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Аналитика и отчеты')">
            <label>
              <span>Активный волонтер за, дней</span>
              <input v-model.number="analyticsSettings.activeVolunteerDays" type="number" min="1" />
            </label>
            <label>
              <span>Целевые часы за период</span>
              <input v-model.number="analyticsSettings.targetHours" type="number" min="0" />
            </label>
            <label>
              <span>Целевые мероприятия за период</span>
              <input v-model.number="analyticsSettings.targetEvents" type="number" min="0" />
            </label>
            <label>
              <span>Периодичность отчета</span>
              <select v-model="analyticsSettings.reportFrequency">
                <option value="weekly">Еженедельно</option>
                <option value="monthly">Ежемесячно</option>
                <option value="quarterly">Ежеквартально</option>
              </select>
            </label>
            <label class="toggle-field settings-form-wide">
              <input v-model="analyticsSettings.exportAuditEnabled" type="checkbox" />
              <span>Писать экспорт отчетов в аудит</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
      </div>
    </template>

    <template v-else-if="selectedSection.id === 'files'">
      <div class="settings-module-shell">
        <section class="detail-panel settings-module-panel">
          <span class="settings-card-icon">
            <FileArchive :size="26" />
          </span>
          <div>
            <p class="eyebrow">Раздел настроек</p>
            <h2>Файлы</h2>
            <p>Лимиты загрузки, разрешенные типы файлов, очистка и сроки хранения вложений.</p>
          </div>
        </section>

        <section class="detail-panel">
          <form class="settings-form settings-form-grid" @submit.prevent="saveModuleSettings('Файлы')">
            <label>
              <span>Максимум изображения, МБ</span>
              <input v-model.number="fileSettings.maxImageMb" type="number" min="1" />
            </label>
            <label>
              <span>Максимум вложения, МБ</span>
              <input v-model.number="fileSettings.maxAttachmentMb" type="number" min="1" />
            </label>
            <label class="settings-form-wide">
              <span>Разрешенные типы</span>
              <input v-model="fileSettings.allowedTypes" type="text" />
            </label>
            <label>
              <span>Хранить файлы, дней</span>
              <input v-model.number="fileSettings.retentionDays" type="number" min="1" />
            </label>
            <label class="toggle-field">
              <input v-model="fileSettings.cleanupEnabled" type="checkbox" />
              <span>Включить очистку неиспользуемых файлов</span>
            </label>
            <div class="form-actions settings-form-wide">
              <button class="primary-action" type="submit">
                <Save :size="17" />
                <span>Сохранить настройки</span>
              </button>
            </div>
          </form>
        </section>
        <p v-if="moduleSaveMessage" class="form-success">{{ moduleSaveMessage }}</p>
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
