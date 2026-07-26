/** internal/service/gacha_service.go */

export interface GachaStateView {
  pool_id: number
  pity_counter: number
  pity_limit: number
  total_draw: number
}

export interface GachaDrawPayload {
  pool_id: number
  /** 仅支持 1 或 10 */
  times: number
}

export interface DrawResult {
  /** 'hero' | 'gold'（RewardService 另支持 diamond/stamina） */
  item_type: string
  item_id: number
  item_count: number
  quality: number
  is_pity: boolean
  is_duplicate: boolean
  converted_shards: number
}

export interface DrawOutput {
  draw_no: string
  diamond: number
  pity_counter: number
  pity_limit: number
  total_draw: number
  results: DrawResult[]
}
