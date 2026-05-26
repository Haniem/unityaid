<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Award, BriefcaseBusiness, CalendarDays, Camera, Clock3, FileCheck2, KeyRound, ListTodo, Save, Star, UserRound } from 'lucide-vue-next'
import { authState, changePassword, fetchCurrentUser } from '../entities/auth/store'
import { fetchGamificationProfile } from '../entities/gamification/api'
import { fetchVolunteer, updateVolunteer } from '../entities/users/api'
import { fetchTasks } from '../entities/tasks/api'
import { fetchEvents } from '../entities/events/api'
import { fetchCertificates } from '../entities/certificates/api'
import { uploadFormFile } from '../entities/forms/api'
import { fetchProfileSchema, fetchProfileValues, saveProfileValues } from '../entities/profileFields/api'
import type { GamificationProfile } from '../entities/gamification/types'
import type { VolunteerProfile } from '../entities/users/types'
import type { TaskItem } from '../entities/tasks/types'
import type { EventItem } from '../entities/events/types'
import type { Certificate } from '../entities/certificates/types'
import type { ProfileFieldGroup, ProfileValue } from '../entities/profileFields/types'

type ProfileTab = 'personal' | 'tasks' | 'work' | 'events' | 'results'
const tabs: { id: ProfileTab; name: string; icon: object }[] = [
  { id: 'personal', name: 'Личная информация', icon: UserRound },
  { id: 'tasks', name: 'Задачи', icon: ListTodo },
  { id: 'work', name: 'Работа', icon: BriefcaseBusiness },
  { id: 'events', name: 'Мероприятия', icon: CalendarDays },
  { id: 'results', name: 'Достижения и документы', icon: Award }
]
const activeTab = ref<ProfileTab>('personal')
const currentPassword = ref('')
const newPassword = ref('')
const isSubmitting = ref(false)
const isPasswordModalOpen = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const gamification = ref<GamificationProfile | null>(null)
const volunteer = ref<VolunteerProfile | null>(null)
const tasks = ref<TaskItem[]>([])
const events = ref<EventItem[]>([])
const certificates = ref<Certificate[]>([])
const groups = ref<ProfileFieldGroup[]>([])
const customValues = ref<Record<string, ProfileValue>>({})
const isUploadingAvatar = ref(false)
const organizationId = computed(() => authState.user?.organizationId || authState.user?.organizations[0]?.organizationId || '')
const customGroups = computed(() => groups.value.filter((group) => group.isActive && group.fields.some((field) => !field.isSystem && field.isActive)))
const totalApprovedHours = computed(() => gamification.value?.totalHours ?? volunteer.value?.totalHours ?? 0)
const initials = computed(() => `${authState.user?.firstName?.[0] ?? ''}${authState.user?.lastName?.[0] ?? ''}`.toUpperCase())
const profileAvatar = computed(() => volunteer.value?.avatarUrl || authState.user?.avatarUrl || '')

function openPasswordModal() {
  currentPassword.value = ''
  newPassword.value = ''
  errorMessage.value = ''
  successMessage.value = ''
  isPasswordModalOpen.value = true
}

async function submitPasswordChange() {
  errorMessage.value = ''
  successMessage.value = ''
  isSubmitting.value = true
  try {
    await changePassword(currentPassword.value, newPassword.value)
    currentPassword.value = ''
    newPassword.value = ''
    successMessage.value = 'Пароль обновлен'
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : 'Не удалось сменить пароль' }
  finally { isSubmitting.value = false }
}

async function uploadAvatar(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !volunteer.value) return
  isUploadingAvatar.value = true
  errorMessage.value = ''
  try {
    const response = await uploadFormFile('/files/profile-avatars', file)
    volunteer.value.avatarUrl = response.url
    await saveBaseProfile()
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить аватар' }
  finally { isUploadingAvatar.value = false }
}

async function saveBaseProfile() {
  if (!authState.user || !volunteer.value) return
  isSubmitting.value = true
  try {
    const response = await updateVolunteer(authState.user.id, {
      firstName: volunteer.value.firstName,
      lastName: volunteer.value.lastName,
      patronymic: volunteer.value.patronymic,
      avatarUrl: volunteer.value.avatarUrl,
      city: volunteer.value.city,
      phone: volunteer.value.phone,
      bio: volunteer.value.bio,
      status: volunteer.value.status,
      interests: volunteer.value.interests,
      skillIds: volunteer.value.skills.map((skill) => skill.id)
    })
    volunteer.value = response.item
    await fetchCurrentUser()
    successMessage.value = 'Данные профиля сохранены'
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить профиль' }
  finally { isSubmitting.value = false }
}

async function saveCustomFields() {
  if (!authState.user || !organizationId.value) return
  isSubmitting.value = true
  try {
    const editable: Record<string, ProfileValue> = {}
    customGroups.value.forEach((group) => group.fields.filter((field) => !field.isSystem && field.editableByUser).forEach((field) => {
      editable[field.code] = customValues.value[field.code] ?? null
    }))
    customValues.value = (await saveProfileValues(organizationId.value, authState.user.id, editable)).values
    successMessage.value = 'Дополнительные поля сохранены'
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить поля' }
  finally { isSubmitting.value = false }
}

async function load() {
  if (!authState.user) return
  errorMessage.value = ''
  const requests = [
    fetchGamificationProfile().then((response) => { gamification.value = response.item }),
    fetchVolunteer(authState.user.id).then((response) => { volunteer.value = response.item }),
    fetchTasks({ assigneeId: authState.user.id }).then((response) => { tasks.value = response.items }),
    fetchEvents().then((response) => { events.value = response.items }),
    fetchCertificates().then((response) => { certificates.value = response.items.filter((item) => item.userId === authState.user?.id) })
  ]
  if (organizationId.value) {
    requests.push(fetchProfileSchema(organizationId.value).then((response) => { groups.value = response.items }))
    requests.push(fetchProfileValues(organizationId.value, authState.user.id).then((response) => { customValues.value = response.values }))
  }
  await Promise.allSettled(requests)
}

function fieldInputType(type: string) {
  if (['number', 'date', 'datetime', 'tel', 'email', 'url'].includes(type)) return type === 'datetime' ? 'datetime-local' : type
  return 'text'
}
function dateText(value?: string | null) { return value ? new Date(value).toLocaleString('ru-RU') : 'Без срока' }
onMounted(load)
</script>

<template>
  <section class="page-section profile-workspace">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Профиль волонтера</p>
        <h1>{{ authState.user?.lastName }} {{ authState.user?.firstName }}</h1>
      </div>
      <button class="secondary-action" type="button" @click="openPasswordModal"><KeyRound :size="18" /><span>Смена пароля</span></button>
    </div>

    <nav class="profile-tabs" aria-label="Разделы профиля">
      <button v-for="tab in tabs" :key="tab.id" :class="{ active: activeTab === tab.id }" type="button" @click="activeTab = tab.id">
        <component :is="tab.icon" :size="17" />{{ tab.name }}
      </button>
    </nav>
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

    <div class="profile-layout">
      <aside class="detail-panel profile-summary">
        <div class="avatar-editor">
          <img v-if="profileAvatar" :src="profileAvatar" alt="" />
          <span v-else class="avatar-fallback">{{ initials }}</span>
          <label class="avatar-upload" title="Сменить аватар">
            <Camera :size="16" />
            <input type="file" accept="image/png,image/jpeg,image/webp,image/gif" :disabled="isUploadingAvatar" @change="uploadAvatar" />
          </label>
        </div>
        <strong>{{ authState.user?.lastName }} {{ authState.user?.firstName }}</strong>
        <small>{{ authState.user?.email }}</small>
        <dl>
          <div><dt>Организация</dt><dd>{{ authState.user?.organizations[0]?.organizationName || 'Не назначена' }}</dd></div>
          <div><dt>Подтвержденные часы</dt><dd class="profile-hours"><Clock3 :size="15" />{{ totalApprovedHours.toFixed(1) }} ч.</dd></div>
          <div><dt>Баллы</dt><dd class="profile-hours"><Star :size="15" />{{ gamification?.points ?? 0 }}</dd></div>
        </dl>
      </aside>

      <main class="profile-content">
        <section v-if="activeTab === 'personal' && volunteer" class="detail-panel">
          <div class="section-heading"><div><p class="eyebrow">Персональные данные</p><h2>Личная информация</h2></div></div>
          <form class="settings-form settings-form-grid" @submit.prevent="saveBaseProfile">
            <label><span>Фамилия</span><input v-model="volunteer.lastName" required /></label>
            <label><span>Имя</span><input v-model="volunteer.firstName" required /></label>
            <label><span>Отчество</span><input v-model="volunteer.patronymic" /></label>
            <label><span>Телефон</span><input v-model="volunteer.phone" type="tel" /></label>
            <label class="settings-form-wide"><span>Электронная почта</span><input :value="volunteer.email" disabled /></label>
            <div class="form-actions settings-form-wide"><button class="primary-action" :disabled="isSubmitting"><Save :size="17" />Сохранить</button></div>
          </form>
        </section>

        <section v-if="activeTab === 'work' && volunteer" class="detail-panel">
          <div class="section-heading"><div><p class="eyebrow">Участие</p><h2>Работа и опыт</h2></div></div>
          <form class="settings-form settings-form-grid" @submit.prevent="saveBaseProfile">
            <label><span>Город</span><input v-model="volunteer.city" /></label>
            <label><span>Статус</span><select v-model="volunteer.status"><option value="new">Новый</option><option value="active">Активный</option><option value="unavailable">Временно недоступен</option><option value="archived">Архивный</option></select></label>
            <label class="settings-form-wide"><span>Опыт и описание</span><textarea v-model="volunteer.bio" rows="4"></textarea></label>
            <label class="settings-form-wide"><span>Интересы</span><textarea v-model="volunteer.interests" rows="3"></textarea></label>
            <div class="form-actions settings-form-wide"><button class="primary-action" :disabled="isSubmitting"><Save :size="17" />Сохранить</button></div>
          </form>
          <form v-for="group in customGroups" :key="group.id" class="profile-custom-group settings-form" @submit.prevent="saveCustomFields">
            <h3>{{ group.name }}</h3>
            <label v-for="field in group.fields.filter((item) => !item.isSystem && item.isActive)" :key="field.id">
              <span>{{ field.name }}<b v-if="field.required">*</b></span>
              <textarea v-if="field.type === 'textarea'" v-model="customValues[field.code] as string" rows="3" :disabled="!field.editableByUser"></textarea>
              <select v-else-if="field.type === 'select'" v-model="customValues[field.code] as string" :disabled="!field.editableByUser">
                <option value="">Не выбрано</option><option v-for="option in field.options" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
              <input v-else-if="field.type === 'checkbox'" v-model="customValues[field.code] as boolean" type="checkbox" :disabled="!field.editableByUser" />
              <input v-else v-model="customValues[field.code] as string" :type="fieldInputType(field.type)" :placeholder="field.placeholder" :disabled="!field.editableByUser" />
            </label>
            <div class="form-actions"><button class="primary-action"><Save :size="17" />Сохранить дополнительные данные</button></div>
          </form>
        </section>

        <section v-if="activeTab === 'tasks'" class="detail-panel">
          <div class="section-heading"><div><p class="eyebrow">Назначения</p><h2>Мои задачи</h2></div><RouterLink class="secondary-action" to="/tasks">Все задачи</RouterLink></div>
          <div class="profile-records">
            <RouterLink v-for="task in tasks.slice(0, 8)" :key="task.id" :to="`/tasks/${task.id}`"><strong>{{ task.title }}</strong><span>{{ task.status }} · {{ dateText(task.dueAt) }}</span></RouterLink>
            <p v-if="!tasks.length" class="empty-state">Назначенных задач нет.</p>
          </div>
        </section>
        <section v-if="activeTab === 'events'" class="detail-panel">
          <div class="section-heading"><div><p class="eyebrow">Календарь</p><h2>Мероприятия</h2></div><RouterLink class="secondary-action" to="/calendar">Календарь</RouterLink></div>
          <div class="profile-records">
            <RouterLink v-for="event in events.slice(0, 8)" :key="event.id" :to="`/calendar/${event.id}`"><strong>{{ event.title }}</strong><span>{{ dateText(event.startsAt) }} · {{ event.location || event.format }}</span></RouterLink>
          </div>
        </section>
        <section v-if="activeTab === 'results'" class="detail-panel">
          <div class="section-heading"><div><p class="eyebrow">Результаты</p><h2>Достижения и документы</h2></div></div>
          <div class="profile-achievement-list">
            <span v-for="achievement in gamification?.achievements ?? []" :key="achievement.id"><Award :size="15" />{{ achievement.name }}</span>
            <span v-if="!gamification?.achievements.length">Достижений пока нет</span>
          </div>
          <div class="profile-records document-records">
            <RouterLink v-for="certificate in certificates" :key="certificate.id" to="/certificates"><FileCheck2 :size="17" /><strong>{{ certificate.title }}</strong><span>{{ new Date(certificate.issuedAt).toLocaleDateString('ru-RU') }}</span></RouterLink>
            <p v-if="!certificates.length" class="empty-state">Выданных документов пока нет.</p>
          </div>
        </section>
      </main>
    </div>

    <div v-if="isPasswordModalOpen" class="modal-backdrop" @click.self="isPasswordModalOpen = false">
      <form class="modal-panel entity-form password-panel" @submit.prevent="submitPasswordChange">
        <div><p class="eyebrow">Безопасность</p><h2>Смена пароля</h2></div>
        <label><span>Текущий пароль</span><input v-model="currentPassword" type="password" autocomplete="current-password" required /></label>
        <label><span>Новый пароль</span><input v-model="newPassword" type="password" autocomplete="new-password" minlength="8" required /></label>
        <div class="form-actions"><button class="secondary-action" type="button" @click="isPasswordModalOpen = false">Отмена</button><button class="primary-action" type="submit" :disabled="isSubmitting">Обновить пароль</button></div>
      </form>
    </div>
  </section>
</template>
