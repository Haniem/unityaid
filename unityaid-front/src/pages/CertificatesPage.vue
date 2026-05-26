<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Award, Download, FileCheck2, Plus, SearchCheck } from 'lucide-vue-next'
import { downloadCertificate, fetchCertificates, generateCertificate } from '../entities/certificates/api'
import type { Certificate, CertificateGeneratePayload, CertificateType } from '../entities/certificates/types'
import { authState } from '../entities/auth/store'
import { fetchUsers } from '../entities/users/api'
import type { User } from '../entities/users/types'
import { fetchOrganizations } from '../entities/organizations/api'
import type { Organization } from '../entities/organizations/types'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import AsyncSelect from '../shared/ui/AsyncSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

const certificates = ref<Certificate[]>([])
const users = ref<User[]>([])
const organizations = ref<Organization[]>([])
const isLoading = ref(false)
const isGenerating = ref(false)
const message = ref('')
const { page, perPage, pageItems } = useClientPagination(certificates, 10)
const certificateTypeOptions = [{ id: 'hours', name: 'Справка о часах' }, { id: 'participation', name: 'Сертификат участия' }]

const form = ref<CertificateGeneratePayload>({
  userId: '',
  organizationId: null,
  type: 'hours',
  title: '',
  description: ''
})

const canManage = computed(() => {
  const user = authState.user
  if (!user) return false
  const systemAdmin = user.systemRoles?.some((role) => role.code === 'system_admin') ?? false
  const organizationManager = user.organizations?.some((item) => ['super_admin', 'org_admin', 'coordinator'].includes(item.role)) ?? false
  return systemAdmin || user.primaryRole === 'super_admin' || organizationManager
})

const typeLabels: Record<CertificateType, string> = {
  hours: 'Справка о часах',
  participation: 'Сертификат участия'
}

async function load() {
  isLoading.value = true
  message.value = ''
  try {
    const [certificateResponse, usersResponse, organizationsResponse] = await Promise.all([
      fetchCertificates(),
      canManage.value ? fetchUsers() : Promise.resolve({ items: [] }),
      canManage.value ? fetchOrganizations() : Promise.resolve({ items: [] })
    ])
    certificates.value = certificateResponse.items
    users.value = usersResponse.items
    organizations.value = organizationsResponse.items
    if (!form.value.userId && users.value.length) {
      form.value.userId = users.value[0].id
    }
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось загрузить сертификаты'
  } finally {
    isLoading.value = false
  }
}

async function submit() {
  if (!form.value.userId) return
  isGenerating.value = true
  message.value = ''
  try {
    await generateCertificate({
      ...form.value,
      organizationId: form.value.organizationId || null,
      title: form.value.title?.trim() || undefined,
      description: form.value.description?.trim() || undefined
    })
    form.value.title = ''
    form.value.description = ''
    message.value = 'Сертификат сформирован'
    await load()
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось сформировать сертификат'
  } finally {
    isGenerating.value = false
  }
}

async function download(item: Certificate) {
  try {
    const blob = await downloadCertificate(item.id)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${item.verifyCode}.pdf`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'Не удалось скачать PDF'
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' })
}

async function loadUserOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchUsers({ search: params.search })
  return { items: response.items.slice((params.page - 1) * params.perPage, params.page * params.perPage).map((user) => ({ id: user.id, name: `${user.lastName} ${user.firstName} - ${user.email}` })) }
}

async function loadOrganizationOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchOrganizations({ search: params.search })
  const pageItems = response.items.slice((params.page - 1) * params.perPage, params.page * params.perPage).map((organization) => ({ id: organization.id, name: organization.name }))
  return { items: params.page === 1 && !params.search ? [{ id: '', name: 'Все организации' }, ...pageItems] : pageItems }
}

onMounted(load)
</script>

<template>
  <section class="page-section">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Сертификаты</p>
        <h1>Справки и сертификаты</h1>
        <p>Сформированные документы с проверочным кодом и PDF для подтверждения участия и часов.</p>
      </div>
      <RouterLink class="secondary-action" to="/certificates/verify">
        <SearchCheck :size="17" />
        <span>Проверить код</span>
      </RouterLink>
    </div>

    <p v-if="message" :class="message.includes('Не удалось') ? 'form-error' : 'form-success'">{{ message }}</p>

    <div class="certificate-layout">
      <form v-if="canManage" class="detail-panel settings-form" @submit.prevent="submit">
        <div class="section-heading">
          <div>
            <p class="eyebrow">Выпуск</p>
            <h2>Новый документ</h2>
          </div>
          <Plus :size="22" />
        </div>

        <label>
          Волонтер
          <AsyncSelect v-model="form.userId" :load-options="loadUserOptions" placeholder="Выберите волонтера" />
        </label>

        <label>
          Организация
          <AsyncSelect v-model="form.organizationId" :load-options="loadOrganizationOptions" placeholder="Все организации" />
        </label>

        <label>
          Тип
          <CustomSelect v-model="form.type" :options="certificateTypeOptions" />
        </label>

        <label>
          Заголовок
          <input v-model="form.title" type="text" placeholder="Оставьте пустым для стандартного" />
        </label>

        <label>
          Описание
          <textarea v-model="form.description" rows="4" placeholder="Дополнительный текст в документе"></textarea>
        </label>

        <button class="primary-action" type="submit" :disabled="isGenerating || !form.userId">
          <Award :size="17" />
          <span>{{ isGenerating ? 'Формирование...' : 'Сформировать' }}</span>
        </button>
      </form>

      <div class="certificate-list">
        <article v-for="item in pageItems" :key="item.id" class="detail-panel certificate-card">
          <div class="certificate-icon">
            <FileCheck2 :size="24" />
          </div>
          <div>
            <p class="eyebrow">{{ typeLabels[item.type] }}</p>
            <h2>{{ item.title }}</h2>
            <p>{{ item.description || 'Документ сформирован в системе «Пульс».' }}</p>
            <div class="certificate-meta">
              <span>{{ item.totalHours }} ч.</span>
              <span>{{ item.verifyCode }}</span>
              <span>{{ formatDate(item.issuedAt) }}</span>
            </div>
          </div>
          <button class="secondary-action" type="button" @click="download(item)">
            <Download :size="17" />
            <span>PDF</span>
          </button>
        </article>

        <PaginationBar v-model:page="page" :per-page="perPage" :total="certificates.length" />
        <p v-if="!isLoading && certificates.length === 0" class="empty-state">Сертификатов пока нет.</p>
      </div>
    </div>
  </section>
</template>
