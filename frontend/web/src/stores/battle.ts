/**
 * 回合制战斗状态机，逐行迁移自旧 mini_2 的战斗代码：
 * - 行动列表在客户端构建（普攻/防御/技能），技能受怒气与冷却门控
 * - target_type 为 enemy/ally 时需点选目标后提交；self/ally_lowest 自动选定立即提交
 * - ally_lowest 按 HP 升序取首位（并列取先者）
 * - pending 防重入；battle 整包替换后选择态重置为首个可行动单位
 * - isLocked 供路由守卫拦截战斗中离开（替代旧 iframe postMessage 锁）
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import {
  startBattle as apiStartBattle,
  submitBattleAction as apiSubmitBattleAction,
  surrenderBattle as apiSurrenderBattle,
} from '@/api/battle'
import type {
  BattleActionType,
  BattleResponse,
  BattleState,
  BattleTargetType,
  BattleUnit,
  ClientAction,
} from '@/types/battle'

function aliveIds(units: BattleUnit[] | undefined): string[] {
  return (units ?? []).filter((unit) => unit.alive).map((unit) => unit.id)
}

function targetsForType(targetType: BattleTargetType, state: BattleState, actorId: string): string[] {
  if (targetType === 'enemy') return aliveIds(state.enemy_units)
  if (targetType === 'ally') return aliveIds(state.player_units)
  if (targetType === 'ally_lowest') {
    const allies = (state.player_units ?? [])
      .filter((unit) => unit.alive)
      .slice()
      .sort((a, b) => (a.hp || 0) - (b.hp || 0))
    return allies[0] ? [allies[0].id] : []
  }
  if (targetType === 'self') return [actorId]
  return []
}

/** enemy/ally 需要玩家点选目标，self/ally_lowest 自动选定 */
export function targetPickRequired(action: ClientAction | null): boolean {
  return Boolean(action && (action.targetType === 'enemy' || action.targetType === 'ally'))
}

function buildActions(state: BattleState, selectedActorId: string): ClientAction[] {
  const actor =
    state.player_units?.find((unit) => unit.id === selectedActorId) ??
    state.player_units?.find((unit) => unit.id === state.current_actor_id)
  if (!actor) return []

  const enemyTargets = aliveIds(state.enemy_units)
  const actions: ClientAction[] = [
    {
      type: 'normal_attack',
      label: '普攻',
      skillId: 0,
      enabled: enemyTargets.length > 0,
      targetType: 'enemy',
      effectType: 'damage',
      validTargets: enemyTargets,
      costRage: 0,
      cooldown: 0,
      cooldownLeft: 0,
      description: '选择一个敌方目标，造成基于攻击力的普通伤害。',
      animationKey: actor.animation_keys?.attack,
    },
    {
      type: 'defend',
      label: '防御',
      skillId: 0,
      enabled: true,
      targetType: 'self',
      effectType: 'defense_buff',
      validTargets: [actor.id],
      costRage: 0,
      cooldown: 0,
      cooldownLeft: 0,
      description: '本回合采取守势，降低受到的伤害。',
      animationKey: actor.animation_keys?.skill,
    },
    {
      type: 'skill',
      label: actor.skill_name || '技能',
      skillId: actor.skill_id || 0,
      enabled:
        Boolean(actor.skill_id) &&
        (actor.rage || 0) >= (actor.skill_cost_rage || 0) &&
        (actor.cooldown_left || 0) === 0,
      targetType: (actor.skill_target || 'enemy') as BattleTargetType,
      effectType: actor.skill_effect || 'damage',
      validTargets: targetsForType((actor.skill_target || 'enemy') as BattleTargetType, state, actor.id),
      costRage: actor.skill_cost_rage || 0,
      cooldown: actor.skill_cooldown || 0,
      cooldownLeft: actor.cooldown_left || 0,
      description: actor.skill_description || '释放该卡牌的专属技能。',
      animationKey: actor.skill_animation,
    },
  ]
  return actions
}

export type TargetSelectResult = 'no-action' | 'invalid' | 'selected' | 'submitted'

export const useBattleStore = defineStore('battle', () => {
  const battle = ref<BattleResponse | null>(null)
  const selectedActorId = ref('')
  const selectedAction = ref<'' | BattleActionType>('')
  const selectedSkillId = ref(0)
  const selectedTargetId = ref('')
  const pending = ref(false)

  const state = computed<BattleState | null>(() => battle.value?.state ?? null)

  /** 战斗进行中（未结算）。旧 isActiveBattle() 等价实现 */
  const isActive = computed(
    () => Boolean(battle.value) && !battle.value?.result && battle.value?.state?.status === 'active',
  )
  /** 路由守卫用：战斗中禁止离开关卡页 */
  const isLocked = computed(() => isActive.value)

  const selectedActor = computed<BattleUnit | null>(
    () => state.value?.player_units?.find((unit) => unit.id === selectedActorId.value) ?? null,
  )

  const clientActions = computed<ClientAction[]>(() =>
    state.value ? buildActions(state.value, selectedActorId.value) : [],
  )

  const currentAction = computed<ClientAction | null>(() => {
    if (!selectedAction.value) return null
    return (
      clientActions.value.find(
        (action) => action.type === selectedAction.value && action.skillId === selectedSkillId.value,
      ) ?? null
    )
  })

  /** 当前行动的可选目标（用于卡牌高亮） */
  const validTargetIds = computed<string[]>(() => currentAction.value?.validTargets ?? [])
  const pickingTarget = computed(() => targetPickRequired(currentAction.value))

  function firstSelectableActorId(current: BattleState | null): string {
    return current?.selectable_actors?.[0] ?? ''
  }

  function resetSelection(): void {
    selectedActorId.value = firstSelectableActorId(state.value)
    selectedAction.value = ''
    selectedSkillId.value = 0
    selectedTargetId.value = ''
  }

  function adoptBattle(response: BattleResponse): void {
    battle.value = response
    resetSelection()
  }

  /** 开战；已有 active 会话时后端直接返回原会话（断线续战） */
  async function start(stageId: number): Promise<BattleResponse> {
    const response = await apiStartBattle(stageId)
    adoptBattle(response)
    return response
  }

  function selectActor(actorId: string): void {
    selectedActorId.value = actorId
    selectedAction.value = ''
    selectedSkillId.value = 0
    selectedTargetId.value = ''
  }

  /** 选择行动：需点选目标的返回 'pick' 等待点卡，否则自动锁定目标并立即提交 */
  async function selectAction(type: BattleActionType, skillId = 0): Promise<'pick' | 'submitted'> {
    selectedAction.value = type
    selectedSkillId.value = skillId
    const action = currentAction.value
    if (targetPickRequired(action)) {
      selectedTargetId.value = ''
      return 'pick'
    }
    selectedTargetId.value = action?.validTargets[0] ?? ''
    await submit()
    return 'submitted'
  }

  /** 点选目标卡：命中合法目标后（enemy/ally 场景）自动提交 */
  async function selectTarget(targetId: string): Promise<TargetSelectResult> {
    const action = currentAction.value
    if (!action) {
      selectedTargetId.value = ''
      return 'no-action'
    }
    if (!action.validTargets.includes(targetId)) {
      selectedTargetId.value = ''
      return 'invalid'
    }
    selectedTargetId.value = targetId
    if (targetPickRequired(action)) {
      await submit()
      return 'submitted'
    }
    return 'selected'
  }

  async function submit(): Promise<void> {
    if (!battle.value || pending.value) return
    if (!selectedAction.value) return
    pending.value = true
    try {
      const response = await apiSubmitBattleAction({
        session_id: battle.value.session_id,
        action: selectedAction.value,
        actor_id: selectedActorId.value,
        target_id: selectedTargetId.value,
        skill_id: selectedSkillId.value,
      })
      adoptBattle(response)
    } finally {
      pending.value = false
    }
  }

  async function surrender(): Promise<void> {
    if (!battle.value || !isActive.value || pending.value) return
    pending.value = true
    try {
      const response = await apiSurrenderBattle(battle.value.session_id)
      adoptBattle(response)
    } finally {
      pending.value = false
    }
  }

  function reset(): void {
    battle.value = null
    selectedActorId.value = ''
    selectedAction.value = ''
    selectedSkillId.value = 0
    selectedTargetId.value = ''
    pending.value = false
  }

  return {
    battle,
    state,
    pending,
    selectedActorId,
    selectedAction,
    selectedSkillId,
    selectedTargetId,
    isActive,
    isLocked,
    selectedActor,
    clientActions,
    currentAction,
    validTargetIds,
    pickingTarget,
    start,
    selectActor,
    selectAction,
    selectTarget,
    surrender,
    reset,
  }
})
