/** internal/service/team_service.go —— GET /team 的 data 为裸数组 */

export interface TeamView {
  slot: number
  player_hero_id: number
  id: number
  hero_config_id: number
  name: string
  quality: number
  level: number
  star: number
}

export interface TeamSlotPayload {
  slot: number
  player_hero_id: number
}

export interface TeamSavePayload {
  slots: TeamSlotPayload[]
}
