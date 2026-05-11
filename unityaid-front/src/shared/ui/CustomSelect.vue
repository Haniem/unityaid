<script setup lang="ts">
import { computed, ref } from 'vue'
import { Check, ChevronDown } from 'lucide-vue-next'
import type { PossibleValue } from '../../entities/forms/types'

const props = defineProps<{
  modelValue: string | null
  options: PossibleValue[]
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const isOpen = ref(false)
const selected = computed(() => props.options.find((item) => item.id === props.modelValue))

function choose(value: string) {
  emit('update:modelValue', value)
  isOpen.value = false
}
</script>

<template>
  <div class="custom-select" :class="{ open: isOpen, disabled }">
    <button class="custom-select-button" type="button" :disabled="disabled" @click="isOpen = !isOpen">
      <span>{{ selected?.name || placeholder || 'Выберите значение' }}</span>
      <ChevronDown :size="17" />
    </button>
    <div v-if="isOpen" class="custom-select-menu">
      <button
        v-for="option in options"
        :key="option.id"
        class="custom-select-option"
        type="button"
        :class="{ selected: option.id === modelValue }"
        @click="choose(option.id)"
      >
        <span>{{ option.name }}</span>
        <Check v-if="option.id === modelValue" :size="16" />
      </button>
    </div>
  </div>
</template>
