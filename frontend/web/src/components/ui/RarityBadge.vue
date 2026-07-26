<script setup lang="ts">
/** 品质徽章（DESIGN.md: Status Badges）。SSR/UR 带扫光。 */
import { computed } from 'vue'

import { qualityBorderClass, qualityName, qualityTextClass } from '@/constants/quality'

const props = defineProps<{ quality: number }>()

const shimmer = computed(() => props.quality >= 5)
</script>

<template>
  <span
    class="inline-flex items-center rounded-full border px-2 py-0.5 font-label-sm text-label-sm"
    :class="[qualityBorderClass(quality), shimmer ? 'shimmer-text' : qualityTextClass(quality)]"
  >
    {{ qualityName(quality) }}
  </span>
</template>
