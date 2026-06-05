<script setup>
import { ref, onMounted, inject } from 'vue'
import { api, INFINITE } from '../api'

const toast = inject('toast')
const items = ref([])
const showModal = ref(false)
const editing = ref(null)
const form = ref({ name: '', unit: '', stock: 0, infinite: false })

async function load() {
  try {
    items.value = await api.listIngredients()
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)

function openNew() {
  editing.value = null
  form.value = { name: '', unit: '', stock: 0, infinite: false }
  showModal.value = true
}
function openEdit(it) {
  editing.value = it
  form.value = {
    name: it.name, unit: it.unit,
    stock: it.stock === INFINITE ? 0 : it.stock,
    infinite: it.stock === INFINITE,
  }
  showModal.value = true
}

async function save() {
  if (!form.value.name) { toast('请填写名称', 'error'); return }
  try {
    const payload = {
      name: form.value.name,
      unit: form.value.unit,
      stock: form.value.infinite ? INFINITE : Number(form.value.stock),
    }
    if (editing.value) {
      await api.updateIngredient(editing.value.id, payload)
      toast('已保存')
    } else {
      await api.createIngredient(payload)
      toast('已添加原材料')
    }
    showModal.value = false
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}

async function toggleInfinite(it) {
  try {
    if (it.stock === INFINITE) {
      await api.setStock(it.id, { stock: 0, infinite: false })
      toast('已改为有限库存')
    } else {
      await api.setStock(it.id, { infinite: true })
      toast('已设为无限库存')
    }
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}

async function remove(it) {
  if (!confirm(`删除原材料「${it.name}」？`)) return
  try {
    await api.deleteIngredient(it.id)
    toast('已删除')
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}
</script>

<template>
  <div>
    <div class="head">
      <h2>原材料 / 库存</h2>
      <button class="btn-primary" @click="openNew">+ 新增原材料</button>
    </div>

    <div class="card table">
      <div class="tr th">
        <span>名称</span>
        <span>单位</span>
        <span>库存</span>
        <span class="c-act">操作</span>
      </div>
      <div v-for="it in items" :key="it.id" class="tr">
        <span class="nm">{{ it.name }}</span>
        <span>{{ it.unit || '-' }}</span>
        <span>
          <span v-if="it.stock === INFINITE" class="tag tag-orange">∞ 无限</span>
          <span v-else :class="['tag', it.stock > 0 ? 'tag-green' : 'tag-gray']">{{ it.stock }}</span>
        </span>
        <span class="c-act">
          <button class="btn-ghost btn-sm" @click="toggleInfinite(it)">
            {{ it.stock === INFINITE ? '设为有限' : '设为无限' }}
          </button>
          <button class="btn-ghost btn-sm" @click="openEdit(it)">编辑</button>
          <button class="btn-danger btn-sm" @click="remove(it)">删除</button>
        </span>
      </div>
      <div v-if="items.length === 0" class="empty">还没有原材料。</div>
    </div>

    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h2>{{ editing ? '编辑原材料' : '新增原材料' }}</h2>
        <div class="row">
          <div class="field" style="flex:2">
            <label>名称</label>
            <input v-model="form.name" placeholder="如：土豆" />
          </div>
          <div class="field" style="flex:1">
            <label>单位</label>
            <input v-model="form.unit" placeholder="个 / 克 / 份" />
          </div>
        </div>
        <div class="field">
          <label class="switch">
            <input type="checkbox" v-model="form.infinite" />
            <span>无限库存（不限量）</span>
          </label>
        </div>
        <div class="field" v-if="!form.infinite">
          <label>库存数量</label>
          <input v-model="form.stock" type="number" min="0" />
        </div>
        <div class="row" style="justify-content:flex-end">
          <button class="btn-ghost" @click="showModal = false">取消</button>
          <button class="btn-primary" @click="save">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h2 { margin: 0; font-size: 18px; }
.table { padding: 8px 18px; }
.tr { display: grid; grid-template-columns: 2fr 1fr 1fr 2fr; align-items: center; gap: 10px; padding: 12px 0; border-bottom: 1px solid var(--border); }
.tr.th { color: var(--muted); font-size: 13px; border-bottom: 2px solid var(--border); }
.nm { font-weight: 600; }
.c-act { display: flex; gap: 6px; }
.empty { color: var(--muted); text-align: center; padding: 30px 0; }
.switch { display: flex; align-items: center; gap: 8px; width: auto; }
.switch input { width: auto; }
@media (max-width: 760px) {
  .tr { grid-template-columns: 1fr 1fr; }
  .c-act { grid-column: 1 / -1; }
}
</style>
