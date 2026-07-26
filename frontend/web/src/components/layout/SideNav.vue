<script setup lang="ts">
/**
 * 桌面端侧边导航（≥768px），替代旧 app-shell.js 注入的 iframe 壳侧栏。
 * 六个入口与旧版一致（旧手写侧栏漏掉的"活动"在此补齐）。
 * 战斗锁定时整体降透明并禁点（路由守卫兜底拦截）。
 */
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { useBattleStore } from '@/stores/battle'

interface NavItem {
  name: string
  label: string
  icon: string
}

const NAV_ITEMS: NavItem[] = [
  { name: 'home', label: '首页', icon: 'fort' },
  { name: 'heroes', label: '英雄', icon: 'swords' },
  { name: 'stages', label: '关卡', icon: 'scrollable_header' },
  { name: 'gacha', label: '召唤', icon: 'auto_awesome' },
  { name: 'starUp', label: '升星', icon: 'upgrade' },
  { name: 'activity', label: '活动', icon: 'celebration' },
]

const route = useRoute()
const router = useRouter()
const battleStore = useBattleStore()
const authStore = useAuthStore()

function isActive(item: NavItem): boolean {
  return route.name === item.name
}

function logout(): void {
  authStore.logout()
  void router.push({ name: 'login' })
}
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-40 hidden w-72 flex-col border-r border-outline-variant bg-surface-container-lowest md:flex"
    :class="battleStore.isLocked ? 'pointer-events-none opacity-55' : ''"
  >
    <div class="flex items-center gap-stack-md border-b border-outline-variant px-gutter py-stack-lg">
      <div
        class="flex h-12 w-12 items-center justify-center rounded-full border-2 border-primary-container bg-ink-wash shadow-[0_0_20px_rgba(255,215,0,0.2)]"
      >
        <span class="material-symbols-outlined text-2xl text-primary-fixed" aria-hidden="true">temple_buddhist</span>
      </div>
      <div>
        <p class="font-display-hero text-title-md text-primary-fixed glow-gold">Mini 西游</p>
        <p class="font-label-sm text-[10px] uppercase tracking-widest text-on-surface-variant">Mythic Journey</p>
      </div>
    </div>

    <nav class="flex-1 space-y-1 overflow-y-auto thin-scrollbar px-3 py-stack-lg">
      <router-link
        v-for="item in NAV_ITEMS"
        :key="item.name"
        :to="{ name: item.name }"
        class="flex items-center gap-stack-md border-l-2 px-4 py-3 font-title-md text-sm transition-colors"
        :class="
          isActive(item)
            ? 'border-primary-fixed bg-ink-wash text-primary-fixed'
            : 'border-transparent text-on-surface-variant hover:bg-ink-wash/60 hover:text-on-surface'
        "
      >
        <span
          class="material-symbols-outlined"
          :class="isActive(item) ? 'icon-filled' : ''"
          aria-hidden="true"
          >{{ item.icon }}</span
        >
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <div class="border-t border-outline-variant p-3">
      <button
        class="focus-ring flex w-full items-center gap-stack-md px-4 py-3 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:text-error"
        @click="logout"
      >
        <span class="material-symbols-outlined" aria-hidden="true">logout</span>
        退出登录
      </button>
    </div>
  </aside>
</template>
