<script setup lang="ts">
/** 模态弹窗：黑色背板 + 金边回纹面板。locked 时禁止 Esc/点遮罩关闭（战斗结算等场景）。 */
import { onBeforeUnmount, onMounted } from 'vue'

import JadePanel from './JadePanel.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    locked?: boolean
    maxWidthClass?: string
  }>(),
  { locked: false, maxWidthClass: 'max-w-md' },
)

const emit = defineEmits<{
  close: []
}>()

function requestClose(): void {
  if (!props.locked) emit('close')
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && props.open) requestClose()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-active-class="transition-opacity duration-150"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-black/90 p-margin-mobile backdrop-blur-sm"
        @click.self="requestClose"
      >
        <JadePanel tone="gold" class="w-full overflow-hidden" :class="maxWidthClass">
          <slot />
        </JadePanel>
      </div>
    </Transition>
  </Teleport>
</template>
