/** 统一响应包（internal/pkg/app/response.go）。成功 code === 0，字段名是 message（不是 msg）。 */
export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export type ApiErrorKind = 'api' | 'http' | 'timeout' | 'non-json' | 'network' | 'auth'

export class ApiError extends Error {
  /** 用户可读中文 */
  readonly friendly: string
  /** 后端原始 message，供关键字规则（活动页/抽卡页）判断 */
  readonly raw: string
  /** 后端业务错误码（40000 等），非业务错误时为 undefined */
  readonly code?: number
  readonly kind: ApiErrorKind

  constructor(friendly: string, raw: string, code?: number, kind: ApiErrorKind = 'api') {
    super(friendly)
    this.name = 'ApiError'
    this.friendly = friendly
    this.raw = raw
    this.code = code
    this.kind = kind
  }
}

/** 把任意异常收敛为可展示文案 */
export function errorText(error: unknown, fallback = '请求失败，请稍后重试'): string {
  if (error instanceof ApiError) return error.friendly
  if (error instanceof Error && error.message) return error.message
  return fallback
}

/** 取后端原始 message（非 ApiError 时返回空串） */
export function errorRaw(error: unknown): string {
  return error instanceof ApiError ? error.raw : ''
}
