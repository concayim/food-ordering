<script setup>
import { ref, onMounted, onBeforeUnmount, inject } from 'vue'
import { api, INFINITE } from '../api'

const toast = inject('toast')
const dishes = ref([])
const ingredients = ref([])
const showModal = ref(false)
const editing = ref(null)
const aiLoading = ref(false)

const blank = () => ({
  name: '', description: '', category: '', kind: 'dish',
  cookingMethod: '', imageUrl: '', onShelf: false, ingredients: [],
})
const form = ref(blank())

// 从来源网站获取烹饪方法
const sourceUrl = ref('')
const urlLoading = ref(false)

// 内联新建原材料
const newIng = ref({ name: '', unit: '', infinite: true, stock: 0 })

async function load() {
  try {
    ;[dishes.value, ingredients.value] = await Promise.all([
      api.listDishes(),
      api.listIngredients(),
    ])
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)

function ingredientName(id) {
  const i = ingredients.value.find((x) => x.id === id)
  return i ? `${i.name}${i.unit ? '(' + i.unit + ')' : ''}` : '未知'
}

function openNew() {
  editing.value = null
  form.value = blank()
  newIng.value = { name: '', unit: '', infinite: true, stock: 0 }
  sourceUrl.value = ''
  showModal.value = true
}
function openEdit(d) {
  editing.value = d
  sourceUrl.value = ''
  form.value = {
    name: d.name, description: d.description,
    category: d.category, kind: d.kind || 'dish', cookingMethod: d.cookingMethod || '',
    imageUrl: d.imageUrl, onShelf: d.onShelf,
    ingredients: (d.ingredients || []).map((di) => ({
      ingredientId: di.ingredientId, quantity: di.quantity,
    })),
  }
  showModal.value = true
}

function addIngredientRow() {
  const first = ingredients.value[0]
  form.value.ingredients.push({ ingredientId: first ? first.id : 0, quantity: 1 })
}
function removeIngredientRow(i) {
  form.value.ingredients.splice(i, 1)
}

// 手动录入并关联一个新原材料
async function createAndAddIngredient() {
  if (!newIng.value.name) { toast('请填写原材料名称', 'error'); return }
  try {
    const created = await api.createIngredient({
      name: newIng.value.name,
      unit: newIng.value.unit,
      stock: newIng.value.infinite ? INFINITE : Number(newIng.value.stock),
    })
    ingredients.value.unshift(created)
    form.value.ingredients.push({ ingredientId: created.id, quantity: 1 })
    newIng.value = { name: '', unit: '', infinite: true, stock: 0 }
    toast(`已新建原材料「${created.name}」并关联`)
  } catch (e) {
    toast(e.message, 'error')
  }
}

// ---------------- 图片：拍照 / 上传 ----------------
const showCamera = ref(false)
const videoEl = ref(null)
const fileInput = ref(null)
let stream = null

async function openCamera() {
  showCamera.value = true
  try {
    stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    await nextFrame()
    if (videoEl.value) {
      videoEl.value.srcObject = stream
      await videoEl.value.play()
    }
  } catch (e) {
    showCamera.value = false
    toast('无法访问摄像头：' + e.message, 'error')
  }
}
function nextFrame() {
  return new Promise((r) => requestAnimationFrame(() => r()))
}
function stopCamera() {
  if (stream) {
    stream.getTracks().forEach((t) => t.stop())
    stream = null
  }
  showCamera.value = false
}
async function capture() {
  const v = videoEl.value
  if (!v) return
  const canvas = document.createElement('canvas')
  canvas.width = v.videoWidth || 640
  canvas.height = v.videoHeight || 480
  canvas.getContext('2d').drawImage(v, 0, 0, canvas.width, canvas.height)
  canvas.toBlob(async (blob) => {
    try {
      const { url } = await api.uploadImage(blob, `dish-${Date.now()}.png`)
      form.value.imageUrl = url
      toast('拍照已保存')
    } catch (e) {
      toast(e.message, 'error')
    } finally {
      stopCamera()
    }
  }, 'image/png')
}
async function onFilePick(e) {
  const file = e.target.files?.[0]
  if (!file) return
  try {
    const { url } = await api.uploadImage(file, file.name)
    form.value.imageUrl = url
    toast('图片已上传')
  } catch (err) {
    toast(err.message, 'error')
  } finally {
    e.target.value = ''
  }
}
onBeforeUnmount(stopCamera)

// ---------------- AI 推荐烹饪方法 ----------------
async function suggestCooking() {
  if (!form.value.name) { toast('请先填写名称', 'error'); return }
  aiLoading.value = true
  try {
    const names = form.value.ingredients
      .map((x) => ingredients.value.find((i) => i.id === Number(x.ingredientId))?.name)
      .filter(Boolean)
    const { cookingMethod } = await api.aiCookingMethod({
      name: form.value.name, kind: form.value.kind, ingredients: names,
    })
    form.value.cookingMethod = cookingMethod
    toast('AI 已生成烹饪方法')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    aiLoading.value = false
  }
}

async function fetchFromUrl() {
  if (!sourceUrl.value) { toast('请填写来源网址', 'error'); return }
  urlLoading.value = true
  try {
    const { cookingMethod } = await api.aiCookingMethodFromURL({
      name: form.value.name, url: sourceUrl.value,
    })
    form.value.cookingMethod = cookingMethod
    toast('已从来源网站获取烹饪方法')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    urlLoading.value = false
  }
}

async function save() {
  if (!form.value.name) { toast('请填写名称', 'error'); return }
  try {
    const payload = {
      ...form.value,
      ingredients: form.value.ingredients
        .filter((x) => x.ingredientId)
        .map((x) => ({ ingredientId: Number(x.ingredientId), quantity: Number(x.quantity) })),
    }
    if (editing.value) {
      await api.updateDish(editing.value.id, payload)
      toast('已保存')
    } else {
      await api.createDish(payload)
      toast('已创建')
    }
    showModal.value = false
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}

async function toggle(d) {
  try {
    await api.toggleShelf(d.id)
    toast(d.onShelf ? '已下架' : '已上架')
    load()
  } catch (e) {
    toast(e.message, 'error')
  }
}
async function remove(d) {
  if (!confirm(`确定删除「${d.name}」？`)) return
  try {
    await api.deleteDish(d.id)
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
      <h2>菜品 / 汤品管理</h2>
      <button class="btn-primary" @click="openNew">+ 录入菜品</button>
    </div>

    <div class="card table">
      <div class="tr th">
        <span class="c-name">名称</span>
        <span class="c-cat">类型/分类</span>
        <span class="c-ing">原材料</span>
        <span class="c-state">状态</span>
        <span class="c-act">操作</span>
      </div>
      <div v-for="d in dishes" :key="d.id" class="tr">
        <span class="c-name">
          <img :src="d.imageUrl" class="mini" />
          <div>
            <div class="dn">{{ d.name }}</div>
            <div class="dd" v-if="d.cookingMethod" title="已配置烹饪方法">🍳 含烹饪方法</div>
            <div class="dd" v-else>{{ d.description }}</div>
          </div>
        </span>
        <span class="c-cat">
          <span :class="['tag', d.kind === 'soup' ? 'tag-orange' : 'tag-gray']">{{ d.kind === 'soup' ? '汤品' : '菜品' }}</span>
          <span class="sub">{{ d.category || '-' }}</span>
        </span>
        <span class="c-ing">
          <span v-for="di in d.ingredients" :key="di.id" class="tag tag-gray ing">
            {{ ingredientName(di.ingredientId) }}×{{ di.quantity }}
          </span>
          <span v-if="!d.ingredients?.length" class="muted">未配置</span>
        </span>
        <span class="c-state">
          <span :class="['tag', d.onShelf ? 'tag-green' : 'tag-gray']">{{ d.onShelf ? '已上架' : '已下架' }}</span>
        </span>
        <span class="c-act">
          <button class="btn-ghost btn-sm" @click="toggle(d)">{{ d.onShelf ? '下架' : '上架' }}</button>
          <button class="btn-ghost btn-sm" @click="openEdit(d)">编辑</button>
          <button class="btn-danger btn-sm" @click="remove(d)">删除</button>
        </span>
      </div>
      <div v-if="dishes.length === 0" class="empty">还没有菜品，点击右上角录入。</div>
    </div>

    <!-- 录入/编辑弹窗 -->
    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h2>{{ editing ? '编辑' : '录入菜品 / 汤品' }}</h2>

        <!-- 图片：拍照 / 上传 -->
        <div class="field">
          <label>图片（可拍照录入）</label>
          <div class="photo">
            <div class="preview" :style="{ backgroundImage: form.imageUrl ? `url(${form.imageUrl})` : 'none' }">
              <span v-if="!form.imageUrl" class="ph">无图片</span>
            </div>
            <div class="photo-ops">
              <button class="btn-primary btn-sm" @click="openCamera">📷 拍照</button>
              <button class="btn-ghost btn-sm" @click="fileInput.click()">🖼️ 上传图片</button>
              <button v-if="form.imageUrl" class="btn-danger btn-sm" @click="form.imageUrl = ''">清除</button>
              <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFilePick" />
              <input v-model="form.imageUrl" placeholder="或直接粘贴图片地址" style="margin-top:8px" />
            </div>
          </div>
        </div>

        <div class="row">
          <div class="field" style="flex:2">
            <label>名称</label>
            <input v-model="form.name" placeholder="如：紫菜蛋花汤" />
          </div>
          <div class="field" style="flex:1">
            <label>类型</label>
            <select v-model="form.kind">
              <option value="dish">菜品</option>
              <option value="soup">汤品</option>
            </select>
          </div>
        </div>
        <div class="field">
          <label>分类</label>
          <input v-model="form.category" placeholder="如：热菜 / 汤品" />
        </div>
        <div class="field">
          <label>描述</label>
          <input v-model="form.description" placeholder="一句话介绍" />
        </div>
        <div class="field">
          <label class="switch">
            <input type="checkbox" v-model="form.onShelf" />
            <span>{{ form.onShelf ? '上架' : '下架' }}</span>
          </label>
        </div>

        <!-- 烹饪方法：手动录入 / AI 推荐 / 从来源网站获取 -->
        <div class="field">
          <label class="lbl-row">
            <span>烹饪方法（上架前可填写）</span>
            <button class="btn-ghost btn-sm" :disabled="aiLoading" @click="suggestCooking">
              {{ aiLoading ? 'AI 生成中…' : '✨ AI 推荐' }}
            </button>
          </label>
          <textarea v-model="form.cookingMethod" rows="5" placeholder="不懂烹饪方式？可点「AI 推荐」生成，或在下方填写来源网址抓取…"></textarea>
          <div class="src-row">
            <input v-model="sourceUrl" placeholder="粘贴来源网站菜谱网址（如 https://...）" />
            <button class="btn-ghost btn-sm" :disabled="urlLoading" @click="fetchFromUrl">
              {{ urlLoading ? '获取中…' : '🌐 从网址获取' }}
            </button>
          </div>
        </div>

        <!-- 原材料：选择已有 / 新建 -->
        <div class="field">
          <label>原材料配置（手动关联，每份用量）</label>
          <div v-for="(rowItem, i) in form.ingredients" :key="i" class="ing-row">
            <select v-model="rowItem.ingredientId">
              <option v-for="ing in ingredients" :key="ing.id" :value="ing.id">
                {{ ing.name }}{{ ing.unit ? '(' + ing.unit + ')' : '' }}
              </option>
            </select>
            <input v-model="rowItem.quantity" type="number" min="0" step="0.1" style="width:90px" />
            <button class="btn-danger btn-sm" @click="removeIngredientRow(i)">移除</button>
          </div>
          <button class="btn-ghost btn-sm" style="margin-top:8px" @click="addIngredientRow" :disabled="ingredients.length === 0">
            + 关联已有原材料
          </button>

          <div class="new-ing">
            <span class="ni-title">没有想要的原材料？手动新建：</span>
            <div class="ni-row">
              <input v-model="newIng.name" placeholder="原材料名称" />
              <input v-model="newIng.unit" placeholder="单位" style="width:80px" />
              <label class="switch sm"><input type="checkbox" v-model="newIng.infinite" /><span>无限</span></label>
              <input v-if="!newIng.infinite" v-model="newIng.stock" type="number" min="0" placeholder="库存" style="width:80px" />
              <button class="btn-primary btn-sm" @click="createAndAddIngredient">新建并关联</button>
            </div>
          </div>
        </div>

        <div class="row" style="justify-content:flex-end; margin-top:8px">
          <button class="btn-ghost" @click="showModal = false">取消</button>
          <button class="btn-primary" @click="save">保存</button>
        </div>
      </div>
    </div>

    <!-- 摄像头拍照 -->
    <div v-if="showCamera" class="modal-mask" @click.self="stopCamera">
      <div class="modal camera">
        <h2>拍照</h2>
        <video ref="videoEl" autoplay playsinline class="cam-video"></video>
        <div class="row" style="justify-content:center; margin-top:12px">
          <button class="btn-ghost" @click="stopCamera">取消</button>
          <button class="btn-primary" @click="capture">📸 拍摄</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
h2 { margin: 0; font-size: 18px; }
.table { padding: 8px 18px; }
.tr { display: grid; grid-template-columns: 2.2fr 1fr 2fr 0.9fr 1.6fr; align-items: center; gap: 10px; padding: 12px 0; border-bottom: 1px solid var(--border); }
.tr.th { color: var(--muted); font-size: 13px; border-bottom: 2px solid var(--border); }
.c-name { display: flex; align-items: center; gap: 10px; }
.mini { width: 42px; height: 42px; border-radius: 8px; object-fit: cover; background: #eee; }
.dn { font-weight: 600; font-size: 14px; }
.dd { color: var(--muted); font-size: 12px; }
.c-cat .sub { display: block; color: var(--muted); font-size: 12px; margin-top: 4px; }
.ing { margin: 2px 4px 2px 0; }
.c-act { display: flex; gap: 6px; flex-wrap: wrap; }
.muted { color: var(--muted); font-size: 13px; }
.empty { color: var(--muted); text-align: center; padding: 30px 0; }
.switch { display: flex; align-items: center; gap: 8px; width: auto; }
.switch input { width: auto; }
.switch.sm { font-size: 13px; }
.ing-row { display: flex; gap: 8px; margin-bottom: 8px; align-items: center; }
.lbl-row { display: flex; justify-content: space-between; align-items: center; }
.src-row { display: flex; gap: 8px; margin-top: 8px; align-items: center; }
.src-row > button { white-space: nowrap; }

.photo { display: flex; gap: 14px; }
.preview { width: 120px; height: 120px; border-radius: 12px; background: #f0f1f3; background-size: cover; background-position: center; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.preview .ph { color: var(--muted); font-size: 13px; }
.photo-ops { display: flex; flex-direction: column; gap: 8px; flex: 1; align-items: flex-start; }
.photo-ops > button { width: auto; }

.new-ing { margin-top: 12px; padding: 12px; background: #fafafa; border-radius: 10px; }
.ni-title { font-size: 13px; color: var(--muted); }
.ni-row { display: flex; gap: 8px; align-items: center; margin-top: 8px; flex-wrap: wrap; }

.camera { width: 520px; }
.cam-video { width: 100%; border-radius: 12px; background: #000; }
@media (max-width: 760px) {
  .tr { grid-template-columns: 1fr 1fr; }
  .c-act { grid-column: 1 / -1; }
  .c-ing, .c-cat, .c-state { display: none; }
}
</style>
