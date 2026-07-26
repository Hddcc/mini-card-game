import { copyFileSync, mkdirSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, type Plugin } from 'vite'

const BACKEND_ORIGIN = 'http://localhost:5290'

/**
 * 后端 router.go 的回落逻辑优先取 <FRONTEND_DIST>/mini_1/code.html，
 * 构建后把 index.html 复制一份过去，让 Go 侧走确定性分支（无需改后端）。
 */
function goFallbackCopy(): Plugin {
  return {
    name: 'go-fallback-copy',
    apply: 'build',
    enforce: 'post',
    // writeBundle 在产物全部落盘后触发（Vite 8/rolldown 下 closeBundle 时机偏早）
    writeBundle() {
      const dist = fileURLToPath(new URL('./dist', import.meta.url))
      mkdirSync(`${dist}/mini_1`, { recursive: true })
      copyFileSync(`${dist}/index.html`, `${dist}/mini_1/code.html`)
    },
  }
}

export default defineConfig(({ command }) => ({
  // 生产由 Go 以 /static 前缀托管构建产物；dev 用 '/' 保持页面 URL 与生产一致
  base: command === 'build' ? '/static/' : '/',
  plugins: [vue(), goFallbackCopy()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // API 与后端下发的绝对图片路径（/static/assets/images/*.png）都转发到 Go
      '/api': { target: BACKEND_ORIGIN, changeOrigin: true },
      '/static': { target: BACKEND_ORIGIN, changeOrigin: true },
    },
  },
}))
