import { computed, type Ref, ref, watch } from 'vue'

export function useClientPagination<T>(items: Ref<T[]>, perPage = 12) {
  const page = ref(1)
  const pageItems = computed(() => items.value.slice((page.value - 1) * perPage, page.value * perPage))
  watch(items, () => {
    page.value = 1
  })
  return { page, perPage, pageItems }
}
