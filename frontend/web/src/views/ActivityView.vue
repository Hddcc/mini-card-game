<script setup lang="ts">
/**
 * 限时活动·福签抽奖（迁移自旧 mini_6）：
 * - state 单包驱动：活动信息 / 截止时间 / 每日配额条 / 奖品余量 / 内嵌中奖记录 / eligible 门控
 * - 抽奖空请求体；失败原因走关键字规则映射为友好中文
 * - 发奖异步（delivery_status 0-3 文案），抽完与关弹窗时重拉 state 与资产
 */
import { computed, onMounted, ref } from 'vue'

import { drawActivityLottery, fetchActivityState } from '@/api/activity'
import { deliveryStatusText, friendlyActivityReason } from '@/constants/messages'
import { IMAGE_ACTIVITY_BANNER, IMAGE_REWARD_GOLD } from '@/constants/heroAssets'
import { useAssetStore } from '@/stores/assets'
import { errorRaw, errorText } from '@/types/api'
import type { ActivityStateView } from '@/types/activity'

const assetStore = useAssetStore()

const state = ref<ActivityStateView | null>(null)
const loadFailed = ref(false)
const drawing = ref(false)

const resultModal = ref<{ open: boolean; icon: string; title: string; description: string }>({
  open: false,
  icon: 'redeem',
  title: '',
  description: '',
})

const ICON_MAP: Record<string, string> = {
  gold: 'account_balance_wallet',
  diamond: 'diamond',
  stamina: 'bolt',
  hero: 'auto_awesome',
}

const active = computed(() => Boolean(state.value?.active && state.value.activity))
const activityName = computed(() => (active.value ? state.value!.activity!.name : '暂无活动'))
const activityDesc = computed(() =>
  active.value ? state.value!.activity!.description : '天庭正在筹备下一场赐福。',
)
const timeText = computed(() => {
  if (!active.value) return ''
  return `截止 ${new Date(state.value!.activity!.end_at).toLocaleString()}`
})
const remainingToday = computed(() => state.value?.quota?.remaining_today ?? 0)
const quotaPercent = computed(() => {
  const used = state.value?.quota?.used_today ?? 0
  const limit = Math.max(1, state.value?.quota?.daily_limit ?? 1)
  return Math.min(100, Math.round((used / limit) * 100))
})
const eligibilityText = computed(() => {
  if (loadFailed.value) return '活动数据加载失败，请稍后重试'
  if (!state.value) return '加载活动资格中...'
  if (!state.value.active) return state.value.reason ? friendlyActivityReason(state.value.reason) : '暂无可参与活动'
  return state.value.eligible ? '可以抽取福签，中奖奖励会陆续到账。' : friendlyActivityReason(state.value.reason || '暂不可参与')
})
const drawDisabled = computed(() => drawing.value || !state.value?.eligible)

async function loadState(): Promise<void> {
  loadFailed.value = false
  try {
    state.value = await fetchActivityState()
    void assetStore.refresh().catch(() => undefined)
  } catch {
    loadFailed.value = true
  }
}

onMounted(loadState)

async function handleDraw(): Promise<void> {
  if (drawDisabled.value) return
  drawing.value = true
  try {
    const output = await drawActivityLottery()
    resultModal.value = {
      open: true,
      icon: ICON_MAP[output.prize.reward_type] || output.prize.icon || 'redeem',
      title: output.prize.name,
      description: output.prize.description,
    }
    await loadState()
  } catch (error) {
    resultModal.value = {
      open: true,
      icon: 'error',
      title: '抽取失败',
      description: friendlyActivityReason(errorRaw(error) || errorText(error, '请稍后重试')),
    }
  } finally {
    drawing.value = false
  }
}

function closeResult(): void {
  resultModal.value.open = false
  void loadState()
}
</script>

<template>
  <div class="mx-auto grid max-w-7xl grid-cols-1 gap-stack-lg xl:grid-cols-[1.2fr_.8fr]">
    <!-- 活动主区 -->
    <section class="flex flex-col justify-center gap-stack-lg">
      <div class="grid items-end gap-stack-lg md:grid-cols-[1fr_220px]">
        <div>
          <div class="mb-4 flex items-center gap-3">
            <span
              class="rounded-full border border-primary-fixed/60 bg-primary-fixed px-3 py-1 font-label-sm text-label-sm text-on-primary-fixed"
              >限时活动</span
            >
            <span class="font-label-sm text-sm text-on-surface-variant">{{ timeText || '活动时间加载中' }}</span>
          </div>
          <h2 class="font-display-hero text-headline-lg leading-tight text-primary-fixed md:text-display-hero">
            {{ activityName }}
          </h2>
          <p class="mt-4 max-w-2xl font-body-md text-title-md text-on-surface-variant">{{ activityDesc }}</p>
        </div>
        <div class="rounded-xl border border-outline-variant bg-surface-container/80 p-5">
          <p class="font-label-sm text-label-sm text-on-surface-variant">今日次数</p>
          <div class="mt-2 flex items-end gap-2">
            <span class="font-stats-num text-[44px] leading-none text-primary-fixed">{{ remainingToday }}</span>
            <span class="pb-1 text-on-surface-variant">次可用</span>
          </div>
          <div class="mt-4 h-2 overflow-hidden rounded-full bg-surface-container-lowest">
            <div class="h-full bg-primary-fixed transition-all" :style="{ width: `${quotaPercent}%` }"></div>
          </div>
        </div>
      </div>

      <!-- 抽签台 -->
      <div class="relative min-h-[360px] overflow-hidden rounded-xl border border-outline-variant bg-ink-wash/80">
        <img
          :src="IMAGE_ACTIVITY_BANNER"
          alt="火焰山活动"
          class="pointer-events-none absolute inset-0 h-full w-full object-cover object-center opacity-45"
        />
        <div class="pointer-events-none absolute inset-0 bg-gradient-to-r from-black/60 via-ink-wash/45 to-black/70"></div>
        <div class="relative z-10 grid h-full items-center gap-stack-lg p-6 md:grid-cols-[1fr_300px] md:p-10">
          <div class="flex justify-center">
            <button
              class="talisman draw-glow flex h-72 w-60 flex-col items-center justify-center gap-4 rounded-xl border-2 border-white/35 text-on-primary-fixed transition-transform active:scale-95"
              :class="drawDisabled ? 'cursor-not-allowed opacity-50' : ''"
              :disabled="drawDisabled"
              @click="handleDraw"
            >
              <span class="material-symbols-outlined icon-filled text-6xl" aria-hidden="true">redeem</span>
              <span class="font-display-hero text-headline-lg-mobile font-extrabold">{{
                drawing ? '抽取中' : '抽取福签'
              }}</span>
              <span class="font-label-sm text-label-sm">每日限次 · 好运加持</span>
            </button>
          </div>
          <div class="space-y-4">
            <div class="rounded-lg border border-outline-variant bg-surface-container-lowest/80 p-4">
              <p class="font-title-md text-title-md text-primary-fixed">今日赐福</p>
              <div class="mt-3 grid grid-cols-2 gap-2 text-sm text-on-surface-variant">
                <span>参与次数</span><span class="text-right text-tertiary-container">每日刷新</span>
                <span>奖励类型</span><span class="text-right text-tertiary-container">多种好礼</span>
                <span>活动状态</span><span class="text-right text-tertiary-container">限时开放</span>
                <span>到账提示</span><span class="text-right text-primary-fixed">稍后查看</span>
              </div>
            </div>
            <div class="rounded-lg border border-outline-variant bg-surface-container-lowest/80 p-4">
              <p class="font-title-md text-title-md text-primary-fixed">活动说明</p>
              <p class="mt-2 text-sm text-on-surface-variant">{{ eligibilityText }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 奖品池 + 中奖记录 -->
    <aside class="space-y-stack-lg xl:pt-8">
      <section class="rounded-xl border border-outline-variant bg-surface-container/80 p-5">
        <div class="mb-4 flex items-center justify-between">
          <h3 class="font-title-md text-title-md text-primary-fixed">奖品池</h3>
          <span class="font-label-sm text-label-sm text-on-surface-variant">动态库存</span>
        </div>
        <div class="space-y-3">
          <p v-if="loadFailed" class="text-sm text-error">奖品池加载失败</p>
          <p v-else-if="!state?.prizes?.length" class="text-sm text-on-surface-variant">暂无奖品信息</p>
          <div
            v-for="prize in state?.prizes ?? []"
            :key="prize.id"
            class="flex items-center justify-between gap-3 rounded-lg border border-outline-variant bg-surface-container-lowest/70 p-3"
          >
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded bg-surface-container-high text-primary-fixed"
              >
                <img v-if="prize.reward_type === 'gold'" :src="IMAGE_REWARD_GOLD" alt="" class="h-8 w-8 object-contain" />
                <span v-else class="material-symbols-outlined" aria-hidden="true">{{
                  ICON_MAP[prize.reward_type] || prize.icon || 'redeem'
                }}</span>
              </div>
              <div class="min-w-0">
                <p class="truncate font-title-md text-body-md text-on-surface">{{ prize.name }}</p>
                <p class="truncate text-xs text-on-surface-variant">{{ prize.description }}</p>
              </div>
            </div>
            <span class="whitespace-nowrap font-label-sm text-label-sm text-primary-fixed">{{
              prize.unlimited ? '不限量' : `余 ${Math.max(0, prize.left_num)}`
            }}</span>
          </div>
        </div>
      </section>

      <section class="rounded-xl border border-outline-variant bg-surface-container/80 p-5">
        <div class="mb-4 flex items-center justify-between">
          <h3 class="font-title-md text-title-md text-primary-fixed">中奖记录</h3>
          <button class="text-on-surface-variant transition-colors hover:text-primary-fixed" title="刷新" @click="loadState">
            <span class="material-symbols-outlined" aria-hidden="true">refresh</span>
          </button>
        </div>
        <div class="space-y-3">
          <p v-if="loadFailed" class="text-sm text-error">记录加载失败</p>
          <p v-else-if="!state?.records?.length" class="text-sm text-on-surface-variant">还没有抽奖记录</p>
          <div
            v-for="record in state?.records ?? []"
            :key="record.draw_no"
            class="rounded-lg border border-outline-variant bg-surface-container-lowest/70 p-3"
          >
            <div class="flex items-center justify-between gap-3">
              <span class="font-title-md text-body-md text-on-surface">{{ record.prize_name }}</span>
              <span class="text-xs text-primary-fixed">{{ deliveryStatusText(record.delivery_status) }}</span>
            </div>
            <p class="mt-1 text-xs text-on-surface-variant">{{ new Date(record.created_at).toLocaleString() }}</p>
          </div>
        </div>
      </section>
    </aside>

    <!-- 抽奖结果弹窗 -->
    <Teleport to="body">
      <div v-if="resultModal.open" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-5">
        <div class="w-full max-w-md rounded-xl border-2 border-primary-fixed bg-surface-container-highest p-7 text-center">
          <span
            class="material-symbols-outlined text-6xl"
            :class="resultModal.icon === 'error' ? 'text-error' : 'text-primary-fixed'"
            aria-hidden="true"
            >{{ resultModal.icon }}</span
          >
          <h3 class="mt-4 font-display-hero text-headline-lg-mobile font-extrabold text-primary-fixed">
            {{ resultModal.title }}
          </h3>
          <p class="mt-3 text-on-surface-variant">{{ resultModal.description }}</p>
          <button
            class="mt-6 rounded-lg bg-primary-fixed px-8 py-3 font-title-md text-title-md text-on-primary-fixed active:scale-95"
            @click="closeResult"
          >
            确认
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
/* 福签按钮质感（源自旧 mini_6） */
.talisman {
  background: linear-gradient(180deg, #ffe16d, #e9c400);
  box-shadow: 0 12px 40px rgba(255, 215, 0, 0.28);
}

.draw-glow {
  box-shadow:
    0 0 28px rgba(255, 225, 109, 0.26),
    inset 0 0 18px rgba(255, 255, 255, 0.12);
}
</style>
