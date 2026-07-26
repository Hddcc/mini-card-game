/** 全局 toast 队列，统一旧页各自的 showToast/alert 实现。 */
import { readonly, ref } from 'vue'

export interface ToastItem {
  id: number
  text: string
}

const toasts = ref<ToastItem[]>([])
let seed = 0

function show(text: string, durationMs = 2500): void {
  const id = ++seed
  toasts.value.push({ id, text })
  setTimeout(() => {
    toasts.value = toasts.value.filter((item) => item.id !== id)
  }, durationMs)
}

export function useToast() {
  return { toasts: readonly(toasts), show }
}
