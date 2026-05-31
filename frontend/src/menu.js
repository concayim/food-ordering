import { ref } from 'vue'
import { api } from './api'

// 已上架菜品列表，供「点餐」与「随机轮盘」共享
export const menuDishes = ref([])

export async function loadMenu() {
  menuDishes.value = await api.listDishes(true)
  return menuDishes.value
}
