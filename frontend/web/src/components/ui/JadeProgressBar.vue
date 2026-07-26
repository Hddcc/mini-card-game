<script setup lang="ts">
/** 玉璧进度条（DESIGN.md: Jade Progress Bars）。严格矩形、内凹轨道、玻璃高光。 */
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    value: number
    max: number
    variant?: 'jade' | 'gold'
    heightClass?: string
  }>(),
  { variant: 'jade', heightClass: 'h-2' },
)

const percent = computed(() => {
  if (props.max <= 0) return 0
  return Math.max(0, Math.min(100, (props.value / props.max) * 100))
})

const fillStyle = computed(() =>
  props.variant === 'gold'
    ? 'linear-gradient(to bottom, #ffe16d, #e9c400)'
    : 'linear-gradient(to bottom, #4ADE80, #00A86B)',
)
</script>

<template>
  <div
    class="relative w-full overflow-hidden rounded-sm border border-black/40 bg-surface-container-lowest shadow-[inset_0_1px_3px_rgba(0,0,0,0.6)]"
    :class="heightClass"
    role="progressbar"
    :aria-valuenow="Math.round(percent)"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div
      class="h-full transition-all duration-300"
      :style="{ width: `${percent}%`, backgroundImage: fillStyle }"
    ></div>
    <div class="jade-gloss pointer-events-none absolute inset-0"></div>
  </div>
</template>
