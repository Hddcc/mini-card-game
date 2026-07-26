<script setup lang="ts">
/** 移动端底部导航（<768px）。旧版移动端直接隐藏导航，这里改为可用的底部 tab（改进项）。 */
import { useRoute } from 'vue-router'

import { useBattleStore } from '@/stores/battle'

const NAV_ITEMS = [
  { name: 'home', label: '首页', icon: 'fort' },
  { name: 'heroes', label: '英雄', icon: 'swords' },
  { name: 'stages', label: '关卡', icon: 'scrollable_header' },
  { name: 'gacha', label: '召唤', icon: 'auto_awesome' },
  { name: 'starUp', label: '升星', icon: 'upgrade' },
  { name: 'activity', label: '活动', icon: 'celebration' },
] as const

const route = useRoute()
const battleStore = useBattleStore()
</script>

<template>
  <nav
    class="fixed bottom-0 z-40 flex w-full items-center justify-around border-t border-outline-variant bg-ink-wash px-2 py-2 md:hidden"
    :class="battleStore.isLocked ? 'pointer-events-none opacity-55' : ''"
  >
    <router-link
      v-for="item in NAV_ITEMS"
      :key="item.name"
      :to="{ name: item.name }"
      class="flex flex-col items-center gap-1 py-1 transition-transform"
      :class="route.name === item.name ? 'scale-110 text-primary-fixed' : 'text-on-surface-variant'"
    >
      <span
        class="material-symbols-outlined"
        :class="route.name === item.name ? 'icon-filled' : ''"
        aria-hidden="true"
        >{{ item.icon }}</span
      >
      <span class="font-label-sm text-[10px]" :class="route.name === item.name ? 'font-bold' : ''">{{
        item.label
      }}</span>
    </router-link>
  </nav>
</template>
