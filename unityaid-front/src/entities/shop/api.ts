import { apiRequest } from '../../shared/api'
import { authState } from '../auth/store'
import type { CreateProductPayload, Product, ShopOrder, Wallet } from './types'

const token = () => authState.token

export const fetchWallet = () =>
  apiRequest<{ item: Wallet }>('/shop/wallet', { token: token() })

export const transferCoins = (payload: { recipientId: string; amount: number; comment?: string }) =>
  apiRequest<{ item: Wallet }>('/shop/wallet/transfers', { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const fetchProducts = () =>
  apiRequest<{ items: Product[] }>('/shop/products', { token: token() })

export const createProduct = (payload: CreateProductPayload) =>
  apiRequest<{ item: Product }>('/shop/products', { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const createShopOrder = (payload: { items: { productId: string; quantity: number }[]; comment?: string }) =>
  apiRequest<{ item: ShopOrder }>('/shop/orders', { method: 'POST', token: token(), body: JSON.stringify(payload) })

export const fetchShopOrders = (scope: 'own' | 'all' = 'own') =>
  apiRequest<{ items: ShopOrder[] }>(`/shop/orders${scope === 'all' ? '?scope=all' : ''}`, { token: token() })

export const updateShopOrderStatus = (id: string, status: string) =>
  apiRequest<{ item: ShopOrder }>(`/shop/orders/${id}`, { method: 'PATCH', token: token(), body: JSON.stringify({ status }) })
