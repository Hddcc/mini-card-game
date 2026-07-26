<script setup lang="ts">
/**
 * 单张抽卡结果卡（迁移自旧 mini_4 renderResults）：
 * - 品质色边框 + 右上角品质角标
 * - 重复神将碎片化：立绘缺角去饱和 + 右下缺口 + 左下 +N 碎片角标 + 名称加"碎片"
 * - 保底标记：名称后缀 " · 保底"
 * - 逐张延迟揭示（10 + index*100ms）
 */
import { computed, onMounted, ref } from 'vue'

import { HERO_IMAGES, HERO_NAMES, IMAGE_REWARD_GOLD, heroImage } from '@/constants/heroAssets'
import { qualityColor, qualityName } from '@/constants/quality'
import type { DrawResult } from '@/types/gacha'

const props = defineProps<{
  item: DrawResult
  index: number
  single: boolean
}>()

const revealed = ref(false)

const isHero = computed(() => props.item.item_type === 'hero')
const isDuplicateHero = computed(
  () => isHero.value && props.item.is_duplicate && Number(props.item.converted_shards || 0) > 0,
)
const rarity = computed(() => qualityName(props.item.quality))
const color = computed(() => qualityColor(props.item.quality))
const image = computed(() => (isHero.value ? (heroImage(props.item.item_id) ?? HERO_IMAGES[1]!) : IMAGE_REWARD_GOLD))
const baseName = computed(() =>
  isHero.value ? (HERO_NAMES[props.item.item_id] ?? `英雄 ${props.item.item_id}`) : `金币 x${props.item.item_count}`,
)
const displayName = computed(() => {
  const name = isDuplicateHero.value ? `${baseName.value}碎片` : baseName.value
  return props.item.is_pity ? `${name} · 保底` : name
})

onMounted(() => {
  setTimeout(() => {
    revealed.value = true
  }, 10 + props.index * 100)
})
</script>

<template>
  <div
    class="flex w-full flex-col items-center transition-all duration-500"
    :class="[revealed ? 'scale-100 opacity-100' : 'scale-50 opacity-0', single ? 'max-w-[260px]' : '']"
  >
    <div
      class="relative aspect-square w-full overflow-hidden rounded-xl border-2 shadow-2xl"
      :style="{ borderColor: color }"
    >
      <img
        :src="image"
        :alt="displayName"
        class="h-full w-full object-cover"
        :class="isDuplicateHero ? 'duplicate-shard-img' : ''"
      />
      <div v-if="isDuplicateHero" class="duplicate-shard-notch"></div>
      <div class="absolute right-0 top-0 bg-black/60 px-2 py-0.5 font-label-sm text-[10px]" :style="{ color }">
        {{ rarity }}
      </div>
      <div
        v-if="isDuplicateHero"
        class="absolute bottom-2 left-2 flex h-8 min-w-8 items-center justify-center rounded-full border border-primary-fixed bg-black/75 px-2 font-stats-num text-primary-fixed"
      >
        +{{ item.converted_shards }}
      </div>
    </div>
    <span class="mt-2 line-clamp-1 text-center font-label-sm text-label-sm">{{ displayName }}</span>
  </div>
</template>
