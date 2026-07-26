import { http } from './http'
import type { DrawOutput, GachaStateView } from '@/types/gacha'

export function fetchGachaState(poolId: number): Promise<GachaStateView> {
  return http.get<GachaStateView>(`/api/v1/gacha/state?pool_id=${poolId}`)
}

export function drawGacha(poolId: number, times: 1 | 10): Promise<DrawOutput> {
  return http.post<DrawOutput>('/api/v1/gacha/draw', { pool_id: poolId, times })
}
