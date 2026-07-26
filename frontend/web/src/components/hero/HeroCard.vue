<script setup lang="ts">
/** 神将卡（DESIGN.md: Cards）：整卡立绘 + 底部渐变压名字/等级/星级。已上阵置灰不可点。 */
import { computed } from 'vue'

import { HERO_IMAGES, heroImage } from '@/constants/heroAssets'
import { qualityName, qualityTextClass, qualityBorderClass } from '@/constants/quality'
import type { HeroView } from '@/types/hero'

const props = withDefaults(
  defineProps<{
    hero: HeroView
    disabled?: boolean
    selected?: boolean
  }>(),
  { disabled: false, selected: false },
)

const emit = defineEmits<{ select: [hero: HeroView] }>()

const image = computed(() => heroImage(props.hero.hero_config_id) ?? HERO_IMAGES[1]!)

function onClick(): void {
  if (!props.disabled) emit('select', props.hero)
}
</script>

<template>
  <div
    class="group relative h-48 cursor-pointer overflow-hidden border-2 transition-all"
    :class="
      disabled
        ? 'border-outline-variant opacity-30 grayscale'
        : selected
          ? 'border-primary-fixed shadow-[0_0_15px_rgba(255,215,0,0.4)]'
          : 'border-outline hover:-translate-y-1 hover:border-primary-fixed'
    "
    @click="onClick"
  >
    <img :src="image" :alt="hero.name" class="absolute inset-0 h-full w-full object-cover" />
    <div
      class="absolute left-2 top-2 rounded-sm border bg-black/70 px-2 py-0.5"
      :class="qualityBorderClass(hero.quality)"
    >
      <span class="font-label-sm text-label-sm" :class="qualityTextClass(hero.quality)">{{
        qualityName(hero.quality)
      }}</span>
    </div>
    <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black to-transparent p-2">
      <p class="truncate font-title-md text-sm text-white">{{ hero.name }}</p>
      <p class="font-label-sm text-[10px] uppercase text-on-surface-variant">
        等级 {{ hero.level }} · {{ hero.star }}星
      </p>
    </div>
  </div>
</template>
