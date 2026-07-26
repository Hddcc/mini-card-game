<script setup lang="ts">
/**
 * 战场单位卡（迁移自旧 mini_2 unitCard）：
 * - 圆顶塔牌造型，左上 ATK / 右上 HP 圆标，中部头像，底部 HP/怒气双条 + DEF/存活标
 * - 我方存活且在 selectable_actors 中可选为行动者；当前行动的合法目标高亮可点
 * - 点击优先级与旧版一致：需点选目标时先当目标，其次选行动者
 */
import { computed, ref } from 'vue'

import { heroImage } from '@/constants/heroAssets'
import type { BattleUnit } from '@/types/battle'

const props = defineProps<{
  unit: BattleUnit
  side: 'player' | 'enemy'
  selectableActors: string[]
  validTargets: string[]
  pickingTarget: boolean
  selectedActorId: string
  selectedTargetId: string
}>()

const emit = defineEmits<{
  selectActor: [unitId: string]
  selectTarget: [unitId: string]
}>()

const imgFailed = ref(false)

const hpPercent = computed(() =>
  props.unit.max_hp ? Math.max(0, Math.round((props.unit.hp * 100) / props.unit.max_hp)) : 0,
)
const ragePercent = computed(() => Math.max(0, Math.min(100, props.unit.rage || 0)))

const selectableActor = computed(
  () => props.side === 'player' && props.unit.alive && props.selectableActors.includes(props.unit.id),
)
const selectableTarget = computed(() => props.unit.alive && props.validTargets.includes(props.unit.id))
const selected = computed(
  () => props.unit.id === props.selectedActorId || props.unit.id === props.selectedTargetId,
)
const selectable = computed(() => selectableActor.value || selectableTarget.value)

/** 我方优先用前端立绘映射，敌方用后端下发的 card_art（旧 mini_2 顺序） */
const artUrl = computed(() => {
  const heroArt = props.side === 'player' ? heroImage(props.unit.config_id) : undefined
  return heroArt || props.unit.card_art || ''
})
const portraitUrl = computed(() => {
  const heroArt = props.side === 'player' ? heroImage(props.unit.config_id) : undefined
  return heroArt || props.unit.portrait_art || props.unit.card_art || ''
})
const fallbackIcon = computed(() => (props.side === 'enemy' ? 'pets' : 'auto_awesome'))

const cardStyle = computed(() =>
  artUrl.value
    ? {
        backgroundImage: `linear-gradient(180deg,rgba(10,8,6,.18),rgba(10,8,6,.78)),url('${artUrl.value}')`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }
    : undefined,
)

function onClick(): void {
  if (selectableTarget.value && props.pickingTarget) {
    emit('selectTarget', props.unit.id)
  } else if (selectableActor.value) {
    emit('selectActor', props.unit.id)
  } else if (selectableTarget.value) {
    emit('selectTarget', props.unit.id)
  }
}
</script>

<template>
  <button
    type="button"
    class="relative h-48 w-32 overflow-hidden rounded-t-full rounded-b-lg border-2 bg-gradient-to-b p-2 transition-all md:h-52 md:w-36 xl:h-56 xl:w-40"
    :class="[
      side === 'enemy' ? 'from-[#4b1717] to-[#1f1111]' : 'from-[#17334b] to-[#101720]',
      selected ? 'scale-[1.03] border-primary-fixed shadow-[0_0_22px_rgba(255,215,0,0.75)]' : 'border-[#7a6a4a]',
      selectable ? 'hover:-translate-y-1 hover:border-primary-fixed' : 'opacity-70',
    ]"
    :style="cardStyle"
    @click="onClick"
  >
    <div class="pointer-events-none absolute inset-1 rounded-t-full rounded-b-md border border-white/10"></div>
    <div
      class="absolute left-2 top-2 flex h-8 w-8 items-center justify-center rounded-full border-2 border-[#5b503d] bg-[#d8d2c2] font-stats-num text-sm text-[#15110c]"
    >
      {{ unit.atk || 0 }}
    </div>
    <div
      class="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-full border-2 border-[#5b1a1a] bg-[#b72b2b] font-stats-num text-sm text-white"
    >
      {{ unit.hp }}
    </div>

    <div
      class="mx-auto mt-7 flex h-20 w-20 items-center justify-center overflow-hidden rounded-xl border-4 border-[#8f8165] bg-surface xl:h-24 xl:w-24"
    >
      <img
        v-if="portraitUrl && !imgFailed"
        :src="portraitUrl"
        :alt="unit.name || 'card art'"
        class="h-full w-full object-cover"
        @error="imgFailed = true"
      />
      <span v-else class="material-symbols-outlined text-3xl text-primary-fixed/80" aria-hidden="true">{{
        fallbackIcon
      }}</span>
    </div>

    <div class="mt-2 text-center">
      <p class="truncate font-title-md text-sm text-on-surface drop-shadow">{{ unit.name }}</p>
      <p class="font-label-sm text-[10px] text-on-surface-variant">{{ unit.role || 'unit' }} · Lv.{{ unit.level }}</p>
    </div>

    <div class="absolute bottom-9 left-3 right-3 h-2 overflow-hidden rounded bg-black/40">
      <div
        class="h-full bg-gradient-to-r from-tertiary-fixed to-[#4ADE80] transition-all"
        :style="{ width: `${hpPercent}%` }"
      ></div>
    </div>
    <div class="absolute bottom-6 left-3 right-3 h-1 overflow-hidden rounded bg-black/40">
      <div class="h-full bg-primary-fixed transition-all" :style="{ width: `${ragePercent}%` }"></div>
    </div>

    <div class="absolute bottom-1 left-2 right-2 flex justify-between font-label-sm text-[10px] text-on-surface-variant">
      <span>DEF {{ unit.def || 0 }}</span><span>{{ unit.alive ? '可战' : '击败' }}</span>
    </div>

    <!-- 防御中 / buff 标记 -->
    <div v-if="unit.defending || (unit.buffs && unit.buffs.length)" class="absolute left-2 top-11 flex flex-col gap-1">
      <span
        v-if="unit.defending"
        class="flex h-6 w-6 items-center justify-center rounded-full border border-tertiary-container/70 bg-black/60"
        title="防御中"
      >
        <span class="material-symbols-outlined text-sm text-tertiary-container" aria-hidden="true">shield</span>
      </span>
      <span
        v-for="(buff, index) in unit.buffs ?? []"
        :key="index"
        class="flex h-6 w-6 items-center justify-center rounded-full border border-primary-fixed/70 bg-black/60"
        :title="`${buff.stat === 'atk' ? '攻击' : '防御'}+${buff.amount}（剩 ${buff.remaining_rounds} 回合）`"
      >
        <span class="material-symbols-outlined text-sm text-primary-fixed" aria-hidden="true">{{
          buff.stat === 'atk' ? 'swords' : 'shield_person'
        }}</span>
      </span>
    </div>
  </button>
</template>
