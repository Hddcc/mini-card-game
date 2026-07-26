<script setup lang="ts">
/**
 * 全屏战斗面板（迁移自旧 mini_2 battlePanel）：
 * 木纹牌桌布景，上敌方战场 / 中提示框 / 下我方手牌，底部日志 + 行动区。
 * 顶部按钮：战斗中为"投降"（需 confirm），结算后为"返回备战"。
 */
import { computed } from 'vue'

import BattleActionBar from './BattleActionBar.vue'
import BattleLogPanel from './BattleLogPanel.vue'
import BattleUnitCard from './BattleUnitCard.vue'
import { useToast } from '@/composables/useToast'
import { useBattleStore } from '@/stores/battle'
import { errorText } from '@/types/api'
import type { ClientAction } from '@/types/battle'

const emit = defineEmits<{
  backToPreparation: []
}>()

const toast = useToast()
const battleStore = useBattleStore()

const state = computed(() => battleStore.state)

const encounterTitle = computed(() =>
  state.value?.encounter_name ? `遭遇 · ${state.value.encounter_name}` : '回合战斗',
)
const roundTitle = computed(
  () => `Round ${state.value?.round || 1} · ${state.value?.status === 'active' ? '选择卡牌' : '结算'}`,
)

const actorText = computed(() => {
  const actor = battleStore.selectedActor
  return actor ? `已选：${actor.name}` : '请选择我方卡牌'
})

const targetText = computed(() => {
  const action = battleStore.currentAction
  if (!action) return '请选择普攻、防御或技能'
  const hint =
    action.targetType === 'enemy'
      ? '选择敌方卡牌'
      : action.targetType === 'ally' || action.targetType === 'ally_lowest'
        ? '选择我方卡牌'
        : action.targetType === 'self'
          ? '作用于自身'
          : '选择目标'
  return `行动：${action.label} · ${hint}`
})

async function onSelectAction(action: ClientAction): Promise<void> {
  try {
    await battleStore.selectAction(action.type, action.skillId)
  } catch (error) {
    toast.show(`行动失败：${errorText(error, '未知错误')}`)
  }
}

async function onSelectTarget(unitId: string): Promise<void> {
  try {
    const result = await battleStore.selectTarget(unitId)
    if (result === 'no-action') toast.show('请先选择普攻、防御或技能')
    else if (result === 'invalid') toast.show('请选择高亮目标')
  } catch (error) {
    toast.show(`行动失败：${errorText(error, '未知错误')}`)
  }
}

async function onTopAction(): Promise<void> {
  if (battleStore.isActive) {
    if (!window.confirm('确认投降？本次开战消耗的体力不会返还，但不会扣除其他资源。')) return
    try {
      await battleStore.surrender()
    } catch (error) {
      toast.show(`投降失败：${errorText(error, '未知错误')}`)
    }
    return
  }
  emit('backToPreparation')
}
</script>

<template>
  <!-- Teleport 到 body：跳出 MainLayout <main> 的 z-10 层叠上下文，
       否则侧栏(z-40)/顶栏(z-30)会压在战斗面板上（与 AppModal 同一模式） -->
  <Teleport to="body">
  <section
    v-if="battleStore.battle && state"
    class="fixed inset-0 z-[55] flex items-center justify-center overflow-hidden bg-black/90 p-2 backdrop-blur-sm"
  >
    <div
      class="relative flex h-[calc(100vh-1rem)] w-full max-w-[calc(100vw-1rem)] flex-col overflow-hidden border-2 border-primary-fixed bg-[#1b1410] shadow-[0_0_40px_rgba(0,0,0,0.65)]"
    >
      <div class="meander-corner meander-tl"></div>
      <div class="meander-corner meander-tr"></div>
      <div class="meander-corner meander-bl"></div>
      <div class="meander-corner meander-br"></div>

      <!-- 顶栏 -->
      <div
        class="flex shrink-0 items-center justify-between gap-stack-md border-b border-outline-variant bg-gradient-to-r from-[#3a2115] via-[#1f1712] to-[#3a2115] px-5 py-3"
      >
        <div>
          <p class="font-label-sm text-label-sm uppercase tracking-widest text-primary-fixed">{{ encounterTitle }}</p>
          <h2 class="font-display-hero text-title-md text-on-surface">{{ roundTitle }}</h2>
        </div>
        <button
          class="px-4 py-2 font-label-sm text-label-sm transition-colors"
          :class="
            battleStore.isActive
              ? 'border border-error/70 text-error hover:bg-error/10'
              : 'border border-outline-variant hover:bg-surface-variant'
          "
          :disabled="battleStore.pending"
          @click="onTopAction"
        >
          {{ battleStore.isActive ? '投降' : '返回备战' }}
        </button>
      </div>

      <!-- 牌桌 -->
      <div
        class="relative min-h-0 flex-1 overflow-hidden bg-[radial-gradient(circle_at_center,#e0c180_0,#a87942_48%,#49301f_100%)] px-4 py-4 md:px-8"
      >
        <div class="pointer-events-none absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-black/40 to-transparent"></div>
        <div class="pointer-events-none absolute inset-x-0 bottom-0 h-28 bg-gradient-to-t from-black/50 to-transparent"></div>
        <div
          class="pointer-events-none absolute bottom-12 left-4 top-12 hidden w-24 rounded-full border-l border-primary-fixed/20 bg-gradient-to-r from-black/25 to-transparent lg:block"
        ></div>
        <div
          class="pointer-events-none absolute bottom-12 right-4 top-12 hidden w-24 rounded-full border-r border-primary-fixed/20 bg-gradient-to-l from-black/25 to-transparent lg:block"
        ></div>

        <div
          class="relative z-10 grid h-full min-h-0 gap-4"
          style="grid-template-rows: minmax(190px, 1fr) minmax(120px, 0.72fr) minmax(230px, 1.1fr)"
        >
          <!-- 敌方战场 -->
          <div class="flex min-h-0 flex-col justify-start">
            <p class="mb-2 text-center font-label-sm text-label-sm tracking-widest text-secondary-fixed">敌方战场</p>
            <div class="flex flex-wrap items-start justify-center gap-3 overflow-hidden xl:gap-4">
              <BattleUnitCard
                v-for="unit in state.enemy_units"
                :key="unit.id"
                :unit="unit"
                side="enemy"
                :selectable-actors="state.selectable_actors ?? []"
                :valid-targets="battleStore.validTargetIds"
                :picking-target="battleStore.pickingTarget"
                :selected-actor-id="battleStore.selectedActorId"
                :selected-target-id="battleStore.selectedTargetId"
                @select-actor="battleStore.selectActor"
                @select-target="onSelectTarget"
              />
            </div>
          </div>

          <!-- 中部提示 -->
          <div class="flex min-h-[76px] items-center justify-center">
            <div
              class="rounded-xl border border-primary-fixed/30 bg-black/30 px-8 py-4 text-center shadow-[0_10px_30px_rgba(0,0,0,.25)] backdrop-blur-sm"
            >
              <p class="font-label-sm text-label-sm text-primary-fixed">{{ actorText }}</p>
              <p class="mt-1 font-body-md text-sm text-on-surface-variant">{{ targetText }}</p>
            </div>
          </div>

          <!-- 我方手牌 -->
          <div class="flex min-h-0 flex-col justify-end">
            <p class="mb-2 text-center font-label-sm text-label-sm tracking-widest text-tertiary-fixed">
              我方手牌 / 阵容
            </p>
            <div class="flex flex-wrap items-end justify-center gap-3 overflow-hidden xl:gap-4">
              <BattleUnitCard
                v-for="unit in state.player_units"
                :key="unit.id"
                :unit="unit"
                side="player"
                :selectable-actors="state.selectable_actors ?? []"
                :valid-targets="battleStore.validTargetIds"
                :picking-target="battleStore.pickingTarget"
                :selected-actor-id="battleStore.selectedActorId"
                :selected-target-id="battleStore.selectedTargetId"
                @select-actor="battleStore.selectActor"
                @select-target="onSelectTarget"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 底部：日志 + 行动 -->
      <div class="grid shrink-0 grid-cols-1 gap-3 border-t border-outline-variant bg-[#171210] p-3 lg:grid-cols-[1fr_380px]">
        <BattleLogPanel :logs="state.logs ?? []" />
        <div class="border border-outline-variant bg-surface-container p-3">
          <p class="mb-2 font-label-sm text-label-sm text-on-surface-variant">行动</p>
          <BattleActionBar
            :actions="battleStore.clientActions"
            :selected-type="battleStore.selectedAction"
            :selected-skill-id="battleStore.selectedSkillId"
            :pending="battleStore.pending"
            @select="onSelectAction"
          />
        </div>
      </div>
    </div>
  </section>
  </Teleport>
</template>
