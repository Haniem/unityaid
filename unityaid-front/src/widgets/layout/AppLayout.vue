<script setup lang="ts">
import TopBar from './TopBar.vue'
</script>

<template>
  <div class="app-layout">
    <div class="app-main">
      <TopBar />
      <RouterView />
    </div>
  </div>
</template>
