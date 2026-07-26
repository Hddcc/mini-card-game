<script setup lang="ts">
/**
 * 顶部资产栏：金币 / 灵玉 / 体力 + hover 提示与体力恢复倒计时。
 * 迁移自旧 status-bar.js（数据改由 useAssetStore 驱动）。
 */
import { computed } from 'vue'

import { IMAGE_REWARD_GOLD } from '@/constants/heroAssets'
import { useAssetStore } from '@/stores/assets'

const assetStore = useAssetStore()

const staminaTooltip = computed(() => {
  if (assetStore.staminaFull || assetStore.countdown <= 0) return '体力'
  const minutes = Math.floor(assetStore.countdown / 60)
  const seconds = assetStore.countdown % 60
  return `体力 / ${minutes}:${String(seconds).padStart(2, '0')} 后恢复 1 点`
})

function formatNumber(value: number): string {
  return value.toLocaleString('zh-CN')
}
</script>

<template>
  <div class="flex items-center gap-2 md:gap-3">
    <div
      class="group relative flex items-center gap-1.5 rounded-full border border-outline-variant bg-surface-container-low px-3 py-1.5"
    >
      <img :src="IMAGE_REWARD_GOLD" alt="金币" class="h-4 w-4 object-contain" />
      <span class="font-stats-num text-sm text-on-surface">{{ formatNumber(assetStore.gold) }}</span>
      <span
        class="pointer-events-none absolute left-1/2 top-full z-50 mt-1 -translate-x-1/2 whitespace-nowrap rounded border border-outline-variant bg-[#1b1711]/95 px-2 py-1 font-label-sm text-[10px] text-on-surface opacity-0 transition-opacity group-hover:opacity-100"
        >金币</span
      >
    </div>

    <div
      class="group relative flex items-center gap-1.5 rounded-full border border-outline-variant bg-surface-container-low px-3 py-1.5"
    >
      <span class="material-symbols-outlined icon-filled text-base text-quality-sr" aria-hidden="true">diamond</span>
      <span class="font-stats-num text-sm text-on-surface">{{ formatNumber(assetStore.diamond) }}</span>
      <span
        class="pointer-events-none absolute left-1/2 top-full z-50 mt-1 -translate-x-1/2 whitespace-nowrap rounded border border-outline-variant bg-[#1b1711]/95 px-2 py-1 font-label-sm text-[10px] text-on-surface opacity-0 transition-opacity group-hover:opacity-100"
        >灵玉</span
      >
    </div>

    <div
      class="group relative flex items-center gap-1.5 rounded-full border border-outline-variant bg-surface-container-low px-3 py-1.5"
    >
      <span class="material-symbols-outlined icon-filled text-base text-tertiary-container" aria-hidden="true"
        >bolt</span
      >
      <span class="font-stats-num text-sm text-on-surface"
        >{{ assetStore.stamina }}/{{ assetStore.maxStamina }}</span
      >
      <span
        class="pointer-events-none absolute left-1/2 top-full z-50 mt-1 -translate-x-1/2 whitespace-nowrap rounded border border-outline-variant bg-[#1b1711]/95 px-2 py-1 font-label-sm text-[10px] text-on-surface opacity-0 transition-opacity group-hover:opacity-100"
        >{{ staminaTooltip }}</span
      >
    </div>
  </div>
</template>
