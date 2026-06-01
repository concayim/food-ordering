<script setup>
import { computed, inject, onMounted, ref } from 'vue'
import { api } from '../api'

const toast = inject('toast')
const purchases = ref([])
const daily = ref([])
const ingredients = ref([])
const showModal = ref(false)
const today = new Date().toISOString().slice(0, 10)
const form = ref({
  ingredientId: '',
  quantity: 1,
  unitPrice: 0,
  totalCost: 0,
  purchaseDate: today,
  note: '',
})

const totalSpend = computed(() =>
  daily.value.reduce((sum, item) => sum + Number(item.totalCost || 0), 0),
)
const todaySpend = computed(() =>
  daily.value.find((item) => item.date === today)?.totalCost || 0,
)

async function load() {
  try {
    const [ings, list, stats] = await Promise.all([
      api.listIngredients(),
      api.listPurchases(),
      api.dailySpend(),
    ])
    ingredients.value = ings || []
    purchases.value = list || []
    daily.value = stats || []
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)

function openNew() {
  form.value = {
    ingredientId: ingredients.value[0]?.id || '',
    quantity: 1,
    unitPrice: 0,
    totalCost: 0,
    purchaseDate: today,
    note: '',
  }
  showModal.value = true
}

function money(n) {
  return Number(n || 0).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function fmtDate(date) {
  return new Date(`${date}T00:00:00`).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    weekday: 'short',
  })
}

function syncTotal() {
  const qty = Number(form.value.quantity)
  const price = Number(form.value.unitPrice)
  if (qty > 0 && price > 0) {
    form.value.totalCost = Number((qty * price).toFixed(2))
  }
}

async function save() {
  if (!form.value.ingredientId) { toast('请选择原材料', 'error'); return }
  if (Number(form.value.quantity) <= 0) { toast('采购数量必须大于 0', 'error'); return }
  if (Number(form.value.totalCost) <= 0 && Number(form.value.unitPrice) <= 0) {
    toast('请填写采购金额或单价', 'error')
    return
  }
  try {
    await api.createPurchase({
      ingredientId: Number(form.value.ingredientId),
      quantity: Number(form.value.quantity),
      unitPrice: Number(form.value.unitPrice),
      totalCost: Number(form.value.totalCost),
      purchaseDate: form.value.purchaseDate,
      note: form.value.note,
    })
    toast('采购已入账')
    showModal.value = false
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}
</script>

<template>
  <div>
    <div class="head">
      <h2>财务统计</h2>
      <button class="btn-primary" @click="openNew">+ 记采购</button>
    </div>

    <div class="metrics">
      <div class="metric card">
        <span>今日原材料支出</span>
        <strong>¥{{ money(todaySpend) }}</strong>
      </div>
      <div class="metric card">
        <span>累计原材料支出</span>
        <strong>¥{{ money(totalSpend) }}</strong>
      </div>
      <div class="metric card">
        <span>采购记录</span>
        <strong>{{ purchases.length }}</strong>
      </div>
    </div>

    <div class="grid">
      <section class="card panel">
        <div class="panel-head">
          <h3>每日支出</h3>
          <span>共 {{ daily.length }} 天</span>
        </div>
        <div class="daily-list">
          <div v-for="row in daily" :key="row.date" class="daily-row">
            <div>
              <b>{{ fmtDate(row.date) }}</b>
              <span>{{ row.purchaseCount }} 笔采购</span>
            </div>
            <strong>¥{{ money(row.totalCost) }}</strong>
          </div>
          <div v-if="daily.length === 0" class="empty">暂无采购支出。</div>
        </div>
      </section>

      <section class="card panel">
        <div class="panel-head">
          <h3>采购流水</h3>
          <span>最近记录</span>
        </div>
        <div class="purchase-list">
          <div v-for="p in purchases" :key="p.id" class="purchase-row">
            <div>
              <b>{{ p.ingredient?.name || '原材料' }}</b>
              <span>{{ p.purchaseDate }} · {{ p.quantity }} {{ p.ingredient?.unit || '' }}</span>
              <em v-if="p.note">{{ p.note }}</em>
            </div>
            <strong>¥{{ money(p.totalCost) }}</strong>
          </div>
          <div v-if="purchases.length === 0" class="empty">暂无采购流水。</div>
        </div>
      </section>
    </div>

    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h2>记录原材料采购</h2>
        <div class="row">
          <div class="field" style="flex:2">
            <label>原材料</label>
            <select v-model="form.ingredientId">
              <option value="" disabled>请选择</option>
              <option v-for="it in ingredients" :key="it.id" :value="it.id">
                {{ it.name }}（{{ it.unit || '无单位' }}）
              </option>
            </select>
          </div>
          <div class="field" style="flex:1">
            <label>采购日期</label>
            <input v-model="form.purchaseDate" type="date" />
          </div>
        </div>
        <div class="row">
          <div class="field">
            <label>数量</label>
            <input v-model="form.quantity" type="number" min="0.01" step="0.01" @input="syncTotal" />
          </div>
          <div class="field">
            <label>单价</label>
            <input v-model="form.unitPrice" type="number" min="0" step="0.01" @input="syncTotal" />
          </div>
          <div class="field">
            <label>总金额</label>
            <input v-model="form.totalCost" type="number" min="0" step="0.01" />
          </div>
        </div>
        <div class="field">
          <label>备注</label>
          <input v-model="form.note" placeholder="供应商 / 批次 / 付款方式" />
        </div>
        <div class="row actions">
          <button class="btn-ghost" @click="showModal = false">取消</button>
          <button class="btn-primary" @click="save">入账</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h2 { margin: 0; font-size: 18px; }
h3 { margin: 0; font-size: 16px; }
.metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 16px; }
.metric { padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.metric span, .panel-head span, .daily-row span, .purchase-row span, .purchase-row em { color: var(--muted); font-size: 13px; font-style: normal; }
.metric strong { font-size: 24px; color: var(--primary); }
.grid { display: grid; grid-template-columns: 0.9fr 1.1fr; gap: 16px; align-items: start; }
.panel { padding: 16px 18px; }
.panel-head { display: flex; justify-content: space-between; align-items: center; padding-bottom: 12px; border-bottom: 1px solid var(--border); }
.daily-row, .purchase-row { display: flex; justify-content: space-between; gap: 12px; padding: 12px 0; border-bottom: 1px solid var(--border); }
.daily-row > div, .purchase-row > div { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.purchase-row b, .purchase-row span, .purchase-row em { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.daily-row strong, .purchase-row strong { color: var(--primary); white-space: nowrap; }
.empty { color: var(--muted); text-align: center; padding: 36px 0; }
.field { flex: 1; }
.actions { justify-content: flex-end; }
@media (max-width: 760px) {
  .metrics, .grid { grid-template-columns: 1fr; }
  .row { flex-direction: column; gap: 0; }
}
</style>
