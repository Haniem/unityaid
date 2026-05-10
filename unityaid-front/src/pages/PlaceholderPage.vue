<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => String(route.meta.title ?? 'Раздел в разработке'))
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">UnityAid</p>
        <h1>{{ title }}</h1>
      </div>
      <p>Этот раздел уже добавлен в навигацию и будет наполнен функционалом в следующих итерациях.</p>
    </div>
  </section>
</template>
