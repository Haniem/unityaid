<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{
  page: number
  perPage: number
  total: number
}>()

const emit = defineEmits<{ 'update:page': [value: number] }>()
const pageCount = computed(() => Math.max(1, Math.ceil(props.total / props.perPage)))
const from = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.perPage + 1))
const to = computed(() => Math.min(props.total, props.page * props.perPage))

function go(value: number) {
  emit('update:page', Math.min(pageCount.value, Math.max(1, value)))
}
</script>

<template>
  <nav v-if="total > perPage" class="pagination-bar" aria-label="Пагинация">
    <span>{{ from }}-{{ to }} из {{ total }}</span>
    <div>
      <button class="icon-button" type="button" :disabled="page <= 1" aria-label="Назад" @click="go(page - 1)">
        <ChevronLeft :size="17" />
      </button>
      <strong>{{ page }} / {{ pageCount }}</strong>
      <button class="icon-button" type="button" :disabled="page >= pageCount" aria-label="Вперед" @click="go(page + 1)">
        <ChevronRight :size="17" />
      </button>
    </div>
  </nav>
</template>
