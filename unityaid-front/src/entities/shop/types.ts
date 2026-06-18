export type CoinTransaction = {
  id: string
  userId: string
  amount: number
  type: string
  description: string
  createdAt: string
}

export type Wallet = {
  userId: string
  balance: number
  transactions: CoinTransaction[]
}

export type Product = {
  id: string
  organizationId?: string
  organizationName?: string
  name: string
  description: string
  price: number
  stock: number
  imageUrl?: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export type CreateProductPayload = {
  organizationId?: string | null
  name: string
  description: string
  price: number
  stock: number
  imageUrl?: string | null
  isActive?: boolean
}

export type OrderItem = {
  id: string
  productId?: string | null
  productName: string
  price: number
  quantity: number
}

export type ShopOrder = {
  id: string
  userId: string
  userName: string
  organizationId?: string | null
  organizationName?: string | null
  status: 'pending' | 'processing' | 'completed' | 'cancelled' | string
  total: number
  comment: string
  items: OrderItem[]
  createdAt: string
  updatedAt: string
}
