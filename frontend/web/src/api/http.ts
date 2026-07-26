/**
 * 统一 HTTP 客户端，替代旧前端 6 份重复的 apiFetch。
 * - Bearer 注入 + 401 统一登出
 * - 8s AbortController 超时（推广自旧升星页）
 * - content-type 非 JSON 检测（后端 NoRoute 对未知路径返回 HTML 200）
 * - code !== 0 时抛 ApiError，friendly 为集中翻译后的中文
 */
import { API_TIMEOUT_MS } from '@/constants'
import { translateMessage } from '@/constants/messages'
import { STORAGE_KEYS } from '@/constants/storage'
import { ApiError, type ApiEnvelope } from '@/types/api'

interface RequestOptions {
  method?: 'GET' | 'POST'
  body?: unknown
  /** 默认 true；登录/注册接口传 false */
  auth?: boolean
  timeoutMs?: number
}

let unauthorizedHandler: (() => void) | null = null

/** 由 main.ts 注册：401 时清理会话并跳转登录页（避免 http 层反向依赖 store/router） */
export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, auth = true, timeoutMs = API_TIMEOUT_MS } = options

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (auth) {
    const token = localStorage.getItem(STORAGE_KEYS.token)
    if (!token) {
      unauthorizedHandler?.()
      throw new ApiError('未登录，请先登录', 'no token', undefined, 'auth')
    }
    headers.Authorization = `Bearer ${token}`
  }

  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const response = await fetch(path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    })

    if (response.status === 401 && auth) {
      unauthorizedHandler?.()
      throw new ApiError('登录已过期，请重新登录', 'unauthorized', 40001, 'auth')
    }

    const contentType = response.headers.get('content-type') ?? ''
    if (!contentType.includes('application/json')) {
      throw new ApiError('接口响应异常，请稍后重试', `non-json response: ${response.status}`, undefined, 'non-json')
    }

    const envelope = (await response.json()) as ApiEnvelope<T>
    if (envelope.code !== 0) {
      throw new ApiError(translateMessage(envelope.message), envelope.message, envelope.code)
    }
    return envelope.data
  } catch (error) {
    if (error instanceof ApiError) throw error
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError('请求超时，请稍后重试', 'timeout', undefined, 'timeout')
    }
    throw new ApiError('网络错误，请检查网络后重试', String(error), undefined, 'network')
  } finally {
    clearTimeout(timer)
  }
}

export const http = {
  get<T>(path: string, options?: Omit<RequestOptions, 'method' | 'body'>): Promise<T> {
    return request<T>(path, options)
  },
  post<T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>): Promise<T> {
    return request<T>(path, { ...options, method: 'POST', body })
  },
}
