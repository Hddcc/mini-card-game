import { http } from './http'
import type { BattleActionPayload, BattleResponse } from '@/types/battle'

/** 已有未过期 active 会话时，后端直接返回该会话（断线续战） */
export function startBattle(stageId: number): Promise<BattleResponse> {
  return http.post<BattleResponse>('/api/v1/stage/battle/start', { stage_id: stageId })
}

export function submitBattleAction(payload: BattleActionPayload): Promise<BattleResponse> {
  return http.post<BattleResponse>('/api/v1/stage/battle/action', payload)
}

export function surrenderBattle(sessionId: number): Promise<BattleResponse> {
  return http.post<BattleResponse>('/api/v1/stage/battle/surrender', { session_id: sessionId })
}
