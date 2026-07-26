/** internal/service/hero_service.go */
import type { RawStarUpAssets } from './player'

export interface StarUpCost {
  shard: number
  gold: number
}

export interface HeroView {
  player_hero_id: number
  /** 与 player_hero_id 同值（后端兼容字段） */
  id: number
  hero_config_id: number
  name: string
  quality: number
  role: string
  level: number
  star: number
  shard: number
  max_star: boolean
  /** omitempty：满星时缺省 */
  next_star_cost?: StarUpCost
}

/** GET /heroes 的 data 是 {heroes:[...]} 包裹（与其他裸数组端点不同） */
export interface HeroListData {
  heroes: HeroView[]
}

export interface StarUpPayload {
  player_hero_id: number
}

export interface StarUpOutput {
  hero: HeroView
  assets: RawStarUpAssets
}
