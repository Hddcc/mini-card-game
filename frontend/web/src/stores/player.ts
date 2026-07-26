import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchProfile, updatePlayerName } from '@/api/player'
import { expMaxForLevel } from '@/constants'
import type { ProfileView } from '@/types/player'

export const usePlayerStore = defineStore('player', () => {
  const profile = ref<ProfileView | null>(null)
  const loading = ref(false)

  const nickname = computed(() => profile.value?.name || profile.value?.nickname || '取经人')
  const level = computed(() => profile.value?.level ?? 1)
  const power = computed(() => profile.value?.power ?? 0)
  /** 经验条分母为客户端约定公式（后端未下发升级经验） */
  const expMax = computed(() => expMaxForLevel(level.value))
  const expPercent = computed(() => {
    const exp = profile.value?.exp ?? 0
    return Math.min(100, Math.round((exp / expMax.value) * 100))
  })

  async function refresh(): Promise<void> {
    loading.value = true
    try {
      profile.value = await fetchProfile()
    } finally {
      loading.value = false
    }
  }

  /** 改名成功后用返回的最新资料覆盖；失败抛 ApiError（服务端中文文案），由视图行内展示 */
  async function updateName(name: string): Promise<void> {
    profile.value = await updatePlayerName({ name })
  }

  function reset(): void {
    profile.value = null
    loading.value = false
  }

  return { profile, loading, nickname, level, power, expMax, expPercent, refresh, updateName, reset }
})
