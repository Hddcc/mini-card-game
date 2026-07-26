/** 数值与后端约定常量。seed 来源见各项注释，改动后端配置时需同步。 */

/** 体力上限（internal/service/player_service.go: MaxStamina） */
export const MAX_STAMINA = 120

/** 体力恢复间隔秒数（internal/service/player_service.go: StaminaRecoverSeconds） */
export const STAMINA_RECOVER_SECONDS = 300

/** 默认卡池「天命召唤」（internal/model/db.go seed） */
export const GACHA_POOL_ID = 1
export const GACHA_COST_ONE = 160
export const GACHA_COST_TEN = 1600
/** 保底次数，仅作接口返回前的展示兜底 */
export const GACHA_PITY_LIMIT_FALLBACK = 90

/** 神将星级上限（internal/service/hero_progression.go: MaxHeroStar） */
export const MAX_HERO_STAR = 10

/** 阵容槽位数 */
export const TEAM_SLOT_COUNT = 5

/** 战斗日志最大保留条数（internal/service/battle_service.go: MaxBattleLogs） */
export const MAX_BATTLE_LOGS = 30

/** 经验条分母：旧大厅页的客户端约定公式（后端未下发升级所需经验） */
export function expMaxForLevel(level: number): number {
  return Math.max(1000, level * 1000)
}

/** 统一 API 请求超时（毫秒），沿用旧升星页的 8s AbortController 模式 */
export const API_TIMEOUT_MS = 8000
