<script setup>
import { ref, onMounted, inject } from 'vue'
import { api } from '../api'

const toast = inject('toast')
const orders = ref([])

const statusMap = {
  pending: { label: '待处理', cls: 'tag-orange' },
  paid: { label: '已支付', cls: 'tag-green' },
  done: { label: '已完成', cls: 'tag-gray' },
  cancelled: { label: '已取消', cls: 'tag-gray' },
}

async function load() {
  try {
    orders.value = await api.listOrders()
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)

async function setStatus(o, status) {
  try {
    await api.updateOrderStatus(o.id, status)
    toast('状态已更新')
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}

async function printOrder(o) {
  try {
    const res = await api.printOrder(o.id)
    toast(res.message || '已提交打印')
  } catch (e) {
    toast(e.message, 'error')
  }
}

async function pushOrder(o) {
  try {
    const res = await api.notifyOrder(o.id)
    toast(res.message || '已推送')
  } catch (e) {
    toast(e.message, 'error')
  }
}

function fmt(t) {
  return new Date(t).toLocaleString('zh-CN')
}
</script>

<template>
  <div>
    <h2>订单记录</h2>
    <div v-if="orders.length === 0" class="empty card">暂无订单。</div>
    <div v-for="o in orders" :key="o.id" class="card order">
      <div class="o-head">
        <div>
          <span class="oid">#{{ o.id }}</span>
          <span :class="['tag', statusMap[o.status]?.cls || 'tag-gray']">
            {{ statusMap[o.status]?.label || o.status }}
          </span>
        </div>
        <span class="time">{{ fmt(o.createdAt) }}</span>
      </div>
      <ul class="items">
        <li v-for="it in o.items" :key="it.id">
          <span>{{ it.dishName }} × {{ it.quantity }}</span>
        </li>
      </ul>
      <div v-if="o.remark" class="remark">备注：{{ o.remark }}</div>
      <div class="o-foot">
        <span class="total">共 {{ o.items.reduce((s, it) => s + it.quantity, 0) }} 件</span>
        <div class="ops">
          <button class="btn-ghost btn-sm" @click="pushOrder(o)">📣 推送</button>
          <button class="btn-ghost btn-sm" @click="printOrder(o)">🖨️ 打印</button>
          <button v-if="o.status === 'pending'" class="btn-primary btn-sm" @click="setStatus(o, 'paid')">标记已支付</button>
          <button v-if="o.status === 'paid'" class="btn-ghost btn-sm" @click="setStatus(o, 'done')">完成</button>
          <button v-if="o.status !== 'cancelled' && o.status !== 'done'" class="btn-danger btn-sm" @click="setStatus(o, 'cancelled')">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
h2 { margin: 0 0 16px; font-size: 18px; }
.order { padding: 16px 18px; margin-bottom: 14px; }
.o-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.oid { font-weight: 700; margin-right: 10px; }
.time { color: var(--muted); font-size: 13px; }
.items { list-style: none; padding: 0; margin: 0; }
.items li { display: flex; justify-content: space-between; padding: 5px 0; font-size: 14px; border-bottom: 1px dashed var(--border); }
.remark { color: var(--muted); font-size: 13px; margin-top: 8px; }
.o-foot { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; }
.total { font-weight: 700; color: var(--primary); font-size: 16px; }
.ops { display: flex; gap: 8px; }
.empty { color: var(--muted); text-align: center; padding: 40px 0; }
</style>
