<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ChevronDown, ChevronRight, Edit3, Lock, Plus, Settings2, Trash2 } from 'lucide-vue-next'
import { authState } from '../entities/auth/store'
import {
  createProfileField,
  createProfileGroup,
  deleteProfileField,
  deleteProfileGroup,
  fetchProfileSchema,
  updateProfileField,
  updateProfileGroup
} from '../entities/profileFields/api'
import type { ProfileField, ProfileFieldGroup, ProfileFieldPayload, ProfileGroupPayload } from '../entities/profileFields/types'

const organizationId = computed(() => authState.user?.organizationId || authState.user?.organizations[0]?.organizationId || '')
const groups = ref<ProfileFieldGroup[]>([])
const expanded = ref<string[]>([])
const message = ref('')
const isGroupModalOpen = ref(false)
const isFieldModalOpen = ref(false)
const editingGroupId = ref('')
const editingFieldId = ref('')
const fieldOptionsText = ref('')
const groupForm = reactive({ name: '', description: '', sortOrder: 100, isActive: true })
const fieldForm = reactive({
  groupId: '', name: '', type: 'text' as ProfileField['type'], required: false, editableByUser: true,
  isActive: true, sortOrder: 100, placeholder: '', help: ''
})

async function loadSchema() {
  if (!organizationId.value) {
    message.value = 'Для настройки полей требуется выбрать организацию.'
    return
  }
  try {
    groups.value = (await fetchProfileSchema(organizationId.value)).items
    if (!expanded.value.length) expanded.value = groups.value.map((group) => group.id)
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось загрузить поля профиля'
  }
}

function toggleGroup(group: ProfileFieldGroup) {
  expanded.value = expanded.value.includes(group.id)
    ? expanded.value.filter((id) => id !== group.id)
    : [...expanded.value, group.id]
}

function openGroupEditor(group?: ProfileFieldGroup) {
  editingGroupId.value = group?.id ?? ''
  groupForm.name = group?.name ?? ''
  groupForm.description = group?.description ?? ''
  groupForm.sortOrder = group?.sortOrder ?? (groups.value.length + 1) * 10
  groupForm.isActive = group?.isActive ?? true
  isGroupModalOpen.value = true
}

function openFieldEditor(group: ProfileFieldGroup, field?: ProfileField) {
  editingFieldId.value = field?.id ?? ''
  fieldForm.groupId = field?.groupId ?? group.id
  fieldForm.name = field?.name ?? ''
  fieldForm.type = field?.type ?? 'text'
  fieldForm.required = field?.required ?? false
  fieldForm.editableByUser = field?.editableByUser ?? true
  fieldForm.isActive = field?.isActive ?? true
  fieldForm.sortOrder = field?.sortOrder ?? (group.fields.length + 1) * 10
  fieldForm.placeholder = field?.placeholder ?? ''
  fieldForm.help = field?.help ?? ''
  fieldOptionsText.value = field?.options.map((option) => `${option.value} | ${option.label}`).join('\n') ?? ''
  isFieldModalOpen.value = true
}

async function saveGroup() {
  const payload: ProfileGroupPayload = { organizationId: organizationId.value, ...groupForm }
  try {
    if (editingGroupId.value) await updateProfileGroup(editingGroupId.value, payload)
    else await createProfileGroup(payload)
    isGroupModalOpen.value = false
    message.value = 'Группа полей сохранена'
    await loadSchema()
  } catch (error) { message.value = error instanceof Error ? error.message : 'Не удалось сохранить группу' }
}

function optionsFromText() {
  return fieldOptionsText.value.split('\n').map((line) => line.trim()).filter(Boolean).map((line) => {
    const [value, label] = line.split('|').map((item) => item.trim())
    return { value, label: label || value }
  })
}

async function saveField() {
  const payload: ProfileFieldPayload = { organizationId: organizationId.value, ...fieldForm, options: optionsFromText() }
  try {
    if (editingFieldId.value) await updateProfileField(editingFieldId.value, payload)
    else await createProfileField(payload)
    isFieldModalOpen.value = false
    message.value = 'Поле профиля сохранено'
    await loadSchema()
  } catch (error) { message.value = error instanceof Error ? error.message : 'Не удалось сохранить поле' }
}

async function removeGroup(group: ProfileFieldGroup) {
  if (group.isSystem || !confirm(`Удалить группу "${group.name}" и ее значения?`)) return
  try { await deleteProfileGroup(group.id, organizationId.value); await loadSchema() }
  catch (error) { message.value = error instanceof Error ? error.message : 'Не удалось удалить группу' }
}

async function removeField(field: ProfileField) {
  if (field.isSystem || !confirm(`Удалить поле "${field.name}"?`)) return
  try { await deleteProfileField(field.id, organizationId.value); await loadSchema() }
  catch (error) { message.value = error instanceof Error ? error.message : 'Не удалось удалить поле' }
}

onMounted(loadSchema)
</script>

<template>
  <section class="page-section profile-fields-settings">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Настройки профиля</p>
        <h1>Поля волонтера</h1>
        <p>Группируйте данные и добавляйте поля, которые будут заполняться в профиле участника.</p>
      </div>
      <div class="profile-fields-actions">
        <button class="secondary-action" type="button" @click="openGroupEditor()"><Plus :size="17" />Группа</button>
        <button class="primary-action" type="button" :disabled="!groups.length" @click="openFieldEditor(groups[0])"><Plus :size="17" />Поле</button>
      </div>
    </div>
    <p v-if="message" class="form-success">{{ message }}</p>
    <div class="profile-field-groups">
      <section v-for="group in groups" :key="group.id" class="profile-field-group">
        <header>
          <button class="profile-group-title" type="button" @click="toggleGroup(group)">
            <ChevronDown v-if="expanded.includes(group.id)" :size="18" /><ChevronRight v-else :size="18" />
            <strong>{{ group.name }}</strong>
            <Lock v-if="group.isSystem" :size="14" class="protected-mark" />
          </button>
          <div class="profile-row-actions">
            <button class="icon-button" type="button" aria-label="Добавить поле" @click="openFieldEditor(group)"><Plus :size="16" /></button>
            <button class="icon-button" type="button" aria-label="Редактировать группу" @click="openGroupEditor(group)"><Edit3 :size="16" /></button>
            <button v-if="!group.isSystem" class="icon-button danger-action" type="button" aria-label="Удалить группу" @click="removeGroup(group)"><Trash2 :size="16" /></button>
          </div>
        </header>
        <div v-if="expanded.includes(group.id)" class="profile-field-list">
          <div v-for="field in group.fields" :key="field.id" class="profile-field-row">
            <span>{{ field.name }} <small>({{ field.type }})</small></span>
            <span class="profile-field-state">{{ field.isActive ? 'Активно' : 'Отключено' }}</span>
            <Lock v-if="field.isSystem" :size="14" class="protected-mark" />
            <button class="icon-button" type="button" aria-label="Редактировать поле" @click="openFieldEditor(group, field)"><Edit3 :size="16" /></button>
            <button v-if="!field.isSystem" class="icon-button danger-action" type="button" aria-label="Удалить поле" @click="removeField(field)"><Trash2 :size="16" /></button>
          </div>
        </div>
      </section>
    </div>

    <div v-if="isGroupModalOpen" class="modal-backdrop" @click.self="isGroupModalOpen = false">
      <form class="modal-panel settings-form profile-field-editor" @submit.prevent="saveGroup">
        <h2>{{ editingGroupId ? 'Редактировать группу' : 'Новая группа' }}</h2>
        <label><span>Название</span><input v-model="groupForm.name" required /></label>
        <label><span>Описание</span><textarea v-model="groupForm.description" rows="3"></textarea></label>
        <label><span>Порядок</span><input v-model.number="groupForm.sortOrder" type="number" /></label>
        <label class="toggle-field"><input v-model="groupForm.isActive" type="checkbox" /><span>Активная группа</span></label>
        <div class="form-actions"><button class="secondary-action" type="button" @click="isGroupModalOpen = false">Отмена</button><button class="primary-action">Сохранить</button></div>
      </form>
    </div>
    <div v-if="isFieldModalOpen" class="modal-backdrop" @click.self="isFieldModalOpen = false">
      <form class="modal-panel settings-form profile-field-editor" @submit.prevent="saveField">
        <div class="section-heading"><h2>{{ editingFieldId ? 'Редактировать поле' : 'Новое поле' }}</h2><Settings2 :size="20" /></div>
        <label><span>Группа</span><select v-model="fieldForm.groupId"><option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option></select></label>
        <label><span>Название</span><input v-model="fieldForm.name" required /></label>
        <label><span>Тип данных</span><select v-model="fieldForm.type"><option value="text">Текст</option><option value="textarea">Многострочный текст</option><option value="number">Число</option><option value="date">Дата</option><option value="tel">Телефон</option><option value="email">Email</option><option value="url">Ссылка</option><option value="select">Выбор ответа</option><option value="multiselect">Несколько ответов</option><option value="checkbox">Флажок</option></select></label>
        <label><span>Подсказка</span><input v-model="fieldForm.placeholder" /></label>
        <label v-if="fieldForm.type === 'select' || fieldForm.type === 'multiselect'"><span>Варианты, по строке: значение | подпись</span><textarea v-model="fieldOptionsText" rows="4"></textarea></label>
        <label class="toggle-field"><input v-model="fieldForm.required" type="checkbox" /><span>Обязательное поле</span></label>
        <label class="toggle-field"><input v-model="fieldForm.editableByUser" type="checkbox" /><span>Волонтер может редактировать</span></label>
        <label class="toggle-field"><input v-model="fieldForm.isActive" type="checkbox" /><span>Активное поле</span></label>
        <div class="form-actions"><button class="secondary-action" type="button" @click="isFieldModalOpen = false">Отмена</button><button class="primary-action">Сохранить</button></div>
      </form>
    </div>
  </section>
</template>
