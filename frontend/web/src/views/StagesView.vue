<script setup lang="ts">
/**
 * 西行关卡（迁移自旧 mini_2）：
 * - 章节横幅 + 体力/队伍战力条
 * - 关卡卡三态：已通关（三星/再次挑战）、可挑战（金框脉冲/挑战）、未解锁（锁定遮罩）
 *   解锁规则照抄：cleared || id===1 || (maxCleared>0 && id<=maxCleared+1)
 * - 开战/续战/行动/投降/结算/下一关续链；战斗中锁导航（路由守卫）
 */
import { computed, onMounted, ref, watch } from 'vue'

import { fetchStageProgress } from '@/api/stage'
import BattlePanel from '@/components/battle/BattlePanel.vue'
import BattleResultModal from '@/components/battle/BattleResultModal.vue'
import JadeProgressBar from '@/components/ui/JadeProgressBar.vue'
import { useToast } from '@/composables/useToast'
import { IMAGE_HOME_BACKGROUND, IMAGE_REWARD_GOLD } from '@/constants/heroAssets'
import { PLACEHOLDER_STAGE, STAGES, nextStageId, type StageMeta } from '@/constants/stages'
import { useAssetStore } from '@/stores/assets'
import { useBattleStore } from '@/stores/battle'
import { usePlayerStore } from '@/stores/player'
import { errorText } from '@/types/api'

const toast = useToast()
const battleStore = useBattleStore()
const assetStore = useAssetStore()
const playerStore = usePlayerStore()

const clearedStageIds = ref<Set<number>>(new Set())
const resultModalOpen = ref(false)

interface StageCardState {
  meta: StageMeta
  cleared: boolean
  challengeable: boolean
  current: boolean
}

const stageCards = computed<StageCardState[]>(() => {
  const cleared = clearedStageIds.value
  const maxCleared = Math.max(0, ...cleared)
  return STAGES.map((meta) => {
    const isCleared = cleared.has(meta.id)
    const challengeable = isCleared || meta.id === 1 || (maxCleared > 0 && meta.id <= maxCleared + 1)
    return {
      meta,
      cleared: isCleared,
      challengeable,
      current: !isCleared && challengeable,
    }
  })
})

const nextStage = computed(() => {
  const stageId = battleStore.battle?.stage_id
  return stageId ? nextStageId(stageId) : null
})

async function loadStageProgress(): Promise<void> {
  try {
    const progress = await fetchStageProgress()
    const next = new Set<number>()
    for (const item of progress) {
      if (item.stage_id && Number(item.status) === 1) next.add(item.stage_id)
    }
    clearedStageIds.value = next
  } catch {
    // 进度加载失败保持现状（旧版行为：静默回退）
  }
}

onMounted(() => {
  void loadStageProgress()
  void playerStore.refresh().catch(() => undefined)
})

async function startStage(stageId: number): Promise<void> {
  if (battleStore.isActive) {
    toast.show('战斗进行中，请先行动或投降')
    return
  }
  try {
    const response = await battleStore.start(stageId)
    if (response.stage_id !== stageId) {
      toast.show('已恢复进行中的战斗')
    }
    void assetStore.refresh().catch(() => undefined)
  } catch (error) {
    toast.show(`开战失败：${errorText(error, '未知错误')}`)
  }
}

/** 结算出现时弹结果窗；胜利时标记通关并重拉进度（旧 showBattleResult 行为） */
watch(
  () => battleStore.battle?.result,
  (result) => {
    if (!result) return
    resultModalOpen.value = true
    if (result.win && battleStore.battle) {
      clearedStageIds.value = new Set([...clearedStageIds.value, battleStore.battle.stage_id])
      void loadStageProgress()
    }
    void assetStore.refresh().catch(() => undefined)
  },
)

function backToPreparation(): void {
  resultModalOpen.value = false
  battleStore.reset()
  void assetStore.refresh().catch(() => undefined)
}

async function startNextStage(): Promise<void> {
  const next = nextStage.value
  if (!next) {
    toast.show('当前章节暂无下一关')
    return
  }
  resultModalOpen.value = false
  battleStore.reset()
  await startStage(next)
}
</script>

<template>
  <div class="mx-auto max-w-7xl space-y-stack-lg">
    <!-- 章节横幅 -->
    <section class="relative h-56 w-full overflow-hidden md:h-72">
      <img
        :src="IMAGE_HOME_BACKGROUND"
        alt=""
        class="absolute inset-0 h-full w-full object-cover object-center opacity-70"
      />
      <div
        class="absolute inset-0 bg-[radial-gradient(circle_at_28%_24%,rgba(255,225,109,.18),transparent_24%),radial-gradient(circle_at_78%_62%,rgba(111,242,174,.1),transparent_24%),linear-gradient(135deg,rgba(45,45,45,.3)_0%,rgba(19,19,19,.72)_62%,rgba(14,14,14,.88)_100%)]"
      ></div>
      <div class="absolute inset-0 bg-gradient-to-t from-background via-background/40 to-transparent"></div>
      <div class="absolute bottom-stack-lg left-gutter md:left-stack-lg">
        <span
          class="mb-2 inline-block rounded-full border border-quality-ssr bg-quality-ssr/20 px-3 py-1 font-label-sm text-label-sm uppercase text-quality-ssr"
          >第一章</span
        >
        <h2 class="font-display-hero text-headline-lg-mobile text-primary-fixed drop-shadow-lg md:text-headline-lg">
          五行山
        </h2>
        <p class="mt-2 max-w-lg font-body-md text-body-md text-on-surface-variant">
          西行之路穿过古老封印，击败山中守卫继续前进。
        </p>
      </div>
    </section>

    <!-- 状态条：体力 + 队伍战力 -->
    <section class="relative flex items-center justify-between overflow-hidden border border-outline-variant bg-ink-wash p-stack-md">
      <div class="flex items-center gap-gutter">
        <div class="flex flex-col">
          <span class="font-label-sm text-label-sm uppercase tracking-widest text-on-surface-variant">体力</span>
          <div class="mt-1 flex items-center gap-2">
            <div class="w-32">
              <JadeProgressBar :value="assetStore.stamina" :max="assetStore.maxStamina" height-class="h-3" />
            </div>
            <span class="font-stats-num text-label-sm text-primary-fixed"
              >{{ assetStore.stamina }}/{{ assetStore.maxStamina }}</span
            >
          </div>
        </div>
        <div class="h-10 w-px bg-outline-variant"></div>
        <div class="flex flex-col">
          <span class="font-label-sm text-label-sm uppercase tracking-widest text-on-surface-variant">队伍战力</span>
          <span class="font-stats-num text-stats-num text-quality-ur">{{
            playerStore.power.toLocaleString('zh-CN')
          }}</span>
        </div>
      </div>
    </section>

    <!-- 关卡卡片 -->
    <section class="grid grid-cols-1 gap-stack-lg pb-8 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="card in stageCards"
        :key="card.meta.id"
        class="relative p-stack-md transition-all"
        :class="
          card.challengeable
            ? card.current
              ? 'group cursor-pointer border-2 border-primary-fixed bg-ink-wash shadow-[0_0_10px_rgba(255,215,0,0.2)] hover:scale-[1.02] active:scale-95'
              : 'group cursor-pointer border-2 border-outline-variant bg-ink-wash hover:scale-[1.02] active:scale-95'
            : 'border-2 border-outline-variant/30 bg-surface-container-low opacity-60'
        "
      >
        <div class="meander-corner meander-tl"></div>
        <div class="meander-corner meander-tr"></div>
        <div class="meander-corner meander-bl"></div>
        <div class="meander-corner meander-br"></div>

        <!-- 锁定遮罩 -->
        <div v-if="!card.challengeable" class="absolute inset-0 z-10 flex items-center justify-center">
          <div class="rounded-full border border-outline-variant bg-black/60 p-4">
            <span class="material-symbols-outlined text-4xl text-on-surface-variant" aria-hidden="true">lock</span>
          </div>
        </div>

        <!-- 角标：三星 / 可挑战 -->
        <div class="absolute right-2 top-2 z-10 flex gap-1">
          <template v-if="card.cleared">
            <span
              v-for="star in 3"
              :key="star"
              class="material-symbols-outlined icon-filled text-quality-ssr"
              aria-hidden="true"
              >star</span
            >
          </template>
          <span
            v-else-if="card.current"
            class="animate-pulse rounded bg-quality-ur px-2 py-0.5 font-label-sm text-[10px] text-white"
            >可挑战</span
          >
        </div>

        <div class="mb-stack-md">
          <p
            class="font-label-sm text-label-sm"
            :class="card.challengeable ? 'text-primary-fixed' : 'text-on-surface-variant'"
          >
            关卡 {{ card.meta.chapter }}-{{ card.meta.id }}
          </p>
          <h3
            class="font-display-hero text-title-md"
            :class="card.challengeable ? 'text-on-surface' : 'text-on-surface-variant'"
          >
            {{ card.meta.name }}
          </h3>
        </div>

        <div class="mb-stack-lg space-y-2">
          <div class="flex justify-between font-label-sm text-label-sm">
            <span class="text-on-surface-variant">推荐战力</span>
            <span :class="card.challengeable ? 'text-on-surface' : 'text-on-surface-variant'">{{
              card.meta.recommendPower.toLocaleString('zh-CN')
            }}</span>
          </div>
          <div class="flex justify-between font-label-sm text-label-sm">
            <span class="text-on-surface-variant">体力消耗</span>
            <span :class="card.challengeable ? 'text-primary-fixed' : 'text-on-surface-variant'">{{
              card.meta.staminaCost
            }}</span>
          </div>
          <div class="flex justify-between font-label-sm text-label-sm">
            <span class="text-on-surface-variant">通关奖励</span>
            <span class="text-on-surface-variant"
              >金币 {{ card.meta.rewardGold }} · 经验 {{ card.meta.rewardExp }}</span
            >
          </div>
        </div>

        <button
          v-if="!card.challengeable"
          disabled
          class="w-full cursor-not-allowed border border-outline-variant/30 bg-surface-container-high py-3 font-label-sm text-label-sm text-on-surface-variant"
        >
          未解锁
        </button>
        <button
          v-else-if="card.current"
          class="relative w-full overflow-hidden bg-primary-fixed py-3 font-label-sm text-label-sm text-on-primary-fixed transition-all hover:brightness-110 active:scale-[0.98] group-hover:shadow-[0_0_15px_rgba(255,215,0,0.4)]"
          @click="startStage(card.meta.id)"
        >
          <span class="relative z-10 font-bold">挑战</span>
          <div
            class="absolute inset-0 translate-x-[-100%] bg-white/20 transition-transform duration-500 group-hover:translate-x-[100%]"
          ></div>
        </button>
        <button
          v-else
          class="w-full border border-outline bg-surface-variant py-3 font-label-sm text-label-sm text-on-surface transition-colors hover:bg-primary-fixed hover:text-on-primary-fixed"
          @click="startStage(card.meta.id)"
        >
          再次挑战
        </button>
      </div>

      <!-- 通关奖励 Bento -->
      <div
        class="flex flex-col gap-gutter border-2 border-outline-variant bg-gradient-to-br from-ink-wash to-surface-container-lowest p-stack-lg md:flex-row lg:col-span-2"
      >
        <div class="flex-1">
          <h4 class="mb-2 font-display-hero text-title-md text-primary-fixed">通关奖励</h4>
          <p class="mb-4 font-body-md text-sm text-on-surface-variant">
            通关本章全部关卡，解锁称号与丰厚补给。
          </p>
          <div class="flex gap-stack-md">
            <div
              class="relative flex h-16 w-16 items-center justify-center border border-outline-variant bg-surface-container-high"
            >
              <img alt="金币奖励" class="h-12 w-12 object-contain opacity-90" :src="IMAGE_REWARD_GOLD" />
              <span class="absolute bottom-0 right-1 font-label-sm text-[10px] text-primary-fixed">x50</span>
            </div>
            <div
              class="relative flex h-16 w-16 items-center justify-center border border-outline-variant bg-surface-container-high"
            >
              <span class="material-symbols-outlined icon-filled text-3xl text-quality-sr" aria-hidden="true"
                >diamond</span
              >
              <span class="absolute bottom-0 right-1 font-label-sm text-[10px] text-primary-fixed">x1</span>
            </div>
          </div>
        </div>
        <div class="h-auto bg-outline-variant md:w-px"></div>
        <div class="flex flex-1 flex-col justify-center">
          <div class="mb-2 flex items-center gap-stack-md">
            <span class="material-symbols-outlined text-quality-ssr" aria-hidden="true">auto_awesome</span>
            <span class="font-label-sm text-label-sm uppercase tracking-widest text-on-surface">全局加成</span>
          </div>
          <p class="font-body-md text-sm italic text-on-surface-variant">"神佑：第一章战斗中火焰伤害 +10%。"</p>
        </div>
      </div>

      <!-- 展示占位关卡（永久锁定，旧 mini_2 的第 4 张卡） -->
      <div class="relative border-2 border-outline-variant/30 bg-surface-container-low p-stack-md opacity-60">
        <div class="meander-corner meander-tl"></div>
        <div class="meander-corner meander-tr"></div>
        <div class="meander-corner meander-bl"></div>
        <div class="meander-corner meander-br"></div>
        <div class="absolute inset-0 z-10 flex items-center justify-center">
          <div class="rounded-full border border-outline-variant bg-black/60 p-4">
            <span class="material-symbols-outlined text-4xl text-on-surface-variant" aria-hidden="true">lock</span>
          </div>
        </div>
        <div class="mb-stack-md">
          <p class="font-label-sm text-label-sm text-on-surface-variant">关卡 {{ PLACEHOLDER_STAGE.chapter }}-4</p>
          <h3 class="font-display-hero text-title-md text-on-surface-variant">{{ PLACEHOLDER_STAGE.name }}</h3>
        </div>
        <div class="mb-stack-lg space-y-2">
          <div class="flex justify-between font-label-sm text-label-sm">
            <span class="text-on-surface-variant">推荐战力</span>
            <span class="text-on-surface-variant">??,???</span>
          </div>
          <div class="flex justify-between font-label-sm text-label-sm">
            <span class="text-on-surface-variant">体力消耗</span>
            <span class="text-on-surface-variant">18</span>
          </div>
        </div>
        <button
          disabled
          class="w-full cursor-not-allowed border border-outline-variant/30 bg-surface-container-high py-3 font-label-sm text-label-sm text-on-surface-variant"
        >
          未解锁
        </button>
      </div>
    </section>

    <!-- 战斗面板与结算 -->
    <BattlePanel @back-to-preparation="backToPreparation" />
    <BattleResultModal
      :open="resultModalOpen"
      :result="battleStore.battle?.result ?? null"
      :has-next-stage="nextStage !== null"
      @close="resultModalOpen = false"
      @back-to-preparation="backToPreparation"
      @next-stage="startNextStage"
    />
  </div>
</template>
