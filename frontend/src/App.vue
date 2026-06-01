<script setup>
import { ref, provide } from 'vue'
import OrderView from './components/OrderView.vue'
import RouletteView from './components/RouletteView.vue'
import DishManage from './components/DishManage.vue'
import IngredientManage from './components/IngredientManage.vue'
import OrderHistory from './components/OrderHistory.vue'
import FinanceView from './components/FinanceView.vue'

const tabs = [
  { key: 'order', label: '点餐', icon: '🍽️' },
  { key: 'roulette', label: '随机轮盘', icon: '🎡' },
  { key: 'dishes', label: '菜品管理', icon: '📋' },
  { key: 'ingredients', label: '原材料 / 库存', icon: '📦' },
  { key: 'finance', label: '财务统计', icon: '💰' },
  { key: 'orders', label: '订单记录', icon: '🧾' },
]
const active = ref('order')

// 简单的全局提示
const toast = ref(null)
let timer = null
function showToast(message, type = 'success') {
  toast.value = { message, type }
  clearTimeout(timer)
  timer = setTimeout(() => (toast.value = null), 2500)
}
provide('toast', showToast)

// 供子组件切换页签（如轮盘抽中后跳转到点餐页）
provide('setTab', (key) => { active.value = key })
</script>

<template>
  <div class="app">
    <header class="topbar">
      <div class="brand">🍜 小馆点餐系统</div>
      <nav class="tabs">
        <button
          v-for="t in tabs"
          :key="t.key"
          :class="['tab', { active: active === t.key }]"
          @click="active = t.key"
        >
          <span>{{ t.icon }}</span> {{ t.label }}
        </button>
      </nav>
    </header>

    <main class="content">
      <OrderView v-if="active === 'order'" />
      <RouletteView v-else-if="active === 'roulette'" />
      <DishManage v-else-if="active === 'dishes'" />
      <IngredientManage v-else-if="active === 'ingredients'" />
      <FinanceView v-else-if="active === 'finance'" />
      <OrderHistory v-else-if="active === 'orders'" />
    </main>

    <transition name="fade">
      <div v-if="toast" :class="['toast', toast.type]">{{ toast.message }}</div>
    </transition>
  </div>
</template>

<style scoped>
.topbar {
  background: #fff;
  padding: 0 28px;
  display: flex;
  align-items: center;
  gap: 36px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.05);
  position: sticky;
  top: 0;
  z-index: 10;
}
.brand { font-size: 20px; font-weight: 700; padding: 16px 0; white-space: nowrap; }
.tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.tab {
  background: transparent;
  color: var(--muted);
  padding: 18px 16px;
  border-radius: 0;
  border-bottom: 3px solid transparent;
  font-size: 15px;
}
.tab.active { color: var(--primary); border-bottom-color: var(--primary); font-weight: 600; }
.content { max-width: 1100px; margin: 0 auto; padding: 24px; }

.toast {
  position: fixed;
  bottom: 32px; left: 50%; transform: translateX(-50%);
  background: #333; color: #fff;
  padding: 12px 22px; border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.25);
  z-index: 200;
}
.toast.error { background: var(--red); }
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
