import { http } from './http'
import type { TeamSavePayload, TeamView } from '@/types/team'

/** data 为裸数组 */
export function fetchTeam(): Promise<TeamView[]> {
  return http.get<TeamView[]>('/api/v1/team')
}

/** 成功时 data 为 {message:"team saved"}，前端只关心 code===0 */
export function saveTeam(payload: TeamSavePayload): Promise<unknown> {
  return http.post<unknown>('/api/v1/team/save', payload)
}
