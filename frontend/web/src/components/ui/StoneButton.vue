<script setup lang="ts">
/** 石纹按钮（DESIGN.md: Stone Buttons）。variant 对应黑金/纯金/朱红/描边四种场景。 */
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'stone' | 'primary' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    type?: 'button' | 'submit'
    disabled?: boolean
    block?: boolean
  }>(),
  { variant: 'stone', size: 'md', type: 'button', disabled: false, block: false },
)

const variantClass = computed(() => {
  switch (props.variant) {
    case 'primary':
      return 'bg-primary-fixed text-on-primary-fixed border border-primary-fixed-dim hover:brightness-110 active:translate-y-[2px] shadow-[0_0_10px_rgba(255,215,0,0.3)]'
    case 'danger':
      return 'border border-error/70 text-error hover:bg-error/10 active:translate-y-[2px]'
    case 'ghost':
      return 'border border-outline-variant text-on-surface-variant hover:text-primary-fixed hover:border-outline'
    default:
      return 'stone-button text-primary-fixed border border-outline-variant'
  }
})

const sizeClass = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'py-2 px-4 font-label-sm text-label-sm'
    case 'lg':
      return 'py-4 px-6 font-title-md text-title-md'
    default:
      return 'py-3 px-5 font-label-sm text-label-sm'
  }
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled"
    class="focus-ring flex items-center justify-center gap-stack-md transition-all disabled:cursor-not-allowed disabled:opacity-40"
    :class="[variantClass, sizeClass, block ? 'w-full' : '']"
  >
    <slot />
  </button>
</template>
