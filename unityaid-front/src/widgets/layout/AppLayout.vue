<script setup lang="ts">
import { ref } from 'vue'
import SidebarMenu from './SidebarMenu.vue'
import TopBar from './TopBar.vue'

const isSidebarCollapsed = ref(false)
</script>

<template>
  <div :class="['app-layout', { 'is-collapsed': isSidebarCollapsed }]">
    <SidebarMenu :collapsed="isSidebarCollapsed" @toggle="isSidebarCollapsed = !isSidebarCollapsed" />
    <div class="app-main">
      <TopBar />
      <RouterView />
    </div>
  </div>
</template>
