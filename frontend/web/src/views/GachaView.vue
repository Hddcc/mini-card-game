<script setup lang="ts">
/**
 * 天命召唤（迁移自旧 mini_4）：
 * - 保底进度（GET /gacha/state）
 * - 单抽 160 / 十连 1600（POST /gacha/draw，times 1|10）
 * - 结果弹窗逐张揭示；重复神将碎片化；保底标记
 * - "灵玉不足"错误特判标题（无正文），其余"召唤失败：<原因>"
 */
import { onMounted, reactive, ref } from 'vue'

import { drawGacha, fetchGachaState } from '@/api/gacha'
import GachaResultModal, { type GachaModalState } from '@/components/gacha/GachaResultModal.vue'
import JadeProgressBar from '@/components/ui/JadeProgressBar.vue'
import { GACHA_COST_ONE, GACHA_COST_TEN, GACHA_PITY_LIMIT_FALLBACK, GACHA_POOL_ID } from '@/constants'
import { HERO_IMAGES } from '@/constants/heroAssets'
import { useAssetStore } from '@/stores/assets'
import { ApiError, errorText } from '@/types/api'

const assetStore = useAssetStore()

const pityCounter = ref(0)
const pityLimit = ref(GACHA_PITY_LIMIT_FALLBACK)
const drawing = ref(false)

const modal = reactive<GachaModalState>({
  open: false,
  phase: 'loading',
  title: '恭喜你获得',
  errorBody: '',
  results: [],
})

function updatePity(counter: number, limit: number): void {
  pityCounter.value = Math.max(0, Number(counter) || 0)
  pityLimit.value = Math.max(1, Number(limit) || GACHA_PITY_LIMIT_FALLBACK)
}

async function loadPityState(): Promise<void> {
  try {
    const state = await fetchGachaState(GACHA_POOL_ID)
    updatePity(state.pity_counter, state.pity_limit)
  } catch {
    updatePity(0, GACHA_PITY_LIMIT_FALLBACK)
  }
}

onMounted(loadPityState)

async function draw(times: 1 | 10): Promise<void> {
  if (drawing.value) return
  drawing.value = true
  modal.open = true
  modal.phase = 'loading'
  modal.title = '恭喜你获得'
  modal.errorBody = ''
  modal.results = []
  try {
    const output = await drawGacha(GACHA_POOL_ID, times)
    modal.phase = 'results'
    modal.title = '恭喜你获得'
    modal.results = output.results ?? []
    updatePity(output.pity_counter, output.pity_limit)
    void assetStore.refresh().catch(() => undefined)
  } catch (error) {
    modal.phase = 'error'
    // 旧版特判：灵玉不足只显示标题（后端 message 仍是 diamond not enough）
    const raw = error instanceof ApiError ? error.raw.toLowerCase() : ''
    if (raw.includes('diamond not enough') || raw.includes('not enough diamond')) {
      modal.title = '灵玉不足'
      modal.errorBody = ''
    } else {
      modal.title = '召唤失败'
      modal.errorBody = `召唤失败：${errorText(error, '未知错误')}`
    }
  } finally {
    drawing.value = false
  }
}

function closeResults(): void {
  modal.open = false
  void assetStore.refresh().catch(() => undefined)
}
</script>

<template>
  <div class="mx-auto flex max-w-5xl flex-col items-center gap-stack-lg">
    <!-- 卡池头部信息 + 保底面板 -->
    <div class="flex w-full flex-col items-start justify-between gap-stack-lg md:flex-row md:items-end">
      <div class="space-y-stack-xs">
        <div class="flex items-center gap-stack-md">
          <span class="shimmer-sweep rounded bg-quality-ur px-3 py-1 font-label-sm text-label-sm text-on-surface"
            >限定</span
          >
        </div>
        <h2 class="font-display-hero text-headline-lg text-primary md:text-display-hero">天命龙影</h2>
        <p class="max-w-lg font-body-md text-body-md text-on-surface-variant">
          天命灵息已被唤醒，消耗灵玉召来传说英雄，加入你的取经之路。
        </p>
      </div>

      <div class="relative min-w-[240px] border border-outline-variant bg-surface-container-high/80 p-stack-lg backdrop-blur-md">
        <div class="meander-corner meander-tl"></div>
        <div class="meander-corner meander-tr"></div>
        <div class="meander-corner meander-bl"></div>
        <div class="meander-corner meander-br"></div>
        <div class="flex flex-col gap-stack-md">
          <div class="flex items-center justify-between">
            <span class="font-label-sm text-label-sm text-on-surface-variant">保底计数</span>
            <span class="font-stats-num text-stats-num text-primary-fixed">{{ pityCounter }} / {{ pityLimit }}</span>
          </div>
          <JadeProgressBar :value="pityCounter" :max="pityLimit" variant="gold" height-class="h-3" />
          <span class="text-center font-label-sm text-[10px] italic text-on-surface-variant"
            >{{ pityLimit }} 次召唤必得高品质英雄</span
          >
        </div>
      </div>
    </div>

    <!-- 主打英雄立绘 -->
    <div class="relative flex w-full flex-1 items-center justify-center py-4">
      <div class="group relative">
        <div
          class="absolute -inset-10 rounded-full bg-primary-fixed/20 blur-3xl transition-all duration-700 group-hover:scale-125"
        ></div>
        <div
          class="relative aspect-[3/4] w-64 transform overflow-hidden rounded-xl border-2 border-primary-fixed shadow-2xl transition-transform hover:scale-105 md:w-96"
        >
          <img
            :src="HERO_IMAGES[1]"
            alt="孙悟空主打英雄立绘"
            class="h-full w-full bg-surface-container object-cover"
          />
          <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black p-stack-lg">
            <span class="font-display-hero text-headline-lg-mobile italic text-quality-ssr">孙悟空</span>
            <div class="mt-1 flex gap-1">
              <span
                v-for="star in 5"
                :key="star"
                class="material-symbols-outlined icon-filled text-sm text-quality-ssr"
                aria-hidden="true"
                >star</span
              >
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 召唤操作 -->
    <div class="flex w-full max-w-3xl flex-col gap-gutter pb-stack-lg md:flex-row">
      <button
        class="group relative flex-1 overflow-hidden border-2 border-outline bg-ink-wash p-6 transition-all hover:border-primary-fixed active:scale-95 disabled:opacity-60"
        :disabled="drawing"
        @click="draw(1)"
      >
        <div class="relative z-10 flex flex-col items-center gap-2">
          <span class="font-title-md text-title-md text-on-surface">召唤一次</span>
          <div class="flex items-center gap-2 text-primary-fixed">
            <span class="material-symbols-outlined text-sm" aria-hidden="true">diamond</span>
            <span class="font-stats-num text-stats-num">{{ GACHA_COST_ONE }}</span>
          </div>
        </div>
        <div class="absolute inset-0 bg-primary-fixed/5 opacity-0 transition-opacity group-hover:opacity-100"></div>
      </button>

      <button
        class="group relative flex-1 overflow-hidden border-2 border-primary-fixed bg-surface-container-highest p-6 shadow-[0_0_20px_rgba(255,215,0,0.2)] transition-all active:scale-95 disabled:opacity-60"
        :disabled="drawing"
        @click="draw(10)"
      >
        <div class="relative z-10 flex flex-col items-center gap-2">
          <span class="shimmer-sweep font-title-md text-title-md text-primary-fixed">召唤十次</span>
          <div class="flex items-center gap-2 text-primary-fixed">
            <span class="material-symbols-outlined text-sm" aria-hidden="true">diamond</span>
            <span class="font-stats-num text-stats-num">{{ GACHA_COST_TEN }}</span>
          </div>
        </div>
        <div class="absolute inset-0 bg-primary-fixed/10 opacity-0 transition-opacity group-hover:opacity-100"></div>
        <div class="absolute -right-4 -top-4 rotate-45 bg-quality-ur px-4 py-1 font-label-sm text-[10px]">额外 +1</div>
      </button>
    </div>

    <GachaResultModal :state="modal" @close="closeResults" />
  </div>
</template>
