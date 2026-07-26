import { http } from './http'
import type { StageProgressView } from '@/types/stage'

/** data 为裸数组，仅含玩家已有记录的关卡 */
export function fetchStageProgress(): Promise<StageProgressView[]> {
  return http.get<StageProgressView[]>('/api/v1/stages/progress')
}
