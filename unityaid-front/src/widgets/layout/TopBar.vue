<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, HelpCircle, LogOut, MessageCircle, Plus, Search, UserRound } from 'lucide-vue-next'
import { authState, logoutRemote } from '../../entities/auth/store'

const router = useRouter()
const isMenuOpen = ref(false)
const popoverRef = ref<HTMLElement | null>(null)
const avatarRef = ref<HTMLElement | null>(null)

const fullName = computed(() => {
  const user = authState.user
  if (!user) return ''
  return [user.lastName, user.firstName, user.patronymic].filter(Boolean).join(' ')
})

function closeOutside(event: MouseEvent) {
  const target = event.target as Node
  if (!isMenuOpen.value) return
  if (popoverRef.value?.contains(target) || avatarRef.value?.contains(target)) return
  isMenuOpen.value = false
}

async function signOut() {
  await logoutRemote()
  await router.push('/login')
}

onMounted(() => document.addEventListener('mousedown', closeOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))
</script>

<template>
  <header class="topbar">
    <div class="topbar-search">
      <Search :size="18" />
      <input type="search" placeholder="Поиск" aria-label="Поиск" />
    </div>

    <div class="topbar-actions">
      <button class="icon-button accent" type="button" aria-label="Создать"><Plus :size="20" /></button>
      <button class="icon-button" type="button" aria-label="Сообщения"><MessageCircle :size="19" /></button>
      <button class="icon-button" type="button" aria-label="Уведомления">
        <Bell :size="19" />
        <span class="badge">57</span>
      </button>
      <button class="icon-button" type="button" aria-label="FAQ"><HelpCircle :size="19" /></button>
      <button ref="avatarRef" class="avatar-button" type="button" aria-label="Меню пользователя" @click="isMenuOpen = !isMenuOpen">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
      </button>
    </div>

    <div v-if="isMenuOpen" ref="popoverRef" class="user-popover">
      <div class="popover-user">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
        <strong>{{ fullName }}</strong>
      </div>

      <RouterLink class="popover-link" to="/profile" @click="isMenuOpen = false">
        <UserRound :size="17" />
        <span>Мой профиль</span>
      </RouterLink>

      <button class="popover-link danger" type="button" @click="signOut">
        <LogOut :size="17" />
        <span>Выйти из аккаунта</span>
      </button>
    </div>
  </header>
</template>
