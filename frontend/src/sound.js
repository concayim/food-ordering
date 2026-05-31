// 轻量音效：基于 Web Audio API，无需音频文件
let ctx = null
export const sound = { muted: false }

function ac() {
  if (!ctx) ctx = new (window.AudioContext || window.webkitAudioContext)()
  if (ctx.state === 'suspended') ctx.resume()
  return ctx
}

function beep(freq, duration = 0.06, type = 'square', gain = 0.05) {
  if (sound.muted) return
  try {
    const c = ac()
    const osc = c.createOscillator()
    const g = c.createGain()
    osc.type = type
    osc.frequency.value = freq
    g.gain.value = gain
    osc.connect(g)
    g.connect(c.destination)
    const t = c.currentTime
    osc.start(t)
    g.gain.setValueAtTime(gain, t)
    g.gain.exponentialRampToValueAtTime(0.0001, t + duration)
    osc.stop(t + duration)
  } catch (e) {
    /* ignore */
  }
}

// 转盘经过一个扇区时的“嗒”声
export function tick() {
  beep(880, 0.04, 'square', 0.04)
}

// 抽中时的小段上扬旋律
export function fanfare() {
  if (sound.muted) return
  const notes = [523, 659, 784, 1047] // C5 E5 G5 C6
  notes.forEach((f, i) => setTimeout(() => beep(f, 0.16, 'triangle', 0.08), i * 110))
}
