<script setup lang="ts">
/** 阵容槽位：空位虚线 + 待命提示；有将显示立绘/名字/品质；右上角清除。 */
import { computed } from 'vue'

import { HERO_IMAGES, heroImage } from '@/constants/heroAssets'
import { qualityName, qualityTextClass } from '@/constants/quality'
import type { HeroView } from '@/types/hero'

const props = withDefaults(
  defineProps<{
    slotIndex: number
    hero: HeroView | null
    selected?: boolean
  }>(),
  { selected: false },
)

defineEmits<{ select: []; clear: [] }>()

const image = computed(() =>
  props.hero ? (heroImage(props.hero.hero_config_id) ?? HERO_IMAGES[1]!) : '',
)
</script>

<template>
  <div
    class="relative flex h-40 cursor-pointer flex-col overflow-hidden border-2 bg-surface-container-low transition-all"
    :class="selected ? 'border-primary-fixed shadow-[0_0_15px_rgba(255,215,0,0.4)]' : 'border-outline-variant'"
    @click="$emit('select')"
  >
    <!-- 空槽 -->
    <div v-if="!hero" class="flex flex-1 flex-col items-center justify-center gap-1 border border-dashed border-outline-variant/60 m-1">
      <span class="material-symbols-outlined text-2xl text-on-surface-variant" aria-hidden="true">add</span>
      <span class="font-label-sm text-[10px] text-on-surface-variant">待命 · {{ slotIndex + 1 }} 号位</span>
    </div>

    <!-- 已上阵 -->
    <template v-else>
      <img :src="image" :alt="hero.name" class="absolute inset-0 h-full w-full object-cover" />
      <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black to-transparent p-2">
        <p class="truncate font-title-md text-xs text-white">{{ hero.name }}</p>
        <p class="font-label-sm text-[10px]" :class="qualityTextClass(hero.quality)">
          {{ qualityName(hero.quality) }} · {{ hero.star }}星
        </p>
      </div>
      <button
        class="absolute right-1 top-1 z-10 flex h-6 w-6 items-center justify-center rounded-full border border-outline-variant bg-black/70 text-on-surface-variant transition hover:border-error hover:text-error"
        title="移出阵容"
        @click.stop="$emit('clear')"
      >
        <span class="material-symbols-outlined text-sm" aria-hidden="true">close</span>
      </button>
    </template>
  </div>
</template>
