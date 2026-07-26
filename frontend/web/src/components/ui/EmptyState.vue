<script setup lang="ts">
withDefaults(
  defineProps<{
    icon?: string
    title: string
    description?: string
  }>(),
  { icon: 'auto_awesome', description: '' },
)
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-stack-md py-12 text-center">
    <div class="flex h-16 w-16 items-center justify-center rounded-full border-2 border-outline-variant bg-ink-wash">
      <span class="material-symbols-outlined text-3xl text-on-surface-variant" aria-hidden="true">{{ icon }}</span>
    </div>
    <p class="font-title-md text-title-md text-on-surface">{{ title }}</p>
    <p v-if="description" class="font-body-md text-sm text-on-surface-variant">{{ description }}</p>
    <slot />
  </div>
</template>
