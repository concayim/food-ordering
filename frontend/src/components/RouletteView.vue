<script setup>
import { ref, computed, onMounted, onBeforeUnmount, inject } from 'vue'
import { menuDishes, loadMenu } from '../menu'
import { addToCart } from '../store'
import { sound, tick, fanfare } from '../sound'

const toast = inject('toast')
const setTab = inject('setTab')
const dishes = menuDishes
const rotation = ref(0)
const spinning = ref(false)
const result = ref(null)
const highlight = ref(-1) // 当前指针指向的扇区索引
const muted = ref(sound.muted)

const wheelEl = ref(null)
let rafId = null

async function load() {
  try {
    await loadMenu()
  } catch (e) {
    toast(e.message, 'error')
  }
}
onMounted(load)
onBeforeUnmount(() => cancelAnimationFrame(rafId))

const seg = computed(() => (dishes.value.length ? 360 / dishes.value.length : 0))

const colors = ['#ff6b35', '#ffb454', '#2ecc71', '#3498db', '#9b59b6', '#e74c3c', '#1abc9c', '#f39c12']
const color = (i) => colors[i % colors.length]

const wheelBg = computed(() => {
  const n = dishes.value.length
  if (!n) return '#eee'
  const stops = dishes.value.map((_, i) => `${color(i)} ${i * seg.value}deg ${(i + 1) * seg.value}deg`)
  return `conic-gradient(from 0deg, ${stops.join(', ')})`
})

function toggleMute() {
  muted.value = !muted.value
  sound.muted = muted.value
}

// 实时读取转盘角度，计算指针指向的扇区并在切换时播放“嗒”声
function track() {
  const n = dishes.value.length
  if (!n || !wheelEl.value) return
  const st = getComputedStyle(wheelEl.value).transform
  let deg = 0
  if (st && st !== 'none') {
    const m = st.match(/matrix\(([^)]+)\)/)
    if (m) {
      const [a, b] = m[1].split(',').map(Number)
      deg = Math.atan2(b, a) * (180 / Math.PI)
    }
  }
  const pointerAngle = ((-deg) % 360 + 360) % 360
  const idx = Math.floor(pointerAngle / seg.value) % n
  if (idx !== highlight.value) {
    highlight.value = idx
    tick()
  }
  if (spinning.value) rafId = requestAnimationFrame(track)
}

function spin() {
  const n = dishes.value.length
  if (n < 2 || spinning.value) return
  spinning.value = true
  result.value = null

  const idx = Math.floor(Math.random() * n)
  const center = idx * seg.value + seg.value / 2
  const desired = (360 - center) % 360
  const cur = ((rotation.value % 360) + 360) % 360
  const delta = (desired - cur + 360) % 360
  rotation.value += 5 * 360 + delta

  cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(track)

  setTimeout(() => {
    spinning.value = false
    cancelAnimationFrame(rafId)
    highlight.value = idx
    result.value = dishes.value[idx]
    fanfare()
    // 抽中后加入购物车并自动跳转到点餐页
    addToCart(result.value.id)
    toast(`抽中「${result.value.name}」，已加入购物车，正在前往点餐页…`)
    setTimeout(() => setTab && setTab('order'), 1800)
  }, 4200)
}

function spinAgain() {
  result.value = null
  spin()
}
</script>

<template>
  <div class="roulette">
    <div class="title-bar">
      <div>
        <h2>🎡 随机轮盘 · 今天吃什么</h2>
        <p class="hint">从已上架菜品中随机抽取一道，抽中后自动加入购物车并跳转到点餐页。</p>
      </div>
      <button class="btn-ghost" @click="toggleMute">{{ muted ? '🔇 已静音' : '🔊 音效开' }}</button>
    </div>

    <div v-if="dishes.length < 2" class="empty card">
      至少需要 2 道上架菜品才能转动轮盘，请先到「菜品管理」中上架。
    </div>

    <div v-else class="stage">
      <div class="wheel-wrap">
        <div class="pointer"></div>
        <div
          ref="wheelEl"
          class="wheel"
          :style="{ background: wheelBg, transform: `rotate(${rotation}deg)`, transition: spinning ? 'transform 4s cubic-bezier(0.18, 0.9, 0.2, 1)' : 'none' }"
        >
          <div
            v-for="(d, i) in dishes"
            :key="d.id"
            class="label-slot"
            :style="{ transform: `rotate(${i * seg + seg / 2}deg) translateY(-92px)` }"
          >
            <span
              class="label"
              :class="{ on: highlight === i }"
              :style="{ transform: `rotate(${-(i * seg + seg / 2)}deg)` }"
            >{{ d.name }}</span>
          </div>
        </div>
        <button class="go" :disabled="spinning" @click="spin">{{ spinning ? '…' : 'GO' }}</button>
      </div>

      <div class="side">
        <transition name="pop">
          <div v-if="result" class="result card">
            <div class="r-thumb" :style="{ backgroundImage: `url(${result.imageUrl})` }"></div>
            <div class="r-info">
              <div class="r-title">🎉 抽中了 <span v-if="result.kind === 'soup'" class="tag tag-orange">汤品</span></div>
              <div class="r-name">{{ result.name }}</div>
              <div class="r-desc">{{ result.description }}</div>
              <div class="r-ops">
                <button class="btn-ghost" @click="spinAgain">再转一次</button>
              </div>
            </div>
          </div>
          <div v-else class="result card placeholder">
            <span>点击 <strong>GO</strong> 开始转动转盘 👆</span>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.title-bar { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
h2 { margin: 0 0 6px; font-size: 18px; }
.hint { color: var(--muted); margin: 0; font-size: 14px; }
.stage { display: flex; gap: 40px; align-items: center; flex-wrap: wrap; }

.wheel-wrap { position: relative; width: 300px; height: 300px; flex-shrink: 0; }
.wheel {
  width: 300px; height: 300px; border-radius: 50%;
  border: 8px solid #fff;
  box-shadow: 0 8px 30px rgba(0,0,0,0.18), inset 0 0 0 2px rgba(0,0,0,0.05);
  position: relative;
}
.label-slot { position: absolute; top: 50%; left: 50%; width: 0; height: 0; transform-origin: 0 0; }
.label {
  position: absolute;
  transform-origin: center;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0,0,0,0.35);
  left: 50%;
  translate: -50% -50%;
  transition: font-size 0.1s, text-shadow 0.1s;
}
.label.on {
  font-size: 16px;
  color: #fff;
  text-shadow: 0 0 8px #fff, 0 1px 2px rgba(0,0,0,0.5);
}
.pointer {
  position: absolute;
  top: -10px; left: 50%; transform: translateX(-50%);
  width: 0; height: 0;
  border-left: 14px solid transparent;
  border-right: 14px solid transparent;
  border-top: 24px solid var(--primary);
  z-index: 5;
  filter: drop-shadow(0 2px 2px rgba(0,0,0,0.25));
}
.go {
  position: absolute;
  top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 64px; height: 64px; border-radius: 50%;
  background: #fff; color: var(--primary);
  font-weight: 800; font-size: 18px;
  box-shadow: 0 3px 10px rgba(0,0,0,0.25);
  z-index: 6;
  border: 3px solid var(--primary);
}
.go:disabled { color: var(--muted); border-color: var(--border); cursor: not-allowed; }

.side { flex: 1; min-width: 260px; }
.result { padding: 18px; display: flex; gap: 16px; min-height: 150px; align-items: center; }
.result.placeholder { justify-content: center; color: var(--muted); }
.r-thumb { width: 110px; height: 110px; border-radius: 12px; background-size: cover; background-position: center; background-color: #eee; flex-shrink: 0; }
.r-title { color: var(--primary); font-weight: 700; }
.r-name { font-size: 20px; font-weight: 700; margin: 4px 0; }
.r-desc { color: var(--muted); font-size: 13px; }
.r-price { color: var(--primary); font-weight: 700; font-size: 18px; margin: 8px 0; }
.r-ops { display: flex; gap: 10px; }
.empty { color: var(--muted); text-align: center; padding: 40px 0; }

.pop-enter-active { transition: all 0.3s ease; }
.pop-enter-from { opacity: 0; transform: scale(0.9); }
</style>
