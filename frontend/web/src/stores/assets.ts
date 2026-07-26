/**
 * 玩家资产 + 体力恢复倒计时。
 * 等价旧 status-bar.js：next_stamina_seconds 逐秒递减，归零自动重拉资产；
 * 旧的 monkey-patch fetch 改为各业务动作显式调用 refresh()/apply()。
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchAssets } from '@/api/player'
import { MAX_STAMINA } from '@/constants'
import type { AssetView } from '@/types/player'

export const useAssetStore = defineStore('assets', () => {
  const assets = ref<AssetView | null>(null)
  /** 距下一点体力恢复的剩余秒数 */
  const countdown = ref(0)
  const loading = ref(false)

  let timer: number | null = null
  let refreshing = false

  const gold = computed(() => assets.value?.gold ?? 0)
  const diamond = computed(() => assets.value?.diamond ?? 0)
  const stamina = computed(() => assets.value?.stamina ?? 0)
  const maxStamina = computed(() => assets.value?.max_stamina ?? MAX_STAMINA)
  const staminaFull = computed(() => stamina.value >= maxStamina.value)

  async function refresh(): Promise<void> {
    if (refreshing) return
    refreshing = true
    loading.value = true
    try {
      const view = await fetchAssets()
      assets.value = view
      countdown.value = view.next_stamina_seconds ?? 0
      startTicker()
    } finally {
      refreshing = false
      loading.value = false
    }
  }

  /** 直接写入接口带回的最新资产（如升星返回值），不发额外请求，倒计时不打断 */
  function apply(view: AssetView): void {
    assets.value = view
  }

  function tick(): void {
    if (!assets.value || staminaFull.value) {
      countdown.value = 0
      return
    }
    if (countdown.value > 1) {
      countdown.value -= 1
    } else {
      // 归零：体力+1 落库靠服务端惰性结算，这里重拉一次同步（自愈刷新）
      void refresh().catch(() => undefined)
    }
  }

  function startTicker(): void {
    if (timer !== null) return
    timer = window.setInterval(tick, 1000)
  }

  function stopTicker(): void {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
  }

  function reset(): void {
    stopTicker()
    assets.value = null
    countdown.value = 0
    loading.value = false
  }

  return {
    assets,
    countdown,
    loading,
    gold,
    diamond,
    stamina,
    maxStamina,
    staminaFull,
    refresh,
    apply,
    reset,
  }
})
