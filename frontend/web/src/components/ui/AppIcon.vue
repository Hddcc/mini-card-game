<script setup lang="ts">
/** Material Symbols 图标统一封装（ligature 名 + 可选实心变体） */
withDefaults(
  defineProps<{
    name: string
    fill?: boolean
    /** 传数字按 px，传字符串原样作为 font-size */
    size?: number | string
  }>(),
  { fill: false, size: undefined },
)
</script>

<template>
  <span
    class="material-symbols-outlined select-none leading-none"
    :class="{ 'icon-filled': fill }"
    :style="size !== undefined ? { fontSize: typeof size === 'number' ? `${size}px` : size } : undefined"
    aria-hidden="true"
    >{{ name }}</span
  >
</template>
