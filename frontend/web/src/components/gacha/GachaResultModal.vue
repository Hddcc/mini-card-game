<script setup lang="ts">
/** 抽卡结果全屏弹窗（迁移自旧 mini_4 result-modal）：召唤中 / 错误（钻石不足特判）/ 结果三态。 */
import GachaResultCard from './GachaResultCard.vue'
import type { DrawResult } from '@/types/gacha'

export interface GachaModalState {
  open: boolean
  phase: 'loading' | 'error' | 'results'
  title: string
  errorBody: string
  results: DrawResult[]
}

defineProps<{ state: GachaModalState }>()

defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="state.open" class="fixed inset-0 z-[100] flex items-center justify-center p-gutter">
      <div class="absolute inset-0 bg-black/95 backdrop-blur-xl"></div>
      <div class="relative z-10 flex w-full max-w-5xl flex-col items-center">
        <h2 class="mb-stack-lg font-display-hero text-headline-lg-mobile text-primary-fixed md:text-headline-lg">
          {{ state.title }}
        </h2>

        <div
          class="mb-12 grid w-full gap-stack-lg"
          :class="
            state.phase === 'results' && state.results.length > 1
              ? 'grid-cols-2 md:grid-cols-5'
              : 'grid-cols-1 justify-items-center'
          "
        >
          <div v-if="state.phase === 'loading'" class="col-span-full text-center text-primary-fixed">召唤中...</div>
          <div v-else-if="state.phase === 'error' && state.errorBody" class="col-span-full text-center text-error">
            {{ state.errorBody }}
          </div>
          <template v-else-if="state.phase === 'results'">
            <GachaResultCard
              v-for="(item, index) in state.results"
              :key="index"
              :item="item"
              :index="index"
              :single="state.results.length === 1"
            />
          </template>
        </div>

        <button
          v-if="state.phase !== 'loading'"
          class="stone-button border-2 border-primary-fixed bg-surface-container-highest px-12 py-3 font-title-md text-title-md text-primary-fixed shadow-[0_0_20px_rgba(255,215,0,0.3)] transition-all hover:scale-105 hover:bg-primary-fixed hover:text-black active:scale-95"
          @click="$emit('close')"
        >
          确认
        </button>
      </div>
    </div>
  </Teleport>
</template>
