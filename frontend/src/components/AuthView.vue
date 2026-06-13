<script setup>
import { ref, inject } from 'vue'
import { api } from '../api'
import { setAuth } from '../store'

const toast = inject('toast')
const isLogin = ref(true)
const username = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  if (!username.value.trim() || !password.value) {
    toast('请输入用户名和密码', 'error'); return
  }
  loading.value = true
  try {
    const fn = isLogin.value ? api.login : api.register
    const result = await fn({ username: username.value.trim(), password: password.value })
    setAuth(result.token, result.user)
    toast(isLogin.value ? '登录成功' : '注册成功')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-wrap">
    <div class="auth-card card">
      <h1 class="brand">🍜 小馆点餐</h1>
      <h2>{{ isLogin ? '登录' : '注册' }}</h2>
      <form @submit.prevent="submit">
        <div class="field">
          <label>用户名</label>
          <input v-model="username" placeholder="请输入用户名" />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="请输入密码" />
          <p v-if="!isLogin" class="hint">至少8位，包含大小写字母、数字和特殊字符</p>
        </div>
        <button type="submit" class="btn-primary btn-block" :disabled="loading">
          {{ loading ? '请稍候…' : (isLogin ? '登录' : '注册') }}
        </button>
      </form>
      <p class="toggle">
        {{ isLogin ? '没有账号？' : '已有账号？' }}
        <a href="#" @click.prevent="isLogin = !isLogin">{{ isLogin ? '去注册' : '去登录' }}</a>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-wrap { display: flex; align-items: center; justify-content: center; min-height: 80vh; }
.auth-card { width: 380px; max-width: 92vw; padding: 40px 32px; text-align: center; }
.brand { font-size: 24px; margin: 0 0 24px; }
h2 { font-size: 18px; margin: 0 0 20px; }
.field { text-align: left; margin-bottom: 16px; }
.field label { font-size: 13px; color: var(--muted); margin-bottom: 6px; display: block; }
.hint { font-size: 12px; color: var(--muted); margin: 6px 0 0; }
.btn-block { width: 100%; padding: 12px; font-size: 16px; margin-top: 8px; }
.toggle { margin-top: 18px; font-size: 14px; color: var(--muted); }
.toggle a { color: var(--primary); text-decoration: none; font-weight: 600; }
</style>
