import { http } from './http'
import type { AssetView, ProfileView, UpdateNamePayload } from '@/types/player'

export function fetchProfile(): Promise<ProfileView> {
  return http.get<ProfileView>('/api/v1/player/profile')
}

export function fetchAssets(): Promise<AssetView> {
  return http.get<AssetView>('/api/v1/player/assets')
}

export function updatePlayerName(payload: UpdateNamePayload): Promise<ProfileView> {
  return http.post<ProfileView>('/api/v1/player/name', payload)
}
