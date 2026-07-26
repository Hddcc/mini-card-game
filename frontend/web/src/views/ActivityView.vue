<script setup lang="ts">
/**
 * 限时活动·福签抽奖（重构自旧 mini_6 布局）：
 * - state 单包驱动：活动信息 / 截止时间 / 每日配额条 / 奖品余量 / 内嵌中奖记录 / eligible 门控
 * - 抽奖空请求体；失败原因走关键字规则映射为友好中文
 * - 发奖异步（delivery_status 0-3 文案），抽完与关弹窗时重拉 state 与资产
 *
 * 布局说明：放弃旧的「左主区 + 右侧栏(奖品池/记录)」三框结构，改为全宽纵向流：
 *   ① 电影感横幅（活动信息 + 抽签台合一）② 奖品池卡片网格 ③ 中奖记录多列面板。
 * 后端接口与业务逻辑保持不变，仅重做展现层。
 */
import { computed, onMounted, ref } from 'vue'

import { drawActivityLottery, fetchActivityState } from '@/api/activity'
import AppModal from '@/components/ui/AppModal.vue'
import JadePanel from '@/components/ui/JadePanel.vue'
import JadeProgressBar from '@/components/ui/JadeProgressBar.vue'
import StoneButton from '@/components/ui/StoneButton.vue'
import { deliveryStatusText, friendlyActivityReason } from '@/constants/messages'
import { IMAGE_ACTIVITY_BANNER, IMAGE_REWARD_GOLD } from '@/constants/heroAssets'
import { qualityColor } from '@/constants/quality'
import { useAssetStore } from '@/stores/assets'
import { errorRaw, errorText } from '@/types/api'
import type { ActivityPrizeView, ActivityStateView } from '@/types/activity'

const assetStore = useAssetStore()

const state = ref<ActivityStateView | null>(null)
const loadFailed = ref(false)
const drawing = ref(false)

const resultModal = ref<{ open: boolean; tone: 'win' | 'error'; icon: string; title: string; description: string }>({
  open: false,
  tone: 'win',
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
const dailyLimit = computed(() => Math.max(1, state.value?.quota?.daily_limit ?? 1))
const usedToday = computed(() => state.value?.quota?.used_today ?? 0)
const remainingToday = computed(() => state.value?.quota?.remaining_today ?? 0)
const eligibilityText = computed(() => {
  if (loadFailed.value) return '活动数据加载失败，请稍后重试'
  if (!state.value) return '加载活动资格中...'
  if (!state.value.active) return state.value.reason ? friendlyActivityReason(state.value.reason) : '暂无可参与活动'
  return state.value.eligible ? '可以抽取福签，中奖奖励会陆续到账。' : friendlyActivityReason(state.value.reason || '暂不可参与')
})
const drawDisabled = computed(() => drawing.value || !state.value?.eligible)

/** 奖品图标：金币用立绘，其余用 reward_type 图标映射，兜底 prize.icon / redeem。 */
function prizeIcon(prize: ActivityPrizeView): string {
  return ICON_MAP[prize.reward_type] || prize.icon || 'redeem'
}
/** 库存文案：不限量 / 已抽罄 / 余 N。 */
function stockLabel(prize: ActivityPrizeView): string {
  if (prize.unlimited) return '不限量'
  const left = Math.max(0, prize.left_num)
  return left <= 0 ? '已抽罄' : `余 ${left}`
}
/** 库存徽标配色：不限量→玉色，抽罄→朱红，有货→金色。 */
function stockTone(prize: ActivityPrizeView): string {
  if (prize.unlimited) return 'text-tertiary-container border-tertiary-container/40 bg-tertiary-container/10'
  return Math.max(0, prize.left_num) <= 0
    ? 'text-error border-error/40 bg-error/10'
    : 'text-primary-fixed border-primary-fixed/40 bg-primary-fixed/10'
}
/** 到账状态配色：已到账→玉色，领取中→金色，其余→中性。 */
function deliveryTone(status: number): string {
  if (status === 1) return 'text-tertiary-container'
  if (status === 0) return 'text-primary-fixed'
  return 'text-on-surface-variant'
}

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
      tone: 'win',
      icon: ICON_MAP[output.prize.reward_type] || output.prize.icon || 'redeem',
      title: output.prize.name,
      description: output.prize.description,
    }
    await loadState()
  } catch (error) {
    resultModal.value = {
      open: true,
      tone: 'error',
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
  <div class="mx-auto flex max-w-6xl flex-col gap-stack-lg">
    <!-- ① 电影感横幅：活动信息 + 抽签台合一 -->
    <JadePanel tone="gold" class="overflow-hidden">
      <div class="relative min-h-[380px]">
        <img
          :src="IMAGE_ACTIVITY_BANNER"
          alt="火焰山活动"
          class="pointer-events-none absolute inset-0 h-full w-full object-cover object-center opacity-40"
        />
        <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-black/85 via-ink-wash/55 to-black/80"></div>
        <div class="ink-overlay absolute inset-0"></div>

        <div class="relative z-10 grid items-center gap-stack-lg p-6 md:grid-cols-[1.35fr_1fr] md:p-10">
          <!-- 左：活动叙事 + 配额 -->
          <div class="flex flex-col gap-stack-md">
            <div class="flex flex-wrap items-center gap-stack-md">
              <span
                class="rounded-full border border-primary-fixed/60 bg-primary-fixed px-3 py-1 font-label-sm text-label-sm text-on-primary-fixed"
                >限时活动</span
              >
              <span class="font-label-sm text-label-sm text-on-surface-variant">{{ timeText || '活动时间加载中' }}</span>
            </div>

            <h2 class="glow-gold font-display-hero text-headline-lg leading-tight text-primary-fixed md:text-display-hero">
              {{ activityName }}
            </h2>
            <p class="max-w-xl font-body-md text-body-md text-on-surface-variant">{{ activityDesc }}</p>

            <!-- 今日配额条 -->
            <div class="mt-2 max-w-md rounded-lg border border-outline-variant bg-surface-container/70 p-4 backdrop-blur-sm">
              <div class="flex items-end justify-between">
                <span class="font-label-sm text-label-sm text-on-surface-variant">今日可抽次数</span>
                <span class="font-stats-num text-sm text-on-surface-variant"
                  >{{ usedToday }} / {{ dailyLimit }} 已用</span
                >
              </div>
              <div class="mt-2 flex items-center gap-stack-md">
                <span class="font-stats-num text-[40px] leading-none text-primary-fixed">{{ remainingToday }}</span>
                <div class="flex-1">
                  <JadeProgressBar :value="usedToday" :max="dailyLimit" variant="gold" height-class="h-2.5" />
                  <p class="mt-1 font-label-sm text-[10px] text-on-surface-variant">次可用 · 每日 0 点刷新</p>
                </div>
              </div>
            </div>

            <!-- 资格提示 -->
            <p class="flex items-center gap-2 font-body-md text-sm text-on-surface-variant">
              <span
                class="material-symbols-outlined text-base"
                :class="active && state?.eligible ? 'text-tertiary-container' : 'text-primary-fixed'"
                aria-hidden="true"
                >{{ active && state?.eligible ? 'verified' : 'info' }}</span
              >
              {{ eligibilityText }}
            </p>
          </div>

          <!-- 右：抽签台 -->
          <div class="flex flex-col items-center gap-4">
            <button
              class="talisman shimmer-sweep focus-ring flex h-72 w-60 flex-col items-center justify-center gap-4 rounded-xl border-2 border-white/40 text-on-primary-fixed transition-transform active:scale-95"
              :class="drawDisabled ? 'cursor-not-allowed opacity-50' : 'hover:scale-[1.02]'"
              :disabled="drawDisabled"
              @click="handleDraw"
            >
              <span class="material-symbols-outlined icon-filled text-6xl" aria-hidden="true">{{
                drawing ? 'hourglass_top' : 'redeem'
              }}</span>
              <span class="font-display-hero text-headline-lg-mobile font-extrabold">{{
                drawing ? '抽取中…' : '抽取福签'
              }}</span>
              <span class="font-label-sm text-label-sm opacity-80">每日限次 · 好运加持</span>
            </button>
            <p class="font-label-sm text-label-sm text-on-surface-variant">
              剩余 <span class="text-primary-fixed">{{ remainingToday }}</span> 次机会
            </p>
          </div>
        </div>
      </div>
    </JadePanel>

    <!-- ② 奖品池：全宽卡片网格 -->
    <section>
      <div class="mb-stack-md flex items-center justify-between">
        <h3 class="font-title-md text-title-md text-primary-fixed">奖品池</h3>
        <span class="font-label-sm text-label-sm text-on-surface-variant">动态库存</span>
      </div>

      <p v-if="loadFailed" class="rounded-lg border border-error/40 bg-error/5 p-4 text-sm text-error">
        奖品池加载失败，请稍后刷新
      </p>
      <p
        v-else-if="!state?.prizes?.length"
        class="rounded-lg border border-outline-variant bg-surface-container/60 p-4 text-sm text-on-surface-variant"
      >
        暂无奖品信息
      </p>

      <div v-else class="grid grid-cols-2 gap-gutter sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        <article
          v-for="prize in state.prizes"
          :key="prize.id"
          class="group relative flex flex-col gap-3 overflow-hidden rounded-xl border border-outline-variant bg-surface-container/80 p-4 transition-colors hover:border-primary-fixed/60"
        >
          <!-- 品质顶条 -->
          <span
            class="absolute inset-x-0 top-0 h-1"
            :style="{ backgroundColor: qualityColor(prize.quality) }"
            aria-hidden="true"
          ></span>

          <div class="flex items-start justify-between">
            <div
              class="flex h-12 w-12 items-center justify-center rounded-lg border border-outline-variant bg-surface-container-high text-primary-fixed"
            >
              <img
                v-if="prize.reward_type === 'gold'"
                :src="IMAGE_REWARD_GOLD"
                alt=""
                class="h-9 w-9 object-contain"
              />
              <span v-else class="material-symbols-outlined text-2xl" aria-hidden="true">{{ prizeIcon(prize) }}</span>
            </div>
            <span
              class="whitespace-nowrap rounded-full border px-2 py-0.5 font-label-sm text-[10px]"
              :class="stockTone(prize)"
              >{{ stockLabel(prize) }}</span
            >
          </div>

          <div class="min-w-0">
            <p class="truncate font-title-md text-body-md text-on-surface">{{ prize.name }}</p>
            <p class="mt-1 line-clamp-2 min-h-[2.5em] font-body-md text-xs text-on-surface-variant">
              {{ prize.description }}
            </p>
          </div>
        </article>
      </div>
    </section>

    <!-- ③ 中奖记录：全宽多列面板 -->
    <section class="rounded-xl border border-outline-variant bg-surface-container/80 p-5">
      <div class="mb-stack-md flex items-center justify-between">
        <h3 class="font-title-md text-title-md text-primary-fixed">中奖记录</h3>
        <button
          class="focus-ring flex items-center gap-1 rounded-md px-2 py-1 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:text-primary-fixed"
          title="刷新"
          @click="loadState"
        >
          <span class="material-symbols-outlined text-base" aria-hidden="true">refresh</span>
          刷新
        </button>
      </div>

      <p v-if="loadFailed" class="text-sm text-error">记录加载失败</p>
      <p v-else-if="!state?.records?.length" class="text-sm text-on-surface-variant">还没有抽奖记录，快去抽一签吧。</p>

      <div
        v-else
        class="thin-scrollbar grid max-h-[320px] grid-cols-1 gap-3 overflow-y-auto pr-1 sm:grid-cols-2 lg:grid-cols-3"
      >
        <div
          v-for="record in state.records"
          :key="record.draw_no"
          class="flex items-center justify-between gap-3 rounded-lg border border-outline-variant bg-surface-container-lowest/70 p-3"
        >
          <div class="min-w-0">
            <p class="truncate font-title-md text-body-md text-on-surface">{{ record.prize_name }}</p>
            <p class="mt-0.5 font-label-sm text-[11px] text-on-surface-variant">
              {{ new Date(record.created_at).toLocaleString() }}
            </p>
          </div>
          <span class="whitespace-nowrap font-label-sm text-label-sm" :class="deliveryTone(record.delivery_status)">{{
            deliveryStatusText(record.delivery_status)
          }}</span>
        </div>
      </div>
    </section>

    <!-- 抽奖结果弹窗（复用 AppModal：金边回纹面板） -->
    <AppModal :open="resultModal.open" max-width-class="max-w-md" @close="closeResult">
      <div class="p-7 text-center">
        <div
          class="mx-auto flex h-20 w-20 items-center justify-center rounded-full border-2"
          :class="
            resultModal.tone === 'error'
              ? 'border-error/60 bg-error/10 text-error'
              : 'divine-glow border-primary-fixed/60 bg-primary-fixed/10 text-primary-fixed'
          "
        >
          <span class="material-symbols-outlined icon-filled text-5xl" aria-hidden="true">{{ resultModal.icon }}</span>
        </div>

        <h3
          class="mt-5 font-display-hero text-headline-lg-mobile font-extrabold"
          :class="resultModal.tone === 'error' ? 'text-error' : 'shimmer-text'"
        >
          {{ resultModal.title }}
        </h3>
        <p class="mt-3 font-body-md text-body-md text-on-surface-variant">{{ resultModal.description }}</p>
        <p v-if="resultModal.tone === 'win'" class="mt-2 font-label-sm text-label-sm text-on-surface-variant">
          奖励将异步发放，可在「中奖记录」查看到账状态。
        </p>

        <StoneButton variant="primary" size="lg" block class="mt-6" @click="closeResult">确认</StoneButton>
      </div>
    </AppModal>
  </div>
</template>

<style scoped>
/* 福签按钮质感（沿用旧 mini_6 的金牌质地 + 辉光） */
.talisman {
  background: linear-gradient(180deg, #ffe16d, #e9c400);
  box-shadow:
    0 12px 40px rgba(255, 215, 0, 0.28),
    0 0 28px rgba(255, 225, 109, 0.26),
    inset 0 0 18px rgba(255, 255, 255, 0.12);
}
</style>
