import { reactive } from 'vue'

// ========== Auth State ==========
const auth = reactive({
  token: localStorage.getItem('auth_token') || null,
  user: JSON.parse(localStorage.getItem('auth_user') || 'null'),
})

export function setAuth(token, user) {
  auth.token = token
  auth.user = user
  localStorage.setItem('auth_token', token || '')
  localStorage.setItem('auth_user', JSON.stringify(user))
}

export function clearAuth() {
  auth.token = null
  auth.user = null
  localStorage.removeItem('auth_token')
  localStorage.removeItem('auth_user')
}

export function isLoggedIn() {
  return !!auth.token
}

export function getAuthToken() {
  return auth.token
}

export { auth }

// ========== Shopping Cart (跨页签共享) ==========
export const cart = reactive({})

export function addToCart(dishId, n = 1) {
  cart[dishId] = (cart[dishId] || 0) + n
}

export function decFromCart(dishId) {
  if (!cart[dishId]) return
  cart[dishId]--
  if (cart[dishId] <= 0) delete cart[dishId]
}

export function clearCart() {
  Object.keys(cart).forEach((k) => delete cart[k])
}
