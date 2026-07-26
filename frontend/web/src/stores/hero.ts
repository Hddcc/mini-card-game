import { defineStore } from 'pinia'
import { ref } from 'vue'

import { fetchHeroes } from '@/api/hero'
import { errorText } from '@/types/api'
import type { HeroView } from '@/types/hero'

export const useHeroStore = defineStore('hero', () => {
  const heroes = ref<HeroView[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function refresh(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      heroes.value = await fetchHeroes()
    } catch (err) {
      error.value = errorText(err, '神将列表加载失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  /** 升星后用接口返回的最新数据替换单条 */
  function replaceHero(updated: HeroView): void {
    heroes.value = heroes.value.map((hero) =>
      hero.player_hero_id === updated.player_hero_id ? updated : hero,
    )
  }

  function byId(playerHeroId: number): HeroView | undefined {
    return heroes.value.find((hero) => hero.player_hero_id === playerHeroId)
  }

  function reset(): void {
    heroes.value = []
    loading.value = false
    error.value = null
  }

  return { heroes, loading, error, refresh, replaceHero, byId, reset }
})
