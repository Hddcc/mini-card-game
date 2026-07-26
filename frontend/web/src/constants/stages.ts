/**
 * 关卡静态配置。后端无 /config/stages 接口，数值取自
 * internal/model/db.go 的 seedStageConfigs（改后端 seed 需同步此表）。
 */

export interface StageMeta {
  id: number
  name: string
  chapter: number
  staminaCost: number
  recommendPower: number
  rewardGold: number
  rewardExp: number
}

export const STAGES: StageMeta[] = [
  { id: 1, name: '花果山试炼', chapter: 1, staminaCost: 6, recommendPower: 500, rewardGold: 1000, rewardExp: 20 },
  { id: 2, name: '水帘洞守卫', chapter: 1, staminaCost: 6, recommendPower: 900, rewardGold: 1500, rewardExp: 30 },
  { id: 3, name: '东海龙宫', chapter: 1, staminaCost: 8, recommendPower: 1300, rewardGold: 2000, rewardExp: 40 },
]

/** 旧关卡页的第 4 张展示占位卡（无后端配置，永久锁定） */
export const PLACEHOLDER_STAGE = { name: '天门试炼', chapter: 1 }

export function stageById(id: number): StageMeta | undefined {
  return STAGES.find((stage) => stage.id === id)
}

export function nextStageId(currentId: number): number | null {
  const index = STAGES.findIndex((stage) => stage.id === currentId)
  if (index < 0 || index + 1 >= STAGES.length) return null
  return STAGES[index + 1]!.id
}
