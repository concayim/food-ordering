<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { api } from '../api'
import { cart, addToCart, decFromCart, clearCart } from '../store'
import { menuDishes, loadMenu } from '../menu'

const toast = inject('toast')
const dishes = menuDishes
const remark = ref('')
const submitting = ref(false)
const detail = ref(null) // 当前查看详情的菜品

async function load() {
  try {
    await loadMenu()
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)

function add(dish) {
  addToCart(dish.id)
}
function dec(dish) {
  decFromCart(dish.id)
}
function openDetail(dish) {
  detail.value = dish
}

const cartList = computed(() =>
  Object.entries(cart).map(([id, qty]) => {
    const dish = dishes.value.find((d) => d.id === Number(id))
    return { dish, qty }
  }).filter((x) => x.dish)
)
const totalCount = computed(() => cartList.value.reduce((s, x) => s + x.qty, 0))

async function submit() {
  if (cartList.value.length === 0) return
  submitting.value = true
  try {
    const items = cartList.value.map((x) => ({ dishId: x.dish.id, quantity: x.qty }))
    const order = await api.createOrder({ remark: remark.value, items })
    toast(`下单成功！订单号 #${order.id}`)
    clearCart()
    remark.value = ''
    load()
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="order-layout">
    <div class="menu">
      <h2>菜单</h2>
      <div v-if="dishes.length === 0" class="empty">暂无上架菜品，请到「菜品管理」中上架。</div>
      <div class="dish-grid">
        <div v-for="d in dishes" :key="d.id" class="card dish-card" @click="openDetail(d)">
          <div class="thumb" :style="{ backgroundImage: `url(${d.imageUrl})` }">
            <span class="cat tag tag-orange" v-if="d.category">{{ d.category }}</span>
            <span class="kind tag" :class="d.kind === 'soup' ? 'tag-orange' : 'tag-gray'">{{ d.kind === 'soup' ? '汤品' : '菜品' }}</span>
          </div>
          <div class="info">
            <div class="name">{{ d.name }}</div>
            <div class="desc">{{ d.description }}</div>
            <div class="bottom">
              <span class="look">👀 查看烹饪方式</span>
              <div class="stepper" v-if="cart[d.id]" @click.stop>
                <button class="btn-ghost btn-sm" @click="dec(d)">−</button>
                <span>{{ cart[d.id] }}</span>
                <button class="btn-primary btn-sm" @click="add(d)">+</button>
              </div>
              <button v-else class="btn-primary btn-sm" @click.stop="add(d)">加入</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <aside class="cart card">
      <h2>🛒 购物车</h2>
      <div v-if="cartList.length === 0" class="empty">还没有选择菜品</div>
      <ul v-else class="cart-list">
        <li v-for="x in cartList" :key="x.dish.id">
          <span class="cn">{{ x.dish.name }}</span>
          <span class="cq">x{{ x.qty }}</span>
        </li>
      </ul>
      <textarea v-model="remark" rows="2" placeholder="备注（如：不要香菜）" />
      <div class="total">
        <span>共选</span>
        <strong>{{ totalCount }} 件</strong>
      </div>
      <button class="btn-primary submit" :disabled="cartList.length === 0 || submitting" @click="submit">
        {{ submitting ? '提交中...' : '下单' }}
      </button>
    </aside>

    <!-- 菜品详情 / 烹饪方式 -->
    <div v-if="detail" class="modal-mask" @click.self="detail = null">
      <div class="modal detail">
        <div class="d-thumb" :style="{ backgroundImage: `url(${detail.imageUrl})` }"></div>
        <div class="d-head">
          <h2>{{ detail.name }}</h2>
          <span class="tag" :class="detail.kind === 'soup' ? 'tag-orange' : 'tag-gray'">{{ detail.kind === 'soup' ? '汤品' : '菜品' }}</span>
          <span v-if="detail.category" class="tag tag-gray">{{ detail.category }}</span>
        </div>
        <p class="d-desc" v-if="detail.description">{{ detail.description }}</p>

        <div class="d-section">
          <h3>原材料</h3>
          <div v-if="detail.ingredients?.length" class="d-ings">
            <span v-for="di in detail.ingredients" :key="di.id" class="tag tag-gray">
              {{ di.ingredient?.name || '原料' }}{{ di.ingredient?.unit ? '(' + di.ingredient.unit + ')' : '' }} × {{ di.quantity }}
            </span>
          </div>
          <p v-else class="muted">未配置原材料</p>
        </div>

        <div class="d-section">
          <h3>🍳 烹饪方式</h3>
          <pre v-if="detail.cookingMethod" class="cooking">{{ detail.cookingMethod }}</pre>
          <p v-else class="muted">该菜品还没有填写烹饪方式。</p>
        </div>

        <div class="d-ops">
          <button class="btn-ghost" @click="detail = null">关闭</button>
          <button class="btn-primary" @click="add(detail); detail = null">加入购物车</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.order-layout { display: grid; grid-template-columns: 1fr 320px; gap: 24px; align-items: start; }
h2 { margin: 0 0 16px; font-size: 18px; }
.dish-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; }
.dish-card { overflow: hidden; cursor: pointer; transition: transform 0.12s, box-shadow 0.12s; }
.dish-card:hover { transform: translateY(-3px); box-shadow: 0 6px 18px rgba(0,0,0,0.1); }
.thumb { height: 130px; background-size: cover; background-position: center; background-color: #eee; position: relative; }
.cat { position: absolute; top: 8px; left: 8px; }
.kind { position: absolute; top: 8px; right: 8px; }
.info { padding: 12px 14px 14px; }
.name { font-weight: 600; font-size: 15px; }
.desc { color: var(--muted); font-size: 12px; margin: 4px 0 10px; height: 32px; overflow: hidden; }
.bottom { display: flex; align-items: center; justify-content: space-between; }
.look { color: var(--primary); font-size: 12px; }
.stepper { display: flex; align-items: center; gap: 10px; }
.cart { padding: 18px; position: sticky; top: 88px; }
.cart-list { list-style: none; padding: 0; margin: 0 0 12px; }
.cart-list li { display: flex; justify-content: space-between; padding: 6px 0; font-size: 14px; border-bottom: 1px dashed var(--border); }
.cart-list .cn { flex: 1; }
.cart-list .cq { color: var(--muted); }
.total { display: flex; justify-content: space-between; align-items: center; margin: 14px 0; font-size: 16px; }
.total strong { color: var(--primary); font-size: 20px; }
.submit { width: 100%; padding: 12px; font-size: 16px; }
.submit:disabled { opacity: 0.5; cursor: not-allowed; }
.empty { color: var(--muted); font-size: 14px; padding: 20px 0; text-align: center; }

.detail { width: 560px; }
.d-thumb { width: 100%; height: 200px; border-radius: 12px; background-size: cover; background-position: center; background-color: #eee; margin-bottom: 14px; }
.d-head { display: flex; align-items: center; gap: 10px; }
.d-head h2 { margin: 0; }
.d-desc { color: var(--muted); margin: 8px 0 0; }
.d-section { margin-top: 18px; }
.d-section h3 { font-size: 15px; margin: 0 0 10px; }
.d-ings { display: flex; flex-wrap: wrap; gap: 6px; }
.cooking { white-space: pre-wrap; word-break: break-word; background: #fafafa; border-radius: 10px; padding: 14px; font-family: inherit; font-size: 14px; line-height: 1.7; margin: 0; }
.muted { color: var(--muted); }
.d-ops { display: flex; justify-content: flex-end; gap: 10px; margin-top: 22px; }
@media (max-width: 760px) {
  .order-layout { grid-template-columns: 1fr; }
  .cart { position: static; }
}
</style>
