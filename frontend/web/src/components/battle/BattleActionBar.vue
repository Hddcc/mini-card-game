<script setup lang="ts">
/** 行动按钮区（迁移自旧 mini_2 battleActionButtons + actionTooltip）。 */
import type { ClientAction } from '@/types/battle'

const props = defineProps<{
  actions: ClientAction[]
  selectedType: string
  selectedSkillId: number
  pending: boolean
}>()

const emit = defineEmits<{
  select: [action: ClientAction]
}>()

function isActive(action: ClientAction): boolean {
  return action.type === props.selectedType && action.skillId === props.selectedSkillId
}

function targetLabel(action: ClientAction): string {
  if (action.targetType === 'enemy') return '敌方单体'
  if (action.targetType === 'ally') return '我方单体'
  if (action.targetType === 'ally_lowest') return '我方最低血量'
  if (action.targetType === 'self') return '自身'
  return '目标'
}

function effectLabel(effectType: string): string {
  const labels: Record<string, string> = {
    damage: '伤害',
    heal: '治疗',
    attack_buff: '攻击提升',
    defense_buff: '防御提升',
  }
  return labels[effectType] || effectType || '效果'
}

function cooldownText(action: ClientAction): string {
  return action.cooldownLeft ? `${action.cooldownLeft} 回合剩余` : `${action.cooldown || 0} 回合`
}

function tooltipTitle(action: ClientAction): string {
  return [
    action.label,
    `目标：${targetLabel(action)}`,
    `效果：${effectLabel(action.effectType)}`,
    `怒气：${action.costRage || 0}`,
    `冷却：${cooldownText(action)}`,
    action.description || '',
  ]
    .filter(Boolean)
    .join('\n')
}

function onSelect(action: ClientAction): void {
  if (!action.enabled || props.pending) return
  emit('select', action)
}
</script>

<template>
  <div class="grid grid-cols-3 gap-2">
    <div v-for="action in actions" :key="`${action.type}-${action.skillId}`" class="group relative">
      <button
        type="button"
        class="w-full px-2 py-3 font-label-sm text-label-sm transition-colors"
        :class="
          !action.enabled || pending
            ? 'cursor-not-allowed bg-surface-container-high text-on-surface-variant'
            : isActive(action)
              ? 'bg-primary-fixed text-on-primary-fixed shadow-[0_0_12px_rgba(255,215,0,0.55)]'
              : 'bg-surface-container-high text-on-surface hover:bg-primary-fixed hover:text-on-primary-fixed'
        "
        :disabled="!action.enabled || pending"
        :title="tooltipTitle(action)"
        @click="onSelect(action)"
      >
        {{ action.label }}
      </button>

      <!-- hover 详情卡 -->
      <div
        class="pointer-events-none absolute bottom-full left-1/2 z-30 mb-2 w-64 -translate-x-1/2 rounded-xl border border-outline-variant bg-[#1b1711]/95 p-3 text-left opacity-0 shadow-2xl transition-opacity group-hover:opacity-100"
      >
        <p class="font-title-md text-sm text-primary-fixed">{{ action.label }}</p>
        <p class="mt-1 font-label-sm text-label-sm text-on-surface-variant">
          目标：{{ targetLabel(action) }} · 效果：{{ effectLabel(action.effectType) }}
        </p>
        <p class="mt-1 font-label-sm text-label-sm text-on-surface-variant">
          怒气 {{ action.costRage || 0 }} · 冷却 {{ cooldownText(action) }}
        </p>
        <p class="mt-2 font-body-md text-xs text-on-surface">{{ action.description || '释放该行动。' }}</p>
      </div>
    </div>
  </div>
</template>
