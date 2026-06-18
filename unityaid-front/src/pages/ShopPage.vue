<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { CheckCircle2, Coins, Gift, Minus, PackageCheck, PackagePlus, Plus, Send, ShoppingCart } from 'lucide-vue-next'
import { createProduct, createShopOrder, fetchProducts, fetchShopOrders, fetchWallet, transferCoins, updateShopOrderStatus } from '../entities/shop/api'
import type { Product, ShopOrder, Wallet } from '../entities/shop/types'
import { authState } from '../entities/auth/store'
import { fetchVolunteers } from '../entities/users/api'
import { fetchOrganizations } from '../entities/organizations/api'
import type { PossibleValue } from '../entities/forms/types'
import AsyncSelect from '../shared/ui/AsyncSelect.vue'
import CustomSelect from '../shared/ui/CustomSelect.vue'
import PaginationBar from '../shared/ui/PaginationBar.vue'
import { useClientPagination } from '../shared/pagination'

type Tab = 'products' | 'orders' | 'manage' | 'catalog'

const products = ref<Product[]>([])
const orders = ref<ShopOrder[]>([])
const allOrders = ref<ShopOrder[]>([])
const wallet = ref<Wallet | null>(null)
const activeTab = ref<Tab>('products')
const errorMessage = ref('')
const successMessage = ref('')
const isLoading = ref(true)
const isSavingProduct = ref(false)
const organizationOptions = ref<PossibleValue[]>([])
const cart = reactive<Record<string, number>>({})
const checkoutComment = ref('')
const transfer = reactive({ recipientId: '', amount: '', comment: '' })
const productForm = reactive({ organizationId: '', name: '', description: '', price: '', stock: '', imageUrl: '' })
const { page: productPage, perPage: productPerPage, pageItems: productPageItems } = useClientPagination(products, 8)
const { page: orderPage, perPage: orderPerPage, pageItems: orderPageItems } = useClientPagination(orders, 8)
const { page: managePage, perPage: managePerPage, pageItems: managePageItems } = useClientPagination(allOrders, 8)

const isSystemAdmin = computed(() => Boolean(authState.user?.systemRoles?.some((role) => role.code === 'system_admin') || authState.user?.primaryRole === 'super_admin'))
const isManager = computed(() => isSystemAdmin.value || Boolean(authState.user?.organizations.some((membership) => ['org_admin', 'coordinator'].includes(membership.role))))
const cartItems = computed(() => products.value.filter((product) => (cart[product.id] ?? 0) > 0).map((product) => ({ product, quantity: cart[product.id] })))
const cartTotal = computed(() => cartItems.value.reduce((sum, item) => sum + item.product.price * item.quantity, 0))
const canCheckout = computed(() => cartItems.value.length > 0 && (wallet.value?.balance ?? 0) >= cartTotal.value)
const canSubmitProduct = computed(() => productForm.organizationId && productForm.name.trim().length > 0 && Number(productForm.price) > 0 && productForm.stock !== '' && Number(productForm.stock) >= 0)
const orderStatusOptions = [
  { id: 'pending', name: 'Новый' },
  { id: 'processing', name: 'В обработке' },
  { id: 'completed', name: 'Выдан' },
  { id: 'cancelled', name: 'Отменен' }
]

function statusLabel(status: string) {
  return orderStatusOptions.find((item) => item.id === status)?.name || status
}

function transactionLabel(type: string) {
  const labels: Record<string, string> = {
    achievement: 'Достижение',
    demo_bonus: 'Демо-бонус',
    transfer_in: 'Получено',
    transfer_out: 'Перевод',
    purchase: 'Покупка',
    refund: 'Возврат'
  }
  return labels[type] || type
}

function changeQuantity(product: Product, delta: number) {
  const next = Math.max(0, Math.min(product.stock, (cart[product.id] ?? 0) + delta))
  if (next === 0) delete cart[product.id]
  else cart[product.id] = next
}

async function load() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    if (isManager.value) await loadOrganizationOptions()
    const [walletResponse, productsResponse, ordersResponse, allOrdersResponse] = await Promise.all([
      fetchWallet(),
      fetchProducts(),
      fetchShopOrders('own'),
      isManager.value ? fetchShopOrders('all') : Promise.resolve({ items: [] as ShopOrder[] })
    ])
    wallet.value = walletResponse.item
    products.value = productsResponse.items
    orders.value = ordersResponse.items
    allOrders.value = allOrdersResponse.items
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось загрузить магазин'
  } finally {
    isLoading.value = false
  }
}

async function loadOrganizationOptions() {
  if (isSystemAdmin.value) {
    organizationOptions.value = (await fetchOrganizations()).items.map((organization) => ({ id: organization.id, name: organization.name }))
  } else {
    organizationOptions.value = (authState.user?.organizations ?? [])
      .filter((membership) => ['org_admin', 'coordinator'].includes(membership.role))
      .map((membership) => ({ id: membership.organizationId, name: membership.organizationName }))
  }
  if (!productForm.organizationId && organizationOptions.value.length > 0) {
    productForm.organizationId = organizationOptions.value[0].id
  }
}

async function checkout() {
  if (!canCheckout.value) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await createShopOrder({
      items: cartItems.value.map((item) => ({ productId: item.product.id, quantity: item.quantity })),
      comment: checkoutComment.value
    })
    Object.keys(cart).forEach((key) => delete cart[key])
    checkoutComment.value = ''
    successMessage.value = 'Заказ создан и ожидает обработки'
    await load()
    activeTab.value = 'orders'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось оплатить заказ'
  }
}

async function submitTransfer() {
  if (!transfer.recipientId || Number(transfer.amount) <= 0) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    wallet.value = (await transferCoins({ recipientId: transfer.recipientId, amount: Number(transfer.amount), comment: transfer.comment })).item
    transfer.recipientId = ''
    transfer.amount = ''
    transfer.comment = ''
    successMessage.value = 'Монеты переведены'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось перевести монеты'
  }
}

async function submitProduct() {
  if (!canSubmitProduct.value || isSavingProduct.value) return
  errorMessage.value = ''
  successMessage.value = ''
  isSavingProduct.value = true
  try {
    await createProduct({
      organizationId: productForm.organizationId,
      name: productForm.name.trim(),
      description: productForm.description.trim(),
      price: Number(productForm.price),
      stock: Number(productForm.stock),
      imageUrl: productForm.imageUrl.trim() || null,
      isActive: true
    })
    productForm.organizationId = organizationOptions.value[0]?.id ?? ''
    productForm.name = ''
    productForm.description = ''
    productForm.price = ''
    productForm.stock = ''
    productForm.imageUrl = ''
    successMessage.value = 'Товар добавлен в каталог'
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось добавить товар'
  } finally {
    isSavingProduct.value = false
  }
}

async function loadVolunteerOptions(params: { search: string; page: number; perPage: number }) {
  const response = await fetchVolunteers({ search: params.search })
  const ownId = authState.user?.id
  const items: PossibleValue[] = response.items
    .filter((item) => item.userId !== ownId)
    .slice((params.page - 1) * params.perPage, params.page * params.perPage)
    .map((item) => ({ id: item.userId, name: `${item.lastName} ${item.firstName} · ${item.email}` }))
  return { items }
}

async function updateOrderStatus(order: ShopOrder, status: string) {
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const response = await updateShopOrderStatus(order.id, status)
    const index = allOrders.value.findIndex((item) => item.id === order.id)
    if (index >= 0) allOrders.value[index] = response.item
    const ownIndex = orders.value.findIndex((item) => item.id === order.id)
    if (ownIndex >= 0) orders.value[ownIndex] = response.item
    successMessage.value = 'Статус заказа обновлен'
    await load()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Не удалось обновить статус заказа'
  }
}

onMounted(load)
</script>

<template>
  <section class="page-section shop-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Корпоративный магазин</p>
        <h1>Магазин наград</h1>
        <p>Товары оплачиваются монетами, которые начисляются за достижения и активность.</p>
      </div>
      <div class="wallet-badge"><Coins :size="19" />{{ wallet?.balance ?? 0 }} монет</div>
    </div>

    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="form-success">{{ successMessage }}</p>

    <nav class="profile-tabs shop-tabs" aria-label="Разделы магазина">
      <button :class="{ active: activeTab === 'products' }" type="button" @click="activeTab = 'products'"><Gift :size="17" />Товары</button>
      <button :class="{ active: activeTab === 'orders' }" type="button" @click="activeTab = 'orders'"><ShoppingCart :size="17" />Мои заказы</button>
      <button v-if="isManager" :class="{ active: activeTab === 'manage' }" type="button" @click="activeTab = 'manage'"><PackageCheck :size="17" />Обработка</button>
      <button v-if="isManager" :class="{ active: activeTab === 'catalog' }" type="button" @click="activeTab = 'catalog'"><PackagePlus :size="17" />Каталог</button>
    </nav>

    <div v-if="isLoading" class="empty-state">Загрузка магазина...</div>

    <div v-else-if="activeTab === 'products'" class="shop-layout">
      <div>
        <div class="shop-grid">
          <article v-for="product in productPageItems" :key="product.id" class="shop-product-card">
            <img v-if="product.imageUrl" :src="product.imageUrl" alt="" />
            <div v-else class="shop-product-placeholder"><Gift :size="28" /></div>
            <div>
              <small>{{ product.organizationName || 'Пульс' }}</small>
              <h2>{{ product.name }}</h2>
              <p>{{ product.description }}</p>
            </div>
            <footer>
              <strong>{{ product.price }} монет</strong>
              <span>{{ product.stock }} шт.</span>
              <div class="quantity-control">
                <button type="button" :disabled="!cart[product.id]" @click="changeQuantity(product, -1)"><Minus :size="15" /></button>
                <b>{{ cart[product.id] ?? 0 }}</b>
                <button type="button" :disabled="product.stock <= (cart[product.id] ?? 0)" @click="changeQuantity(product, 1)"><Plus :size="15" /></button>
              </div>
            </footer>
          </article>
        </div>
        <PaginationBar v-model:page="productPage" :per-page="productPerPage" :total="products.length" />
      </div>

      <aside class="detail-panel cart-panel">
        <div class="section-heading">
          <div><p class="eyebrow">Корзина</p><h2>Оплата монетами</h2></div>
          <ShoppingCart :size="20" />
        </div>
        <div class="cart-list">
          <p v-if="!cartItems.length" class="muted-text">Корзина пока пустая.</p>
          <div v-for="item in cartItems" :key="item.product.id">
            <span>{{ item.product.name }}</span>
            <strong>{{ item.quantity }} × {{ item.product.price }}</strong>
          </div>
        </div>
        <label class="form-field"><span class="field-label">Комментарий к заказу</span><textarea v-model="checkoutComment" rows="3"></textarea></label>
        <div class="cart-total"><span>Итого</span><strong>{{ cartTotal }} монет</strong></div>
        <button class="primary-action" type="button" :disabled="!canCheckout" @click="checkout"><CheckCircle2 :size="17" />Оплатить</button>

        <div class="shop-transfer">
          <div class="section-heading"><div><p class="eyebrow">Перевод</p><h2>Передать монеты</h2></div><Send :size="20" /></div>
          <form class="settings-form" @submit.prevent="submitTransfer">
            <AsyncSelect v-model="transfer.recipientId" :load-options="loadVolunteerOptions" placeholder="Выберите пользователя" />
            <input v-model="transfer.amount" type="number" min="1" placeholder="Количество монет" />
            <input v-model="transfer.comment" placeholder="Комментарий" />
            <button class="secondary-action" type="submit">Перевести</button>
          </form>
        </div>
      </aside>
    </div>

    <div v-else-if="activeTab === 'orders'" class="shop-orders">
      <article v-for="order in orderPageItems" :key="order.id" class="detail-panel shop-order-card">
        <header><span class="status-pill">{{ statusLabel(order.status) }}</span><strong>{{ order.total }} монет</strong></header>
        <p>{{ order.items.map((item) => `${item.productName} × ${item.quantity}`).join(', ') }}</p>
        <small>{{ new Date(order.createdAt).toLocaleString('ru-RU') }}</small>
      </article>
      <p v-if="!orders.length" class="empty-state">Заказов пока нет.</p>
      <PaginationBar v-model:page="orderPage" :per-page="orderPerPage" :total="orders.length" />
    </div>

    <div v-else-if="activeTab === 'manage'" class="shop-orders">
      <article v-for="order in managePageItems" :key="order.id" class="detail-panel shop-order-card manager-order-card">
        <header><div><span class="status-pill">{{ statusLabel(order.status) }}</span><strong>{{ order.userName }}</strong></div><strong>{{ order.total }} монет</strong></header>
        <p>{{ order.items.map((item) => `${item.productName} × ${item.quantity}`).join(', ') }}</p>
        <small>{{ order.organizationName || 'Без организации' }} · {{ new Date(order.createdAt).toLocaleString('ru-RU') }}</small>
        <CustomSelect :model-value="order.status" :options="orderStatusOptions" @update:model-value="(value) => updateOrderStatus(order, value)" />
      </article>
      <p v-if="!allOrders.length" class="empty-state">Заказов на обработку нет.</p>
      <PaginationBar v-model:page="managePage" :per-page="managePerPage" :total="allOrders.length" />
    </div>

    <div v-else-if="activeTab === 'catalog'" class="shop-layout">
      <form class="detail-panel settings-form settings-form-grid" @submit.prevent="submitProduct">
        <div class="section-heading settings-form-wide">
          <div><p class="eyebrow">Каталог</p><h2>Новый товар</h2></div>
          <PackagePlus :size="20" />
        </div>
        <label class="settings-form-wide"><span>Организация</span><CustomSelect v-model="productForm.organizationId" :options="organizationOptions" placeholder="Выберите организацию" /></label>
        <label><span>Название</span><input v-model="productForm.name" required /></label>
        <label><span>Цена в монетах</span><input v-model="productForm.price" type="number" min="1" required /></label>
        <label><span>Остаток</span><input v-model="productForm.stock" type="number" min="0" required /></label>
        <label class="settings-form-wide"><span>Ссылка на картинку</span><input v-model="productForm.imageUrl" type="url" placeholder="https://..." /></label>
        <label class="settings-form-wide"><span>Описание</span><textarea v-model="productForm.description" rows="4"></textarea></label>
        <div class="form-actions settings-form-wide">
          <button class="primary-action" type="submit" :disabled="!canSubmitProduct || isSavingProduct">
            <PackagePlus :size="17" />Добавить товар
          </button>
        </div>
      </form>

      <aside class="detail-panel cart-panel">
        <div class="section-heading">
          <div><p class="eyebrow">Витрина</p><h2>Текущие товары</h2></div>
          <Gift :size="20" />
        </div>
        <div class="cart-list">
          <div v-for="product in products.slice(0, 6)" :key="product.id">
            <span>{{ product.name }}</span>
            <strong>{{ product.price }} · {{ product.stock }} шт.</strong>
          </div>
          <p v-if="!products.length" class="muted-text">Каталог пока пустой.</p>
        </div>
      </aside>
    </div>
  </section>
</template>
