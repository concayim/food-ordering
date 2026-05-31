<script setup>
import { ref } from 'vue'
import { apiGroups, methodColor } from '../api-docs'

const section = ref('api') // requirements | changelog | api
const expanded = ref(null)
const tryPath = ref('')
const tryBody = ref('')
const tryResult = ref('')
const trying = ref(false)

function toggle(i) {
  expanded.value = expanded.value === i ? null : i
}

async function sendRequest(ep) {
  trying.value = true
  tryResult.value = ''
  const path = tryPath.value || ep.path.replace(':id', '1')
  const url = path.startsWith('/api') ? path : '/api' + path.replace(/^\/api/, '')
  try {
    const opts = { headers: {} }
    if (ep.method !== 'GET' && ep.method !== 'DELETE') {
      opts.method = ep.method
      opts.headers['Content-Type'] = 'application/json'
      if (tryBody.value.trim()) opts.body = tryBody.value
    } else {
      opts.method = ep.method
    }
    const res = await fetch(url, opts)
    const text = await res.text()
    let pretty = text
    try { pretty = JSON.stringify(JSON.parse(text), null, 2) } catch (_) {}
    tryResult.value = `HTTP ${res.status}\n\n${pretty}`
  } catch (e) {
    tryResult.value = '请求失败: ' + e.message
  } finally {
    trying.value = false
  }
}

function prepTry(ep, gi, ei) {
  expanded.value = `${gi}-${ei}`
  tryPath.value = ep.path.replace(':id', '1')
  tryBody.value = ep.body && ep.body !== '同 POST' && !ep.body.startsWith('multipart')
    ? (ep.body.startsWith('{') ? ep.body : '')
    : ''
  tryResult.value = ''
}
</script>

<template>
  <div class="docs">
    <h2>📚 项目文档</h2>
    <div class="sec-tabs">
      <button :class="['st', { on: section === 'requirements' }]" @click="section = 'requirements'">需求文档</button>
      <button :class="['st', { on: section === 'changelog' }]" @click="section = 'changelog'">更新日志</button>
      <button :class="['st', { on: section === 'api' }]" @click="section = 'api'">接口文档</button>
    </div>

    <!-- 需求文档 -->
    <div v-if="section === 'requirements'" class="card doc-body">
      <h3>1. 项目概述</h3>
      <p>小馆点餐系统：Go + Gin + SQLite 后端，Vue 3 前端。面向小型餐饮的点餐、库存与后厨管理。</p>
      <h3>2. 核心功能</h3>
      <ul>
        <li><strong>点餐</strong>：浏览上架菜品/汤品，点击查看烹饪方式与原材料，购物车下单（按件数，无价格）</li>
        <li><strong>随机轮盘</strong>：从上架列表随机抽取，音效+高亮，抽中自动加购并跳转</li>
        <li><strong>菜品管理</strong>：拍照录入、烹饪方法（手动/AI/网址抓取）、原材料手动关联</li>
        <li><strong>原材料库存</strong>：有限或无限库存，下单自动扣减</li>
        <li><strong>订单记录</strong>：状态流转 pending → paid → done / cancelled</li>
      </ul>
      <h3>3. 非功能要求</h3>
      <ul>
        <li>API Key 仅通过环境变量配置，不写入代码库</li>
        <li>库存约定：<code>stock = -1</code> 表示无限</li>
        <li>AI 兼容 OpenAI 风格 <code>/v1/chat/completions</code></li>
      </ul>
      <h3>4. 环境变量</h3>
      <table class="tbl">
        <tr><th>变量</th><th>说明</th></tr>
        <tr><td><code>AI_API_KEY</code></td><td>大模型 Key（AI 功能必填）</td></tr>
        <tr><td><code>AI_BASE_URL</code></td><td>服务地址，如 http://127.0.0.1:8317/v1</td></tr>
        <tr><td><code>AI_MODEL</code></td><td>模型名，如 gpt-5.5</td></tr>
      </table>
      <p class="muted">完整版见仓库 <code>docs/REQUIREMENTS.md</code></p>
    </div>

    <!-- 更新日志 -->
    <div v-if="section === 'changelog'" class="card doc-body">
      <div class="log-item">
        <div class="ver">v0.3.0 <span class="date">2026-05-31</span></div>
        <ul>
          <li>移除菜品价格，改为件数统计</li>
          <li>点击菜品查看烹饪方式详情弹窗</li>
          <li>从网址获取烹饪方法；项目文档可视化页</li>
          <li>AI 配置改为环境变量，不再硬编码密钥</li>
        </ul>
      </div>
      <div class="log-item">
        <div class="ver">v0.2.0 <span class="date">2026-05-31</span></div>
        <ul>
          <li>随机轮盘（音效、高亮、自动跳转）</li>
          <li>汤品类型、拍照录入、AI 烹饪推荐</li>
          <li>原材料手动新建并关联</li>
        </ul>
      </div>
      <div class="log-item">
        <div class="ver">v0.1.0 <span class="date">2026-05-31</span></div>
        <ul>
          <li>初版：点餐、上下架、原材料库存、订单</li>
        </ul>
      </div>
      <p class="muted">完整版见仓库 <code>CHANGELOG.md</code></p>
    </div>

    <!-- 接口文档 -->
    <div v-if="section === 'api'" class="api-wrap">
      <p class="hint">基础地址 <code>/api</code>，错误响应统一为 <code>{"error":"..."}</code>。可展开接口并在线发送请求（GET 可直接试，POST 需填写 Body）。</p>

      <div v-for="(group, gi) in apiGroups" :key="group.name" class="card group">
        <h3>{{ group.name }}</h3>
        <div v-for="(ep, ei) in group.endpoints" :key="ep.path + ep.method" class="ep">
          <div class="ep-head" @click="toggle(`${gi}-${ei}`)">
            <span class="m" :style="{ background: methodColor[ep.method] }">{{ ep.method }}</span>
            <code class="p">{{ ep.path }}</code>
            <span class="d">{{ ep.desc }}</span>
            <span class="arrow">{{ expanded === `${gi}-${ei}` ? '▾' : '▸' }}</span>
          </div>
          <div v-if="expanded === `${gi}-${ei}`" class="ep-body">
            <p v-if="ep.note" class="note">{{ ep.note }}</p>
            <div v-if="ep.query" class="block">
              <label>Query</label>
              <div v-for="q in ep.query" :key="q.name" class="qrow"><code>{{ q.name }}</code> — {{ q.note }} <span v-if="q.example">例: {{ q.example }}</span></div>
            </div>
            <div v-if="ep.body" class="block">
              <label>Request Body</label>
              <pre>{{ ep.body }}</pre>
            </div>
            <div v-if="ep.response" class="block">
              <label>Response 示例</label>
              <pre>{{ ep.response }}</pre>
            </div>
            <div class="try">
              <label>在线调试</label>
              <input v-model="tryPath" placeholder="请求路径，如 /api/dishes?onShelf=true" />
              <textarea v-if="['POST','PUT','PATCH'].includes(ep.method)" v-model="tryBody" rows="4" placeholder="JSON Body（可选）"></textarea>
              <button class="btn-primary btn-sm" :disabled="trying" @click="sendRequest(ep)">{{ trying ? '请求中…' : '发送请求' }}</button>
              <pre v-if="tryResult" class="result">{{ tryResult }}</pre>
            </div>
            <button class="btn-ghost btn-sm" @click="prepTry(ep, gi, ei)">重置调试区</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
h2 { margin: 0 0 16px; font-size: 18px; }
.sec-tabs { display: flex; gap: 8px; margin-bottom: 20px; }
.st { background: #f0f1f3; color: var(--muted); padding: 8px 16px; border-radius: 8px; }
.st.on { background: var(--primary); color: #fff; }
.doc-body { padding: 22px 24px; }
.doc-body h3 { margin: 20px 0 10px; font-size: 15px; }
.doc-body h3:first-child { margin-top: 0; }
.doc-body ul { padding-left: 20px; line-height: 1.8; }
.doc-body p { line-height: 1.7; }
.tbl { width: 100%; border-collapse: collapse; font-size: 14px; margin: 10px 0; }
.tbl th, .tbl td { border: 1px solid var(--border); padding: 8px 12px; text-align: left; }
.tbl th { background: #fafafa; }
.muted { color: var(--muted); font-size: 13px; margin-top: 16px; }
code { background: #f0f1f3; padding: 2px 6px; border-radius: 4px; font-size: 13px; }

.log-item { margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px dashed var(--border); }
.ver { font-weight: 700; font-size: 16px; margin-bottom: 8px; }
.date { font-weight: 400; color: var(--muted); font-size: 13px; margin-left: 8px; }
.log-item ul { padding-left: 20px; line-height: 1.8; }

.hint { color: var(--muted); font-size: 14px; margin: 0 0 16px; }
.group { padding: 16px 18px; margin-bottom: 16px; }
.group h3 { margin: 0 0 12px; font-size: 15px; }
.ep { border: 1px solid var(--border); border-radius: 10px; margin-bottom: 8px; overflow: hidden; }
.ep-head { display: flex; align-items: center; gap: 10px; padding: 12px 14px; cursor: pointer; }
.ep-head:hover { background: #fafafa; }
.m { color: #fff; font-size: 11px; font-weight: 700; padding: 3px 8px; border-radius: 4px; flex-shrink: 0; }
.p { font-size: 13px; flex-shrink: 0; }
.d { color: var(--muted); font-size: 13px; flex: 1; }
.arrow { color: var(--muted); }
.ep-body { padding: 0 14px 14px; border-top: 1px solid var(--border); background: #fafafa; }
.block { margin-top: 12px; }
.block label { display: block; font-size: 12px; color: var(--muted); margin-bottom: 6px; font-weight: 600; }
.block pre, .result { background: #fff; border: 1px solid var(--border); border-radius: 8px; padding: 10px; font-size: 12px; overflow-x: auto; white-space: pre-wrap; word-break: break-all; margin: 0; }
.note { font-size: 13px; color: var(--primary); margin: 10px 0 0; }
.qrow { font-size: 13px; margin: 4px 0; }
.try { margin-top: 14px; }
.try input, .try textarea { margin-bottom: 8px; }
.try button { margin-top: 4px; }
.result { margin-top: 10px; max-height: 240px; overflow-y: auto; }
</style>
