import { http } from './http'
import type { ActivityDrawOutput, ActivityStateView } from '@/types/activity'

export function fetchActivityState(): Promise<ActivityStateView> {
  return http.get<ActivityStateView>('/api/v1/activity/lottery/state')
}

/** 请求体为空对象（后端无参数）；发奖异步，抽完需重拉 state 与资产 */
export function drawActivityLottery(): Promise<ActivityDrawOutput> {
  return http.post<ActivityDrawOutput>('/api/v1/activity/lottery/draw', {})
}
