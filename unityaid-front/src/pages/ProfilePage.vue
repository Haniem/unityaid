<script setup lang="ts">
import { authState } from '../entities/auth/store'
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Профиль</p>
        <h1>{{ authState.user?.lastName }} {{ authState.user?.firstName }}</h1>
      </div>
      <p>{{ authState.user?.email }}</p>
    </div>

    <div class="profile-panel">
      <img :src="authState.user?.avatarUrl ?? 'https://i.pravatar.cc/160?img=12'" alt="" />
      <dl>
        <div>
          <dt>Роль</dt>
          <dd>{{ authState.user?.primaryRole }}</dd>
        </div>
        <div>
          <dt>Язык</dt>
          <dd>{{ authState.user?.locale === 'en' ? 'English' : 'Русский' }}</dd>
        </div>
        <div>
          <dt>Организации</dt>
          <dd>{{ authState.user?.organizations.map((item) => item.organizationName).join(', ') }}</dd>
        </div>
      </dl>
    </div>
  </section>
</template>
