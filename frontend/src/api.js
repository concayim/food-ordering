import { getAuthToken } from './store'

const BASE = '/api'

async function request(url, options = {}) {
  const headers = { 'Content-Type': 'application/json' }
  const token = getAuthToken()
  if (token) headers['Authorization'] = 'Bearer ' + token

  const res = await fetch(BASE + url, { headers, ...options })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `请求失败 (${res.status})`)
  }
  return data
}

export const api = {
  // ========== 认证 ==========
  login: (payload) => request('/auth/login', { method: 'POST', body: JSON.stringify(payload) }),
  register: (payload) => request('/auth/register', { method: 'POST', body: JSON.stringify(payload) }),
  getMe: () => request('/auth/me'),

  // 菜品
  listDishes: (onShelfOnly = false) =>
    request(`/dishes${onShelfOnly ? '?onShelf=true' : ''}`),
  createDish: (dish) => request('/dishes', { method: 'POST', body: JSON.stringify(dish) }),
  updateDish: (id, dish) => request(`/dishes/${id}`, { method: 'PUT', body: JSON.stringify(dish) }),
  toggleShelf: (id) => request(`/dishes/${id}/shelf`, { method: 'PATCH' }),
  deleteDish: (id) => request(`/dishes/${id}`, { method: 'DELETE' }),

  // 原材料
  listIngredients: () => request('/ingredients'),
  createIngredient: (ing) => request('/ingredients', { method: 'POST', body: JSON.stringify(ing) }),
  updateIngredient: (id, ing) => request(`/ingredients/${id}`, { method: 'PUT', body: JSON.stringify(ing) }),
  setStock: (id, payload) => request(`/ingredients/${id}/stock`, { method: 'PATCH', body: JSON.stringify(payload) }),
  deleteIngredient: (id) => request(`/ingredients/${id}`, { method: 'DELETE' }),

  // 财务
  listPurchases: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request(`/finance/purchases${qs ? `?${qs}` : ''}`)
  },
  createPurchase: (purchase) => request('/finance/purchases', { method: 'POST', body: JSON.stringify(purchase) }),
  dailySpend: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request(`/finance/daily-spend${qs ? `?${qs}` : ''}`)
  },

  // 订单
  listOrders: () => request('/orders'),
  createOrder: (order) => request('/orders', { method: 'POST', body: JSON.stringify(order) }),
  updateOrderStatus: (id, status) => request(`/orders/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  getOrderReceipt: (id) => request(`/orders/${id}/receipt`),
  printOrder: (id) => request(`/orders/${id}/print`, { method: 'POST' }),

  // 打印机
  printerStatus: () => request('/printer/status'),
  printerTest: () => request('/printer/test', { method: 'POST' }),

  // 订单推送
  notifyStatus: () => request('/notify/status'),
  notifyTest: () => request('/notify/test', { method: 'POST' }),
  notifyOrder: (id) => request(`/orders/${id}/notify`, { method: 'POST' }),

  // 图片上传
  uploadImage: async (blob, filename = 'photo.png') => {
    const fd = new FormData()
    fd.append('file', blob, filename)
    const headers = {}
    const token = getAuthToken()
    if (token) headers['Authorization'] = 'Bearer ' + token
    const res = await fetch(BASE + '/upload', { method: 'POST', body: fd, headers })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error || '上传失败')
    return data
  },

  // AI
  aiCookingMethod: (payload) =>
    request('/ai/cooking-method', { method: 'POST', body: JSON.stringify(payload) }),
  aiCookingMethodFromURL: (payload) =>
    request('/ai/cooking-method-from-url', { method: 'POST', body: JSON.stringify(payload) }),
}

export const INFINITE = -1
