<script setup lang="ts">
import { onMounted } from 'vue'
import { loadTenantTheme } from '../../entities/tenantSettings/theme'
import TopBar from './TopBar.vue'

onMounted(() => {
  loadTenantTheme().catch(() => undefined)
})
</script>

<template>
  <div class="app-layout">
    <div class="app-main">
      <TopBar />
      <RouterView />
    </div>
  </div>
</template>
