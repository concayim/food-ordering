<script setup>
import { ref, provide, computed } from 'vue'
import { isLoggedIn, clearAuth } from './store'
import AuthView from './components/AuthView.vue'
import OrderView from './components/OrderView.vue'
import RouletteView from './components/RouletteView.vue'
import DishManage from './components/DishManage.vue'
import IngredientManage from './components/IngredientManage.vue'
import OrderHistory from './components/OrderHistory.vue'
import FinanceView from './components/FinanceView.vue'

const active = ref('order')
const loggedIn = computed(isLoggedIn)

const tabs = [
  { key: 'order', label: '点餐', icon: '🍽️' },
  { key: 'roulette', label: '随机轮盘', icon: '🎡' },
  { key: 'dishes', label: '菜品管理', icon: '📋' },
  { key: 'ingredients', label: '原材料 / 库存', icon: '📦' },
  { key: 'finance', label: '财务统计', icon: '💰' },
  { key: 'orders', label: '订单记录', icon: '🧾' },
]

const toast = ref(null)
let timer = null
function showToast(message, type = 'success') {
  toast.value = { message, type }
  clearTimeout(timer)
  timer = setTimeout(() => (toast.value = null), 2500)
}
provide('toast', showToast)
provide('setTab', (key) => { active.value = key })

function logout() {
  clearAuth()
  active.value = 'order'
  showToast('已退出登录')
}
</script>

<template>
  <div class="app">
    <template v-if="loggedIn">
      <header class="topbar">
        <div class="brand">🍜 小馆点餐</div>
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
        <button class="btn-ghost btn-sm logout-btn" @click="logout">退出</button>
      </header>

      <main class="content">
        <OrderView v-if="active === 'order'" />
        <RouletteView v-else-if="active === 'roulette'" />
        <DishManage v-else-if="active === 'dishes'" />
        <IngredientManage v-else-if="active === 'ingredients'" />
        <FinanceView v-else-if="active === 'finance'" />
        <OrderHistory v-else-if="active === 'orders'" />
      </main>
    </template>

    <AuthView v-else />

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
.tabs { display: flex; gap: 6px; flex-wrap: wrap; flex: 1; }
.tab {
  background: transparent;
  color: var(--muted);
  padding: 18px 16px;
  border-radius: 0;
  border-bottom: 3px solid transparent;
  font-size: 15px;
}
.tab.active { color: var(--primary); border-bottom-color: var(--primary); font-weight: 600; }
.logout-btn { flex-shrink: 0; }
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
