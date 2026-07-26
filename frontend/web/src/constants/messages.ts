/**
 * 后端错误 message（英文串）→ 用户可读中文的集中映射。
 * 后端仅用通用错误码（40000 等），细分语义都在 message 文案里
 * （见 internal/handler、internal/service 各处返回值）。
 * 服务端已中文化的文案（登录、改名等）直接透传。
 */

const MESSAGE_MAP: Record<string, string> = {
  // 认证（internal/middleware/auth.go）
  'missing token': '登录已过期，请重新登录',
  'invalid token format': '登录已过期，请重新登录',
  'invalid token': '登录已过期，请重新登录',
  // 抽卡（internal/service/gacha_service.go）
  'diamond not enough': '灵玉不足',
  'not enough diamond': '灵玉不足',
  'gacha pool closed': '卡池未开放',
  'gacha pool item empty': '卡池奖励配置为空',
  'times must be 1 or 10': '抽卡次数仅支持单抽或十连',
  // 神将 / 升星（internal/service/hero_service.go）
  'hero not owned': '未拥有该神将',
  'hero already max star': '该神将已满星',
  'hero shard not enough': '碎片不足',
  'gold not enough': '金币不足',
  'hero star cost missing': '升星配置缺失',
  // 阵容（internal/service/team_service.go）
  'team size must be between 1 and 5': '阵容需要 1 到 5 名神将',
  'slot out of range': '阵位超出范围',
  'duplicate slot': '阵位重复',
  'invalid player_hero_id': '神将数据无效',
  'duplicate player_hero_id': '同一神将不能重复上阵',
  'one or more heroes do not belong to player': '包含未拥有的神将',
  // 关卡 / 战斗（internal/service/stage_service.go、battle_service.go）
  'stage not found': '关卡不存在',
  'previous stage is not cleared': '请先通关前置关卡',
  'please save your team first': '请先在英雄页保存阵容',
  'not enough stamina': '体力不足',
  'active battle already exists': '已有进行中的战斗',
  'battle session not found': '战斗会话不存在',
  'battle session finished': '战斗已结束',
  'battle session expired': '战斗会话已过期，请重新开战',
  'invalid battle action': '无效的战斗行动',
  'battle config missing': '战斗配置缺失',
  // 日常任务（internal/service/task_service.go）
  'task already claimed': '任务奖励已领取',
  'task not completed': '任务尚未完成',
  // 活动抽奖（internal/service/activity_lottery_service.go）
  'daily draw limit reached': '今日抽取次数已用完',
  'ip draw limit reached': '今日参与人数较多，请明日再来',
  'player blacklisted': '暂时无法参与本次活动',
  'ip blacklisted': '暂时无法参与本次活动',
  'no active activity': '活动暂未开放',
  'draw request is processing': '请求处理中，请稍候再试',
}

function containsCJK(text: string): boolean {
  return /[一-鿿]/.test(text)
}

/** 翻译后端 message：中文透传，英文查表（精确→包含），未命中原样返回。 */
export function translateMessage(raw: string): string {
  if (!raw) return '请求失败，请稍后重试'
  if (containsCJK(raw)) return raw
  const normalized = raw.trim().toLowerCase()
  const exact = MESSAGE_MAP[normalized]
  if (exact) return exact
  for (const [key, value] of Object.entries(MESSAGE_MAP)) {
    if (normalized.includes(key)) return value
  }
  return raw
}

/**
 * 活动页失败原因的关键字规则（沿用旧 mini_6 文案，命中即用，
 * 应用于 state.reason 与抽奖失败 message）。
 */
const ACTIVITY_REASON_RULES: Array<{ match: string[]; text: string }> = [
  { match: ['daily', 'limit'], text: '今日福签已抽完，明日再来。' },
  { match: ['ip', 'limit'], text: '今日参与人数较多，请明日再来。' },
  { match: ['inactive'], text: '活动暂未开放。' },
  { match: ['not', 'found'], text: '活动正在筹备中。' },
  { match: ['black'], text: '暂时无法参与本次活动。' },
]

export function friendlyActivityReason(raw: string): string {
  if (!raw) return '暂时无法参与，请稍后再试。'
  if (containsCJK(raw)) return raw
  const normalized = raw.toLowerCase()
  const rule = ACTIVITY_REASON_RULES.find((item) => item.match.every((key) => normalized.includes(key)))
  return rule ? rule.text : '暂时无法参与，请稍后再试。'
}

/** 活动奖励发放状态文案（异步发奖，0 待发放；沿用旧 mini_6） */
export const DELIVERY_STATUS_TEXT: Record<number, string> = {
  0: '领取中',
  1: '已到账',
  2: '稍后到账',
  3: '请稍后查看',
}

export function deliveryStatusText(status: number): string {
  return DELIVERY_STATUS_TEXT[status] ?? '请稍后查看'
}
