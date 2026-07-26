import { http } from './http'
import type { LoginOutput, LoginPayload, RegisterOutput, RegisterPayload } from '@/types/auth'

export function login(payload: LoginPayload): Promise<LoginOutput> {
  return http.post<LoginOutput>('/api/v1/auth/login', payload, { auth: false })
}

export function register(payload: RegisterPayload): Promise<RegisterOutput> {
  return http.post<RegisterOutput>('/api/v1/auth/register', payload, { auth: false })
}
