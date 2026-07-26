<script setup lang="ts">
/** 全局 toast 出口，挂在 App.vue，队列由 useToast 驱动。 */
import { useToast } from '@/composables/useToast'

const { toasts } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed bottom-6 left-1/2 z-[90] flex -translate-x-1/2 flex-col items-center gap-2">
      <TransitionGroup
        enter-active-class="transition duration-200"
        enter-from-class="opacity-0 translate-y-2"
        leave-active-class="transition duration-200"
        leave-to-class="opacity-0"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="rounded-lg border border-primary-fixed bg-[#1b1711]/95 px-6 py-3 font-title-md text-sm text-primary-fixed shadow-2xl"
        >
          {{ toast.text }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
