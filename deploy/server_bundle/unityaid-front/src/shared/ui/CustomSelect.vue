<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Check, ChevronDown, Search } from 'lucide-vue-next'
import type { PossibleValue } from '../../entities/forms/types'

const props = defineProps<{
  modelValue: string | null | undefined
  options: readonly PossibleValue[]
  disabled?: boolean
  placeholder?: string
  searchable?: boolean
  pageSize?: number
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const isOpen = ref(false)
const query = ref('')
const page = ref(1)
const root = ref<HTMLElement | null>(null)

const selected = computed(() => props.options.find((item) => item.id === props.modelValue))
const isSearchable = computed(() => props.searchable !== false)
const pageSize = computed(() => props.pageSize || 20)
const filteredOptions = computed(() => {
  const value = query.value.trim().toLowerCase()
  if (!value) return props.options
  return props.options.filter((item) => item.name.toLowerCase().includes(value))
})
const visibleOptions = computed(() => filteredOptions.value.slice(0, page.value * pageSize.value))
const hasMore = computed(() => visibleOptions.value.length < filteredOptions.value.length)

function choose(value: string) {
  emit('update:modelValue', value)
  isOpen.value = false
}

function closeOutside(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

watch(query, () => {
  page.value = 1
})

watch(isOpen, (open) => {
  if (open) query.value = ''
})

onMounted(() => document.addEventListener('mousedown', closeOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))
</script>

<template>
  <div ref="root" class="custom-select" :class="{ open: isOpen, disabled }">
    <button class="custom-select-button" type="button" :disabled="disabled" @click="isOpen = !isOpen">
      <span>{{ selected?.name || placeholder || 'Выберите значение' }}</span>
      <ChevronDown :size="17" />
    </button>
    <div v-if="isOpen" class="custom-select-menu">
      <label v-if="isSearchable" class="custom-select-search">
        <Search :size="15" />
        <input v-model="query" type="search" placeholder="Поиск" @click.stop />
      </label>
      <button
        v-for="option in visibleOptions"
        :key="option.id"
        class="custom-select-option"
        type="button"
        :class="{ selected: option.id === modelValue }"
        @click="choose(option.id)"
      >
        <span>{{ option.name }}</span>
        <Check v-if="option.id === modelValue" :size="16" />
      </button>
      <button v-if="hasMore" class="custom-select-more" type="button" @click="page += 1">
        Показать еще {{ Math.min(pageSize, filteredOptions.length - visibleOptions.length) }}
      </button>
      <p v-if="filteredOptions.length === 0" class="custom-select-empty">Ничего не найдено</p>
    </div>
  </div>
</template>
