<script setup lang="ts">
/**
 * 登录后主壳：侧栏（桌面）/ 底部 tab（移动）+ 顶栏（页面标题 + 资产栏）。
 * 进壳拉取资产并启动体力倒计时；战斗锁定时刷新/关页弹原生确认
 * （后端会话 30 分钟内可续战，见 battle_service.go）。
 */
import { onBeforeUnmount, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import AssetBar from '@/components/layout/AssetBar.vue'
import MobileTabBar from '@/components/layout/MobileTabBar.vue'
import SideNav from '@/components/layout/SideNav.vue'
import InkBackground from '@/components/ui/InkBackground.vue'
import { useAssetStore } from '@/stores/assets'
import { useBattleStore } from '@/stores/battle'

const route = useRoute()
const assetStore = useAssetStore()
const battleStore = useBattleStore()

function onBeforeUnload(event: BeforeUnloadEvent): void {
  if (battleStore.isLocked) {
    event.preventDefault()
  }
}

onMounted(() => {
  void assetStore.refresh().catch(() => undefined)
  window.addEventListener('beforeunload', onBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', onBeforeUnload)
})
</script>

<template>
  <div class="min-h-screen md:pl-72">
    <InkBackground />
    <SideNav />

    <header
      class="sticky top-0 z-30 flex items-center justify-between border-b border-outline-variant bg-background/85 px-margin-mobile py-3 backdrop-blur md:px-margin-desktop"
    >
      <h1 class="font-display-hero text-title-md text-on-surface md:text-headline-lg-mobile">
        {{ route.meta.title ?? 'Mini 西游' }}
      </h1>
      <AssetBar />
    </header>

    <main class="relative z-10 px-margin-mobile py-stack-lg pb-24 md:px-margin-desktop md:pb-stack-lg">
      <router-view />
    </main>

    <MobileTabBar />
  </div>
</template>
