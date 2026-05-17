<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Award, Clock3, FileSpreadsheet, HelpCircle, Link2, MailPlus, Pencil, Plus, Search, Trash2, Upload, UsersRound } from 'lucide-vue-next'
import { createSkill, deleteSkill, fetchSkills, fetchVolunteers, updateVolunteer } from '../entities/users/api'
import type { Skill, VolunteerProfile } from '../entities/users/types'
import { fetchEditForm } from '../entities/forms/api'
import type { BackendForm, FormModel } from '../entities/forms/types'
import { createInvitation, fetchInvitations } from '../entities/invitations/api'
import type { Invitation } from '../entities/invitations/types'
import { commitVolunteerImport, previewVolunteerImport } from '../entities/volunteerImports/api'
import type { ImportPreview } from '../entities/volunteerImports/types'
import DynamicForm from '../shared/ui/DynamicForm.vue'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { modelFromForm, nullable, stringValue } from '../shared/forms'
import { useClientPagination } from '../shared/pagination'

const volunteers = ref<VolunteerProfile[]>([])
const skills = ref<Skill[]>([])
const search = ref('')
const skillId = ref('')
const editing = ref<VolunteerProfile | null>(null)
const isModalOpen = ref(false)
const isHelpOpen = ref(false)
const loadError = ref('')
const errorMessage = ref('')
const skillName = ref('')
const skillError = ref('')
const formSchema = ref<BackendForm | null>(null)
const formModel = ref<FormModel>({})
const invitations = ref<Invitation[]>([])
const inviteEmail = ref('')
const inviteRole = ref('volunteer')
const inviteMessage = ref('')
const inviteError = ref('')
const importPreview = ref<ImportPreview | null>(null)
const importError = ref('')
const importMessage = ref('')
const isImporting = ref(false)
const { page, perPage, pageItems } = useClientPagination(volunteers, 12)

const totalHours = computed(() => volunteers.value.reduce((sum, item) => sum + item.totalHours, 0))
const averageLevel = computed(() => {
  if (!volunteers.value.length) return 0
  return volunteers.value.reduce((sum, item) => sum + item.level, 0) / volunteers.value.length
})

async function openEditModal(item: VolunteerProfile) {
  editing.value = item
  formSchema.value = await fetchEditForm('volunteers', item.userId)
  formModel.value = modelFromForm(formSchema.value)
  errorMessage.value = ''
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  editing.value = null
  errorMessage.value = ''
  formSchema.value = null
  formModel.value = {}
}

async function load() {
  loadError.value = ''
  try {
    const [volunteersResponse, skillsResponse, invitationsResponse] = await Promise.all([
      fetchVolunteers({ search: search.value, skillId: skillId.value }),
      fetchSkills(),
      fetchInvitations().catch(() => ({ items: [] }))
    ])
    volunteers.value = volunteersResponse.items
    skills.value = skillsResponse.items
    invitations.value = invitationsResponse.items
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : 'Не удалось загрузить волонтеров'
  }
}

async function submit() {
  if (!editing.value) return
  errorMessage.value = ''
  try {
    const response = await updateVolunteer(editing.value.userId, {
      firstName: stringValue(formModel.value.firstName),
      lastName: stringValue(formModel.value.lastName),
      patronymic: nullable(formModel.value.patronymic) as string | null,
      avatarUrl: nullable(formModel.value.avatarUrl) as string | null,
      city: nullable(formModel.value.city) as string | null,
      phone: nullable(formModel.value.phone) as string | null,
      bio: stringValue(formModel.value.bio),
      status: stringValue(formModel.value.status) || 'active',
      interests: stringValue(formModel.value.interests),
      skillIds: Array.isArray(formModel.value.skillIds) ? formModel.value.skillIds : []
    })
    const index = volunteers.value.findIndex((item) => item.userId === response.item.userId)
    if (index >= 0) volunteers.value[index] = response.item
    closeModal()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось сохранить профиль'
  }
}

async function addSkill() {
  const name = skillName.value.trim()
  if (!name) return
  skillError.value = ''
  try {
    const response = await createSkill({ name })
    if (!skills.value.some((item) => item.id === response.item.id)) {
      skills.value = [...skills.value, response.item].sort((a, b) => a.name.localeCompare(b.name))
    }
    skillName.value = ''
  } catch (error) {
    skillError.value = error instanceof Error ? error.message : 'Не удалось добавить навык'
  }
}

async function removeSkill(item: Skill) {
  if (!confirm(`Удалить навык "${item.name}"?`)) return
  await deleteSkill(item.id)
  skills.value = skills.value.filter((skill) => skill.id !== item.id)
  if (skillId.value === item.id) skillId.value = ''
}

async function inviteUser() {
  const email = inviteEmail.value.trim()
  if (!email) return
  inviteError.value = ''
  inviteMessage.value = ''
  try {
    const response = await createInvitation({ email, role: inviteRole.value })
    invitations.value = [response.item, ...invitations.value]
    inviteEmail.value = ''
    inviteMessage.value = 'Приглашение создано'
  } catch (error) {
    inviteError.value = error instanceof Error ? error.message : 'Не удалось создать приглашение'
  }
}

async function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  importError.value = ''
  importMessage.value = ''
  try {
    importPreview.value = await previewVolunteerImport(file)
  } catch (error) {
    importPreview.value = null
    importError.value = error instanceof Error ? error.message : 'Не удалось разобрать файл'
  } finally {
    input.value = ''
  }
}

async function commitImport() {
  if (!importPreview.value || importPreview.value.validCount === 0) return
  importError.value = ''
  importMessage.value = ''
  isImporting.value = true
  try {
    const response = await commitVolunteerImport(importPreview.value.items)
    importPreview.value = null
    importMessage.value = `Импортировано профилей: ${response.imported}`
    await load()
  } catch (error) {
    importError.value = error instanceof Error ? error.message : 'Не удалось импортировать волонтеров'
  } finally {
    isImporting.value = false
  }
}

function statusLabel(status: VolunteerProfile['status']) {
  const labels: Record<VolunteerProfile['status'], string> = {
    new: 'Новый',
    active: 'Активный',
    unavailable: 'Недоступен',
    archived: 'Архив'
  }
  return labels[status] || 'Активный'
}

function invitationStatusLabel(status: Invitation['status']) {
  const labels: Record<Invitation['status'], string> = {
    pending: 'Ожидает',
    accepted: 'Принято',
    revoked: 'Отозвано',
    expired: 'Истекло'
  }
  return labels[status] || status
}

let searchTimer: number | undefined
watch([search, skillId], () => {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(load, 250)
})

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Команда</p>
        <h1>Волонтеры</h1>
        <p>{{ volunteers.length }} профилей, {{ totalHours.toFixed(1) }} часов, средний уровень {{ averageLevel.toFixed(1) }}</p>
      </div>
      <button class="secondary-action" type="button" @click="isHelpOpen = true">
        <HelpCircle :size="18" />
        <span>Действия</span>
      </button>
    </div>

    <div class="metric-grid volunteer-metrics">
      <article class="metric-card">
        <UsersRound :size="22" />
        <span>Волонтеры</span>
        <strong>{{ volunteers.length }}</strong>
      </article>
      <article class="metric-card">
        <Clock3 :size="22" />
        <span>Подтвержденные часы</span>
        <strong>{{ totalHours.toFixed(1) }}</strong>
      </article>
      <article class="metric-card">
        <Award :size="22" />
        <span>Средний уровень</span>
        <strong>{{ averageLevel.toFixed(1) }}</strong>
      </article>
      <article class="metric-card">
        <Plus :size="22" />
        <span>Навыки</span>
        <strong>{{ skills.length }}</strong>
      </article>
    </div>

    <div class="filter-bar">
      <label class="search-field">
        <Search :size="18" />
        <input v-model="search" type="search" placeholder="Поиск по имени, email, городу или описанию" />
      </label>
      <label class="toggle-field">
        <span>Навык</span>
        <CustomSelect v-model="skillId" :options="[{ id: '', name: 'Все' }, ...skills.map((skill) => ({ id: skill.id, name: skill.name }))]" />
      </label>
    </div>

    <p v-if="loadError" class="form-error">{{ loadError }}</p>

    <div v-else class="volunteers-layout">
      <div v-if="volunteers.length" class="directory-grid">
        <article v-for="item in pageItems" :key="item.id" class="directory-card volunteer-card">
          <div class="directory-card-main">
            <div class="volunteer-card-head">
              <img :src="item.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
              <span class="status-pill">{{ item.level }} уровень</span>
              <span :class="['status-pill', `volunteer-status-${item.status}`]">{{ statusLabel(item.status) }}</span>
            </div>
            <h2>{{ item.lastName }} {{ item.firstName }}</h2>
            <p>{{ item.bio || 'Профиль пока не заполнен.' }}</p>
            <p v-if="item.interests" class="volunteer-interests">{{ item.interests }}</p>
            <div class="skill-list">
              <span v-for="skill in item.skills" :key="skill.id">{{ skill.name }}</span>
              <span v-if="!item.skills.length">Навыки не указаны</span>
            </div>
            <small>{{ item.email }} · {{ item.city || 'Город не указан' }}</small>
          </div>
          <div class="card-actions">
            <button class="icon-button" type="button" aria-label="Редактировать профиль" @click="openEditModal(item)">
              <Pencil :size="17" />
            </button>
          </div>
        </article>
      </div>
      <PaginationBar v-if="volunteers.length" v-model:page="page" :per-page="perPage" :total="volunteers.length" />
      <div v-else class="empty-state">Волонтеры не найдены.</div>

      <aside class="detail-panel skill-panel">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Справочник</p>
            <h2>Навыки</h2>
          </div>
        </div>
        <form class="inline-input" @submit.prevent="addSkill">
          <label class="sr-only" for="skill-name">Название нового навыка</label>
          <input id="skill-name" v-model="skillName" placeholder="Новый навык" />
          <button class="primary-action" type="submit">Добавить</button>
        </form>
        <p v-if="skillError" class="form-error">{{ skillError }}</p>
        <div class="skill-admin-list">
          <span v-for="skill in skills" :key="skill.id">
            {{ skill.name }}
            <button type="button" :aria-label="`Удалить ${skill.name}`" @click="removeSkill(skill)">
              <Trash2 :size="14" />
            </button>
          </span>
        </div>
        <div class="module-tool-block">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Доступ</p>
              <h2>Приглашения</h2>
            </div>
            <MailPlus :size="20" />
          </div>
          <form class="invite-row" @submit.prevent="inviteUser">
            <input v-model="inviteEmail" type="email" placeholder="email сотрудника" />
            <select v-model="inviteRole" aria-label="Роль приглашения">
              <option value="volunteer">Волонтер</option>
              <option value="coordinator">Координатор</option>
              <option value="org_admin">Администратор</option>
            </select>
            <button class="primary-action icon-button-wide" type="submit">
              <MailPlus :size="16" />
            </button>
          </form>
          <p v-if="inviteError" class="form-error">{{ inviteError }}</p>
          <p v-if="inviteMessage" class="form-success">{{ inviteMessage }}</p>
          <div class="invite-list compact-list">
            <article v-for="invite in invitations.slice(0, 4)" :key="invite.id" class="invite-card">
              <div>
                <span>{{ invite.email }}</span>
                <small>{{ invitationStatusLabel(invite.status) }} · {{ invite.role }}</small>
                <small class="copy-line"><Link2 :size="13" /> {{ invite.link }}</small>
              </div>
            </article>
            <p v-if="!invitations.length" class="muted-text">Приглашений пока нет.</p>
          </div>
        </div>

        <div class="module-tool-block">
          <div class="section-heading">
            <div>
              <p class="eyebrow">Загрузка</p>
              <h2>Импорт волонтеров</h2>
            </div>
            <FileSpreadsheet :size="20" />
          </div>
          <label class="file-drop">
            <Upload :size="18" />
            <span>CSV, TSV или XLSX</span>
            <input type="file" accept=".csv,.tsv,.xlsx" @change="handleImportFile" />
          </label>
          <p v-if="importError" class="form-error">{{ importError }}</p>
          <p v-if="importMessage" class="form-success">{{ importMessage }}</p>
          <div v-if="importPreview" class="import-preview">
            <div class="import-summary">
              <span>{{ importPreview.validCount }} готово</span>
              <span>{{ importPreview.errorCount }} с ошибками</span>
            </div>
            <div class="import-table">
              <div v-for="row in importPreview.items.slice(0, 5)" :key="row.row" :class="['import-row', { 'has-errors': row.errors.length }]">
                <span>#{{ row.row }}</span>
                <strong>{{ row.email || 'email не указан' }}</strong>
                <small>{{ row.errors.length ? row.errors.join(', ') : `${row.lastName} ${row.firstName}` }}</small>
              </div>
            </div>
            <button class="primary-action" type="button" :disabled="isImporting || importPreview.validCount === 0" @click="commitImport">
              Импортировать
            </button>
          </div>
        </div>
      </aside>
    </div>

    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
      <form class="modal-panel entity-form" @submit.prevent="submit">
        <h2>{{ formSchema?.meta.title || 'Профиль волонтера' }}</h2>
        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
        <DynamicForm v-if="formSchema" v-model="formModel" :form="formSchema" />
        <div class="form-actions">
          <button class="secondary-action" type="button" @click="closeModal">Отмена</button>
          <button class="primary-action" type="submit">Сохранить</button>
        </div>
      </form>
    </div>

    <div v-if="isHelpOpen" class="modal-backdrop" @click.self="isHelpOpen = false">
      <article class="modal-panel detail-panel">
        <h2>Что можно делать на странице</h2>
        <ul class="help-list">
          <li>Просматривать волонтеров, их навыки, часы, баллы и уровень.</li>
          <li>Искать профили по имени, email, городу и описанию.</li>
          <li>Фильтровать список по выбранному навыку.</li>
          <li>Редактировать профиль волонтера и набор навыков.</li>
          <li>Пополнять справочник навыков для дальнейших назначений.</li>
        </ul>
        <div class="form-actions">
          <button class="primary-action" type="button" @click="isHelpOpen = false">Понятно</button>
        </div>
      </article>
    </div>
  </section>
</template>
