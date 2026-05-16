<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Check, ChevronDown, Loader2, Search } from 'lucide-vue-next'
import type { PossibleValue } from '../../entities/forms/types'

const props = defineProps<{
  modelValue: string | null | undefined
  loadOptions: (params: { search: string; page: number; perPage: number }) => Promise<{ items: readonly PossibleValue[] }>
  disabled?: boolean
  placeholder?: string
  perPage?: number
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const root = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const query = ref('')
const page = ref(1)
const options = ref<PossibleValue[]>([])
const selectedOption = ref<PossibleValue | null>(null)
const isLoading = ref(false)
const hasMore = ref(false)

const selected = computed(() => options.value.find((item) => item.id === props.modelValue) || selectedOption.value)
const perPage = computed(() => props.perPage || 20)

async function load(reset = true) {
  if (reset) page.value = 1
  isLoading.value = true
  try {
    const response = await props.loadOptions({ search: query.value, page: page.value, perPage: perPage.value })
    options.value = reset ? [...response.items] : [...options.value, ...response.items]
    hasMore.value = response.items.length >= perPage.value
  } finally {
    isLoading.value = false
  }
}

function choose(option: PossibleValue) {
  selectedOption.value = option
  emit('update:modelValue', option.id)
  isOpen.value = false
}

async function loadMore() {
  page.value += 1
  await load(false)
}

function closeOutside(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

let timer: number | undefined
watch(query, () => {
  window.clearTimeout(timer)
  timer = window.setTimeout(() => load(true), 250)
})

watch(isOpen, (open) => {
  if (open && options.value.length === 0) load(true)
})

onMounted(() => document.addEventListener('mousedown', closeOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))
</script>

<template>
  <div ref="root" class="custom-select async-select" :class="{ open: isOpen, disabled }">
    <button class="custom-select-button" type="button" :disabled="disabled" @click="isOpen = !isOpen">
      <span>{{ selected?.name || placeholder || 'Выберите значение' }}</span>
      <ChevronDown :size="17" />
    </button>
    <div v-if="isOpen" class="custom-select-menu">
      <label class="custom-select-search">
        <Search :size="15" />
        <input v-model="query" type="search" placeholder="Поиск" @click.stop />
      </label>
      <button
        v-for="option in options"
        :key="option.id"
        class="custom-select-option"
        :class="{ selected: option.id === modelValue }"
        type="button"
        @click="choose(option)"
      >
        <span>{{ option.name }}</span>
        <Check v-if="option.id === modelValue" :size="16" />
      </button>
      <button v-if="hasMore" class="custom-select-more" type="button" :disabled="isLoading" @click="loadMore">
        <Loader2 v-if="isLoading" :size="15" />
        <span>Загрузить еще</span>
      </button>
      <p v-if="!isLoading && options.length === 0" class="custom-select-empty">Ничего не найдено</p>
    </div>
  </div>
</template>
