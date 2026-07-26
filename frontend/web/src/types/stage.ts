/** internal/service/stage_service.go —— GET /stages/progress 的 data 为裸数组，只含玩家已有记录 */

export interface StageProgressView {
  stage_id: number
  /** 1 = 已通关 */
  status: number
  best_power: number
  first_passed_at?: string
}
