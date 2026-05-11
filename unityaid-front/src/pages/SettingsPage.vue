<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Save, Search, ShieldCheck, SlidersHorizontal } from 'lucide-vue-next'
import { fetchSystemRoles, fetchUsers, updateUserSystemRoles } from '../entities/users/api'
import type { SystemRole, User } from '../entities/users/types'

const users = ref<User[]>([])
const roles = ref<SystemRole[]>([])
const search = ref('')
const selectedUserId = ref('')
const selectedRoleIds = ref<string[]>([])
const loadError = ref('')
const saveMessage = ref('')
const isSaving = ref(false)

const filteredUsers = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return users.value
  return users.value.filter((user) => {
    const name = `${user.lastName} ${user.firstName} ${user.email}`.toLowerCase()
    return name.includes(query)
  })
})

const selectedUser = computed(() => users.value.find((user) => user.id === selectedUserId.value) || null)

async function load() {
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
        <h1>Системные роли</h1>
        <p>Управление правами пользователя на уровне всей системы отдельно от ролей внутри организаций.</p>
      </div>
    </div>

    <p v-if="loadError" class="form-error">{{ loadError }}</p>

    <div v-else class="settings-layout">
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
  </section>
</template>
