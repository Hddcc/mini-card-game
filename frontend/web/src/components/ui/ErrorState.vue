<script setup lang="ts">
import StoneButton from './StoneButton.vue'

withDefaults(
  defineProps<{
    message: string
    retryText?: string
  }>(),
  { retryText: '重试' },
)

defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-stack-md py-12 text-center">
    <div class="flex h-16 w-16 items-center justify-center rounded-full border-2 border-error bg-ink-wash">
      <span class="material-symbols-outlined text-3xl text-error" aria-hidden="true">error</span>
    </div>
    <p class="font-body-md text-body-md text-on-surface">{{ message }}</p>
    <StoneButton variant="ghost" size="sm" @click="$emit('retry')">
      <span class="material-symbols-outlined text-base" aria-hidden="true">refresh</span>
      {{ retryText }}
    </StoneButton>
  </div>
</template>
