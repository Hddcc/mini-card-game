/** internal/service/task_service.go —— GET /tasks/daily 的 data 为裸数组 */

/** 任务状态：0 未完成 / 1 已完成未领 / 2 已领取 */
export const TASK_STATUS = {
  inProgress: 0,
  claimable: 1,
  claimed: 2,
} as const

export interface TaskView {
  task_id: number
  name: string
  /** gacha_draw | stage_fight | stage_win */
  event_type: string
  progress: number
  target_count: number
  status: number
  reward_gold: number
  reward_diamond: number
  claimed_at: string | null
}

export interface TaskClaimPayload {
  task_id: number
}
