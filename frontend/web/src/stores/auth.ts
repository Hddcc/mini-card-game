import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { login as apiLogin, register as apiRegister } from '@/api/auth'
import { STORAGE_KEYS } from '@/constants/storage'
import type { LoginOutput, LoginPayload, RegisterOutput, RegisterPayload } from '@/types/auth'

import { useAssetStore } from './assets'
import { useBattleStore } from './battle'
import { useHeroStore } from './hero'
import { usePlayerStore } from './player'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(STORAGE_KEYS.token))

  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(payload: LoginPayload): Promise<LoginOutput> {
    const output = await apiLogin(payload)
    token.value = output.token
    localStorage.setItem(STORAGE_KEYS.token, output.token)
    return output
  }

  function register(payload: RegisterPayload): Promise<RegisterOutput> {
    return apiRegister(payload)
  }

  /** 清会话 + 重置所有游戏 store（登出与 401 共用；路由跳转由调用方负责） */
  function logout(): void {
    token.value = null
    localStorage.removeItem(STORAGE_KEYS.token)
    useAssetStore().reset()
    usePlayerStore().reset()
    useHeroStore().reset()
    useBattleStore().reset()
  }

  return { token, isAuthenticated, login, register, logout }
})
