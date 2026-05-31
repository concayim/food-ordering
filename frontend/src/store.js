import { reactive } from 'vue'

// 全局购物车，跨页签共享：{ dishId: quantity }
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
