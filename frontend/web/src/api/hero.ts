import { http } from './http'
import type { HeroListData, HeroView, StarUpOutput, StarUpPayload } from '@/types/hero'

/** GET /heroes 的 data 为 {heroes:[...]}，这里剥壳返回数组 */
export async function fetchHeroes(): Promise<HeroView[]> {
  const data = await http.get<HeroListData>('/api/v1/heroes')
  return data.heroes ?? []
}

export function starUpHero(payload: StarUpPayload): Promise<StarUpOutput> {
  return http.post<StarUpOutput>('/api/v1/heroes/star-up', payload)
}
