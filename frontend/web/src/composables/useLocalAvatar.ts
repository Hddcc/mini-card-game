/**
 * 本地头像（迁移自旧 mini_5）：纯客户端存储，不上传服务器。
 * FileReader → canvas 居中方形裁剪 512x512 → JPEG(0.86) → localStorage。
 * 注意：key 为全局（跨账号共享头像），与旧版行为一致。
 */
import { ref } from 'vue'

import { DEFAULT_AVATAR_IMAGE } from '@/constants/heroAssets'
import { STORAGE_KEYS } from '@/constants/storage'

export function useLocalAvatar(onError: (message: string) => void) {
  const avatarSrc = ref(localStorage.getItem(STORAGE_KEYS.localAvatar) || DEFAULT_AVATAR_IMAGE)

  function resizeAndStore(dataUrl: string): void {
    const image = new Image()
    image.onload = () => {
      const size = 512
      const canvas = document.createElement('canvas')
      canvas.width = size
      canvas.height = size
      const ctx = canvas.getContext('2d')
      if (!ctx) return
      const side = Math.min(image.width, image.height)
      const sx = Math.max(0, (image.width - side) / 2)
      const sy = Math.max(0, (image.height - side) / 2)
      ctx.drawImage(image, sx, sy, side, side, 0, 0, size, size)
      const avatarData = canvas.toDataURL('image/jpeg', 0.86)
      try {
        localStorage.setItem(STORAGE_KEYS.localAvatar, avatarData)
        avatarSrc.value = avatarData
      } catch {
        onError('头像图片过大，请选择更小的图片')
      }
    }
    image.onerror = () => onError('头像图片无法解析，请换一张图片试试')
    image.src = dataUrl
  }

  function handleFileChange(event: Event): void {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return
    if (!file.type || !file.type.startsWith('image/')) {
      onError('请选择图片文件作为头像')
      return
    }
    const reader = new FileReader()
    reader.onload = () => resizeAndStore(String(reader.result || ''))
    reader.onerror = () => onError('头像读取失败，请换一张图片试试')
    reader.readAsDataURL(file)
  }

  return { avatarSrc, handleFileChange }
}
