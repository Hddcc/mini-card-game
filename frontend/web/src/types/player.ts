/**
 * internal/service/player_service.go
 * 注意：/player/profile 同时输出 snake_case 与 PascalCase 两套遗留字段，
 * 前端只建模 snake_case 一套。
 */

export interface ProfileView {
  unique_id: number
  player_id: number
  user_id: number
  name: string
  nickname: string
  level: number
  exp: number
  avatar: string
  power: number
  created_at: string
  updated_at: string
}

export interface AssetView {
  player_id: number
  gold: number
  diamond: number
  stamina: number
  max_stamina: number
  next_stamina_seconds: number
  stamina_recover_every: number
}

/**
 * /heroes/star-up 返回的 assets 内嵌的是无 json tag 的 model.PlayerAsset，
 * 序列化为 PascalCase —— 全项目唯一的字段风格例外，经 normalizeAssets 归一化。
 */
export interface RawStarUpAssets {
  PlayerID: number
  Gold: number
  Diamond: number
  Stamina: number
  StaminaUpdatedAt: string | null
  CreatedAt: string
  UpdatedAt: string
}

export interface UpdateNamePayload {
  name: string
}
