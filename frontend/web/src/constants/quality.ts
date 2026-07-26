/** 神将品质（1-5）展示映射，替代旧页 4 处重复定义。 */

export type Quality = 1 | 2 | 3 | 4 | 5

/** 品质名（与旧页一致：1/2 都显示 N） */
export const QUALITY_NAMES: Record<number, string> = {
  5: 'SSR',
  4: 'SR',
  3: 'R',
  2: 'N',
  1: 'N',
}

export const QUALITY_TEXT_CLASS: Record<number, string> = {
  5: 'text-quality-ssr',
  4: 'text-quality-sr',
  3: 'text-quality-r',
  2: 'text-quality-n',
  1: 'text-quality-n',
}

/** 品质色 hex（内联样式场景，如抽卡结果边框） */
export const QUALITY_COLORS: Record<number, string> = {
  5: '#FFD700',
  4: '#DA70D6',
  3: '#1E90FF',
  2: '#A9A9A9',
  1: '#A9A9A9',
}

export function qualityColor(quality: number): string {
  return QUALITY_COLORS[quality] ?? QUALITY_COLORS[1]!
}

export const QUALITY_BORDER_CLASS: Record<number, string> = {
  5: 'border-quality-ssr',
  4: 'border-quality-sr',
  3: 'border-quality-r',
  2: 'border-quality-n',
  1: 'border-quality-n',
}

export function qualityName(quality: number): string {
  return QUALITY_NAMES[quality] ?? 'N'
}

export function qualityTextClass(quality: number): string {
  return QUALITY_TEXT_CLASS[quality] ?? QUALITY_TEXT_CLASS[1]!
}

export function qualityBorderClass(quality: number): string {
  return QUALITY_BORDER_CLASS[quality] ?? QUALITY_BORDER_CLASS[1]!
}
