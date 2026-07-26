// 字体自托管（国内 Google Fonts CDN 不稳），随 /static/assets 分发
import '@fontsource/epilogue/600.css'
import '@fontsource/epilogue/700.css'
import '@fontsource/epilogue/800.css'
import '@fontsource/hanken-grotesk/400.css'
import '@fontsource/hanken-grotesk/500.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/600.css'
import 'material-symbols/outlined.css'

import '@/styles/main.css'
import '@/styles/effects.css'

import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from '@/api/http'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)
app.use(createPinia())
app.use(router)

// 401 统一处理：清会话 + 重置 store + 跳登录页
setUnauthorizedHandler(() => {
  useAuthStore().logout()
  void router.push({ name: 'login' })
})

app.mount('#app')
