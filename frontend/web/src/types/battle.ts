/** internal/service/battle_service.go —— 三个接口（start/action/surrender）返回同一 BattleResponse */

export type BattleStatus = 'active' | 'won' | 'lost' | 'abandoned'
export type BattleActionType = 'normal_attack' | 'skill' | 'defend'
export type BattleTargetType = 'enemy' | 'ally' | 'ally_lowest' | 'self'

export interface BattleBuff {
  stat: string
  amount: number
  remaining_rounds: number
  source_skill_id: number
}

export interface BattleUnit {
  /** 我方 p<slot>，敌方 e<slot>_<index> */
  id: string
  side: 'player' | 'enemy'
  board_side: string
  slot: number
  board_slot: number
  name: string
  role: string
  source_id: number
  player_hero_id?: number
  config_id: number
  level: number
  star: number
  max_hp: number
  hp: number
  atk: number
  def: number
  rage: number
  skill_id: number
  skill_name: string
  skill_target: string
  skill_effect: string
  skill_multiplier: number
  skill_duration: number
  skill_stat_delta: number
  skill_cost_rage: number
  skill_cooldown: number
  skill_description: string
  skill_effect_key: string
  skill_animation: string
  card_art: string
  portrait_art: string
  animation_keys: {
    attack: string
    skill: string
    hit: string
    defeat: string
    idle: string
  }
  cooldown_left: number
  defending: boolean
  alive: boolean
  buffs?: BattleBuff[]
  failure_hint_tag?: string
}

export interface BattleLog {
  round: number
  actor_id: string
  actor_name: string
  action: string
  target_id?: string
  target_name?: string
  damage?: number
  heal?: number
  buff_stat?: string
  buff_amount?: number
  effect_key?: string
  animation_key?: string
  text: string
}

export interface BattleActionView {
  type: string
  label: string
  skill_id?: number
  enabled: boolean
  target_type?: string
  effect_type?: string
  effect_key?: string
  animation_key?: string
  valid_targets?: string[]
  description?: string
}

export interface BattleState {
  version: number
  round: number
  status: BattleStatus
  encounter_id?: number
  encounter_name?: string
  current_actor_id: string
  selectable_actors: string[]
  player_units: BattleUnit[]
  enemy_units: BattleUnit[]
  available_actions: BattleActionView[]
  selected_targets: string[]
  logs: BattleLog[]
  failure_hint?: string
}

export interface BattleResult {
  win: boolean
  reward_gold: number
  reward_exp: number
  stamina: number
  max_stamina: number
  next_stamina_seconds: number
  best_power: number
  failure_hint?: string
}

export interface BattleResponse {
  session_id: number
  stage_id: number
  status: string
  expires_at: string
  state: BattleState
  /** omitempty：仅结算时出现 */
  result?: BattleResult
}

export interface BattleStartPayload {
  stage_id: number
}

export interface BattleActionPayload {
  session_id: number
  action: BattleActionType
  actor_id: string
  target_id: string
  skill_id: number
}

export interface BattleSurrenderPayload {
  session_id: number
}

/** 客户端构建的行动项（旧 mini_2 在前端拼装行动列表，后端 available_actions 仅作参考） */
export interface ClientAction {
  type: BattleActionType
  label: string
  skillId: number
  enabled: boolean
  targetType: BattleTargetType
  effectType: string
  validTargets: string[]
  costRage: number
  cooldown: number
  cooldownLeft: number
  description: string
  animationKey?: string
}
