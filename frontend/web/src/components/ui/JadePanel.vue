<script setup lang="ts">
/** 带四角回纹角饰的容器（DESIGN.md: Ornate Borders），全站唯一回纹实现。 */
withDefaults(
  defineProps<{
    /** gold 用于弹窗/强调面板（金边），subtle 用于普通卡片（暗边+金角） */
    tone?: 'gold' | 'subtle'
    corners?: boolean
  }>(),
  { tone: 'subtle', corners: true },
)
</script>

<template>
  <div
    class="relative bg-ink-wash"
    :class="tone === 'gold' ? 'border-2 border-primary-fixed' : 'border-2 border-outline-variant'"
  >
    <template v-if="corners">
      <div class="meander-corner meander-tl"></div>
      <div class="meander-corner meander-tr"></div>
      <div class="meander-corner meander-bl"></div>
      <div class="meander-corner meander-br"></div>
    </template>
    <slot />
  </div>
</template>
