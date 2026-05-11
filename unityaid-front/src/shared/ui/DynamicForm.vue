<script setup lang="ts">
import { computed } from 'vue'
import { HelpCircle, ImagePlus } from 'lucide-vue-next'
import type { BackendForm, FormField, FormModel, FormValue } from '../../entities/forms/types'
import { uploadFormFile } from '../../entities/forms/api'
import { datetimeLocalValue } from '../forms'
import CustomSelect from './CustomSelect.vue'

const props = defineProps<{
  form: BackendForm
  modelValue: FormModel
}>()

const emit = defineEmits<{ 'update:modelValue': [value: FormModel] }>()
const fields = computed(() => props.form.fields)

function update(code: string, value: FormValue) {
  emit('update:modelValue', { ...props.modelValue, [code]: value })
}

function textInputType(field: FormField) {
  if (field.type === 'email' || field.type === 'tel' || field.type === 'url' || field.type === 'number') return field.type
  return 'text'
}

function valueOf(field: FormField) {
  return props.modelValue[field.code]
}

function isSelected(field: FormField, id: string) {
  const value = valueOf(field)
  return Array.isArray(value) && value.includes(id)
}

function toggleMulti(field: FormField, id: string, checked: boolean) {
  const current = Array.isArray(valueOf(field)) ? [...(valueOf(field) as string[])] : []
  update(field.code, checked ? [...new Set([...current, id])] : current.filter((item) => item !== id))
}

async function upload(field: FormField, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !field.uploadEndpoint) return
  const response = await uploadFormFile(field.uploadEndpoint, file)
  update(field.code, response.url)
  input.value = ''
}
</script>

<template>
  <div class="dynamic-form-grid">
    <label v-for="field in fields" :key="field.code" :class="['form-field', `field-${field.type || 'text'}`, { wide: field.type === 'textarea' || field.type === 'file' }]">
      <span class="field-label">
        <span class="field-title">
          {{ field.name || field.code }}<b v-if="field.require" class="required-star">*</b>
          <span v-if="field.help" class="help-anchor" tabindex="0">
            <HelpCircle :size="14" />
            <span class="help-tooltip">{{ field.help }}</span>
          </span>
        </span>
      </span>

      <textarea
        v-if="field.type === 'textarea'"
        :value="String(valueOf(field) ?? '')"
        :rows="field.rows || 4"
        :maxlength="field.maxLength || undefined"
        :placeholder="field.placeholder"
        :disabled="field.disabled"
        @input="update(field.code, ($event.target as HTMLTextAreaElement).value)"
      />

      <CustomSelect
        v-else-if="field.type === 'select' && !field.multi"
        :model-value="String(valueOf(field) ?? '')"
        :options="field.possibleValues || []"
        :disabled="field.disabled"
        @update:model-value="update(field.code, $event)"
      />

      <div v-else-if="field.type === 'select' && field.multi" class="multi-select-list">
        <label v-for="option in field.possibleValues || []" :key="option.id" class="multi-select-option">
          <input type="checkbox" :checked="isSelected(field, option.id)" :disabled="field.disabled" @change="toggleMulti(field, option.id, ($event.target as HTMLInputElement).checked)" />
          <span>{{ option.name }}</span>
        </label>
      </div>

      <div v-else-if="field.type === 'file'" class="file-field">
        <label class="file-picker">
          <ImagePlus :size="18" />
          <span>{{ valueOf(field) ? 'Заменить файл' : 'Выбрать файл' }}</span>
          <input type="file" :accept="field.accept" :disabled="field.disabled" @change="upload(field, $event)" />
        </label>
        <img v-if="valueOf(field)" class="file-preview" :src="String(valueOf(field))" alt="" />
      </div>

      <input
        v-else-if="field.type === 'datetime'"
        type="datetime-local"
        :value="datetimeLocalValue(valueOf(field))"
        :disabled="field.disabled"
        @input="update(field.code, ($event.target as HTMLInputElement).value)"
      />

      <input
        v-else
        :type="textInputType(field)"
        :value="String(valueOf(field) ?? '')"
        :maxlength="field.maxLength || undefined"
        :min="field.min ?? undefined"
        :placeholder="field.placeholder"
        :required="field.require"
        :disabled="field.disabled"
        @input="update(field.code, ($event.target as HTMLInputElement).value)"
      />
    </label>
  </div>
</template>
