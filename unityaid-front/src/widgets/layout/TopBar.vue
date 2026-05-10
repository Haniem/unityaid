<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, HelpCircle, Languages, LogOut, MessageCircle, Plus, Search, UserRound } from 'lucide-vue-next'
import { authState, logoutRemote, setLocale } from '../../entities/auth/store'

const router = useRouter()
const isMenuOpen = ref(false)

const fullName = computed(() => {
  const user = authState.user
  if (!user) {
    return ''
  }
  return [user.lastName, user.firstName, user.patronymic].filter(Boolean).join(' ')
})

function changeLocale(locale: 'ru' | 'en') {
  setLocale(locale)
}

async function signOut() {
  await logoutRemote()
  await router.push('/login')
}
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
      <button class="avatar-button" type="button" aria-label="Меню пользователя" @click="isMenuOpen = !isMenuOpen">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
      </button>
    </div>

    <div v-if="isMenuOpen" class="user-popover">
      <div class="popover-user">
        <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
        <strong>{{ fullName }}</strong>
      </div>

      <RouterLink class="popover-link" to="/profile" @click="isMenuOpen = false">
        <UserRound :size="17" />
        <span>Мой профиль</span>
      </RouterLink>

      <div class="popover-row">
        <span><Languages :size="17" /> Язык</span>
        <div class="segmented">
          <button
            type="button"
            :class="{ active: authState.user?.locale === 'ru' }"
            @click="changeLocale('ru')"
          >
            Русский
          </button>
          <button
            type="button"
            :class="{ active: authState.user?.locale === 'en' }"
            @click="changeLocale('en')"
          >
            EN
          </button>
        </div>
      </div>

      <button class="popover-link danger" type="button" @click="signOut">
        <LogOut :size="17" />
        <span>Выйти из аккаунта</span>
      </button>
    </div>
  </header>
</template>
