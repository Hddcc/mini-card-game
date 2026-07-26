<script setup lang="ts">
withDefaults(defineProps<{ text?: string }>(), { text: '加载中…' })
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-stack-md py-10 text-on-surface-variant">
    <span class="material-symbols-outlined animate-spin text-3xl text-primary-fixed" aria-hidden="true"
      >progress_activity</span
    >
    <p class="font-label-sm text-label-sm">{{ text }}</p>
  </div>
</template>
