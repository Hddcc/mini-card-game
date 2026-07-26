<script setup lang="ts">
/** 输入框（DESIGN.md: Input Fields）：墨底 + 金色底边，聚焦发光。 */
withDefaults(
  defineProps<{
    modelValue: string
    type?: string
    placeholder?: string
    icon?: string
    disabled?: boolean
    maxlength?: number
    name?: string
  }>(),
  { type: 'text', placeholder: '', icon: undefined, disabled: false, maxlength: undefined, name: undefined },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

function onInput(event: Event): void {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="group relative">
    <span
      v-if="icon"
      class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors group-focus-within:text-primary-container"
      aria-hidden="true"
      >{{ icon }}</span
    >
    <input
      :value="modelValue"
      :type="type"
      :placeholder="placeholder"
      :disabled="disabled"
      :maxlength="maxlength"
      :name="name"
      class="w-full border-b-2 border-outline-variant bg-surface-container-low py-3 text-on-surface outline-none transition-all placeholder:text-surface-variant focus:border-primary-container disabled:opacity-50"
      :class="icon ? 'px-10' : 'px-4'"
      @input="onInput"
    />
  </div>
</template>
