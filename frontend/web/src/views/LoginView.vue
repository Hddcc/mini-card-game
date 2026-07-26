<script setup lang="ts">
/**
 * 登录 / 注册（迁移自旧 mini_1）：
 * - tab 切换（注册态多昵称字段）
 * - 注册成功行内绿色提示，900ms 后切回登录 tab
 * - 登录成功存 token 并跳转（支持 ?redirect= 回跳）
 * - 成功时卡片 bounce 动效、金色粒子背景
 */
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppInput from '@/components/ui/AppInput.vue'
import GoldParticles from '@/components/ui/GoldParticles.vue'
import InkBackground from '@/components/ui/InkBackground.vue'
import JadePanel from '@/components/ui/JadePanel.vue'
import StoneButton from '@/components/ui/StoneButton.vue'
import { useAuthStore } from '@/stores/auth'
import { errorText } from '@/types/api'

type Mode = 'login' | 'register'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const mode = ref<Mode>('login')
const username = ref('')
const password = ref('')
const nickname = ref('')
const submitting = ref(false)
const bouncing = ref(false)

const alert = ref<{ type: 'success' | 'error'; text: string } | null>(null)

function switchMode(next: Mode): void {
  mode.value = next
  alert.value = null
}

function bounceCard(): void {
  bouncing.value = true
  setTimeout(() => {
    bouncing.value = false
  }, 500)
}

async function handleSubmit(): Promise<void> {
  if (submitting.value) return
  submitting.value = true
  alert.value = null
  try {
    if (mode.value === 'login') {
      await authStore.login({ username: username.value, password: password.value })
      alert.value = { type: 'success', text: '登录成功' }
      bounceCard()
      const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
      await router.push(redirect || { name: 'home' })
    } else {
      await authStore.register({
        username: username.value,
        password: password.value,
        nickname: nickname.value.trim(),
      })
      alert.value = { type: 'success', text: '注册成功' }
      bounceCard()
      setTimeout(() => {
        switchMode('login')
      }, 900)
    }
  } catch (error) {
    alert.value = { type: 'error', text: errorText(error, '请求失败，请检查输入') }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden p-gutter">
    <InkBackground />
    <GoldParticles />

    <main class="relative z-10 flex w-full max-w-[420px] flex-col gap-stack-lg">
      <!-- 品牌区 -->
      <header class="flex flex-col items-center gap-stack-md text-center">
        <div
          class="flex h-20 w-20 items-center justify-center rounded-full border-2 border-primary-container bg-ink-wash shadow-[0_0_20px_rgba(255,215,0,0.2)]"
        >
          <span class="material-symbols-outlined text-[40px] text-primary-fixed" aria-hidden="true"
            >temple_buddhist</span
          >
        </div>
        <div>
          <h1 class="glow-gold mb-1 font-display-hero text-headline-lg-mobile text-primary-fixed md:text-display-hero">
            Mini 西游
          </h1>
          <p class="font-label-sm text-label-sm uppercase tracking-widest text-on-surface-variant">
            古老智慧 · 现代征途
          </p>
        </div>
      </header>

      <!-- 认证卡片 -->
      <JadePanel class="p-stack-lg shadow-2xl" :class="bouncing ? 'animate-bounce' : ''">
        <div class="mb-stack-lg flex border-b border-outline-variant">
          <button
            class="flex-1 border-b-2 py-stack-md font-title-md text-title-md transition-all"
            :class="
              mode === 'login'
                ? 'border-primary-fixed text-primary-fixed'
                : 'border-transparent text-on-surface-variant'
            "
            @click="switchMode('login')"
          >
            登录
          </button>
          <button
            class="flex-1 border-b-2 py-stack-md font-title-md text-title-md transition-all"
            :class="
              mode === 'register'
                ? 'border-primary-fixed text-primary-fixed'
                : 'border-transparent text-on-surface-variant'
            "
            @click="switchMode('register')"
          >
            注册
          </button>
        </div>

        <div
          v-if="alert"
          class="mb-stack-md flex items-center gap-stack-md border p-stack-md"
          :class="
            alert.type === 'success'
              ? 'border-tertiary-container bg-on-tertiary-container/40 text-tertiary-fixed'
              : 'animate-pulse border-secondary bg-secondary-container text-on-secondary-container'
          "
        >
          <span class="material-symbols-outlined text-[18px]" aria-hidden="true">{{
            alert.type === 'success' ? 'check_circle' : 'error'
          }}</span>
          <span class="font-label-sm text-label-sm">{{ alert.text }}</span>
        </div>

        <form class="flex flex-col gap-stack-md" @submit.prevent="handleSubmit">
          <div class="space-y-stack-xs">
            <label class="ml-1 font-label-sm text-label-sm uppercase text-on-surface-variant">账号</label>
            <AppInput v-model="username" icon="person" name="username" placeholder="请输入账号" />
          </div>

          <div v-if="mode === 'register'" class="space-y-stack-xs">
            <label class="ml-1 font-label-sm text-label-sm uppercase text-on-surface-variant">昵称</label>
            <AppInput v-model="nickname" icon="badge" name="nickname" placeholder="请输入昵称" />
          </div>

          <div class="space-y-stack-xs">
            <label class="ml-1 font-label-sm text-label-sm uppercase text-on-surface-variant">密码</label>
            <AppInput v-model="password" icon="lock" name="password" type="password" placeholder="请输入密码" />
          </div>

          <div class="mt-stack-md">
            <StoneButton type="submit" size="lg" block :disabled="submitting" class="group">
              <span>{{ submitting ? '请稍候…' : mode === 'login' ? '登录' : '注册' }}</span>
              <span
                class="material-symbols-outlined transition-transform group-hover:translate-x-1"
                aria-hidden="true"
                >swords</span
              >
            </StoneButton>
          </div>
        </form>
      </JadePanel>

      <!-- 页脚点缀 -->
      <div
        class="flex items-center justify-around opacity-60 grayscale transition-all duration-700 hover:grayscale-0"
      >
        <div class="flex flex-col items-center gap-1">
          <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">verified_user</span>
          <span class="font-label-sm text-[10px]">安全保障</span>
        </div>
        <div class="flex flex-col items-center gap-1">
          <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">public</span>
          <span class="font-label-sm text-[10px]">99+ 领域</span>
        </div>
        <div class="flex flex-col items-center gap-1">
          <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">auto_awesome</span>
          <span class="font-label-sm text-[10px]">神赐随机</span>
        </div>
      </div>
    </main>
  </div>
</template>
