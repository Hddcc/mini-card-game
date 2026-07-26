<script setup lang="ts">
/**
 * 战斗结算弹窗三态（迁移自旧 mini_2 showBattleResult）：
 * 胜利（金色勋章 + 扫光标题）/ 已投降（灰旗，由 failure_hint 含"投降"判定）/ 失败（红骷髅）。
 * 展示金币/经验奖励与剩余体力；胜利时提供"下一关"续战。
 */
import { computed } from 'vue'

import AppModal from '@/components/ui/AppModal.vue'
import { IMAGE_REWARD_GOLD } from '@/constants/heroAssets'
import type { BattleResult } from '@/types/battle'

const props = defineProps<{
  open: boolean
  result: BattleResult | null
  hasNextStage: boolean
}>()

defineEmits<{
  close: []
  backToPreparation: []
  nextStage: []
}>()

/** 旧版判定：非胜利且 failure_hint 含"投降" */
const surrendered = computed(() => Boolean(props.result && !props.result.win && /投降/.test(props.result.failure_hint || '')))

const statusMeta = computed(() => {
  if (!props.result) return null
  if (props.result.win) {
    return { title: '胜利', icon: 'military_tech', titleClass: 'shimmer-text', ringClass: 'border-primary-fixed', iconClass: 'text-primary-fixed' }
  }
  if (surrendered.value) {
    return { title: '已投降', icon: 'flag', titleClass: 'text-outline', ringClass: 'border-outline', iconClass: 'text-outline' }
  }
  return { title: '失败', icon: 'skull', titleClass: 'text-error', ringClass: 'border-error', iconClass: 'text-error' }
})

const subTitle = computed(() => {
  if (!props.result) return ''
  return props.result.win ? '挑战成功' : props.result.failure_hint || '挑战未通过'
})
</script>

<template>
  <AppModal :open="open" @close="$emit('close')">
    <div v-if="result && statusMeta" class="relative overflow-hidden">
      <div v-if="result.win" class="pointer-events-none absolute inset-0 bg-primary-fixed/10"></div>
      <div class="relative z-10 flex flex-col items-center p-stack-lg text-center">
        <div
          class="mb-stack-md flex h-20 w-20 items-center justify-center rounded-full border-4"
          :class="statusMeta.ringClass"
        >
          <span class="material-symbols-outlined text-4xl" :class="statusMeta.iconClass" aria-hidden="true">{{
            statusMeta.icon
          }}</span>
        </div>
        <h2 class="mb-1 font-display-hero text-headline-lg-mobile" :class="statusMeta.titleClass">
          {{ statusMeta.title }}
        </h2>
        <p class="mb-stack-lg font-label-sm text-label-sm uppercase tracking-widest text-on-surface-variant">
          {{ subTitle }}
        </p>

        <div class="mb-stack-lg w-full space-y-stack-md rounded border border-outline-variant bg-surface-container-high/50 p-stack-md">
          <div class="flex items-center justify-between">
            <span class="font-label-sm text-label-sm text-on-surface-variant">奖励</span>
            <div class="flex gap-4">
              <div class="flex items-center gap-1">
                <img :src="IMAGE_REWARD_GOLD" alt="金币" class="h-6 w-6 object-contain" />
                <span class="font-stats-num text-sm text-on-surface">+{{ result.reward_gold || 0 }}</span>
              </div>
              <div class="flex items-center gap-1">
                <span class="material-symbols-outlined text-lg text-quality-ssr" aria-hidden="true">military_tech</span>
                <span class="font-stats-num text-sm text-quality-ssr">+{{ result.reward_exp || 0 }} 经验</span>
              </div>
            </div>
          </div>
          <div class="h-px w-full bg-outline-variant"></div>
          <div class="flex items-center justify-between">
            <span class="font-label-sm text-label-sm text-on-surface-variant">剩余体力</span>
            <span class="font-stats-num text-sm text-primary-fixed"
              >{{ result.stamina || 0 }}/{{ result.max_stamina || 120 }}</span
            >
          </div>
        </div>

        <div class="flex w-full gap-stack-md">
          <button
            class="flex-1 border border-outline-variant py-3 font-label-sm text-label-sm transition-colors hover:bg-surface-variant"
            @click="$emit('backToPreparation')"
          >
            返回备战
          </button>
          <button
            v-if="result.win"
            class="flex-1 bg-primary-fixed py-3 font-label-sm text-label-sm font-bold text-on-primary-fixed shadow-[0_0_10px_rgba(255,215,0,0.3)] disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="!hasNextStage"
            :title="hasNextStage ? '挑战下一关' : '当前章节暂无可挑战的下一关'"
            @click="$emit('nextStage')"
          >
            {{ hasNextStage ? '下一关' : '暂无下一关' }}
          </button>
        </div>
      </div>
    </div>
  </AppModal>
</template>
