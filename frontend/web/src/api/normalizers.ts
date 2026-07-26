/** 收拢后端字段风格不一致的归一化逻辑。 */
import { MAX_STAMINA, STAMINA_RECOVER_SECONDS } from '@/constants'
import type { AssetView, RawStarUpAssets } from '@/types/player'

function isRawStarUpAssets(value: RawStarUpAssets | AssetView): value is RawStarUpAssets {
  return typeof (value as RawStarUpAssets).Gold === 'number'
}

/**
 * /heroes/star-up 返回 PascalCase 资产（model.PlayerAsset 无 json tag），
 * 统一转为 AssetView。升星不消耗体力，体力倒计时字段沿用 prev。
 */
export function normalizeAssets(raw: RawStarUpAssets | AssetView, prev?: AssetView | null): AssetView {
  if (!isRawStarUpAssets(raw)) return raw
  return {
    player_id: raw.PlayerID,
    gold: raw.Gold,
    diamond: raw.Diamond,
    stamina: raw.Stamina,
    max_stamina: prev?.max_stamina ?? MAX_STAMINA,
    next_stamina_seconds: prev?.next_stamina_seconds ?? 0,
    stamina_recover_every: prev?.stamina_recover_every ?? STAMINA_RECOVER_SECONDS,
  }
}
