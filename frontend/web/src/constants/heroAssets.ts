/**
 * 美术资源路径。全部为 /static 绝对路径：dev 由 Vite proxy 转发到后端，
 * 生产由 Go 托管 dist（public/assets/images 随构建产物分发）。
 * 与后端 seed（internal/model/db.go）中的 hero_config_id 对应。
 */

export const HERO_IMAGES: Record<number, string> = {
  1: '/static/assets/images/hero-sun-wukong.png',
  2: '/static/assets/images/hero-zhu-bajie.png',
  3: '/static/assets/images/hero-sha-wujing.png',
  4: '/static/assets/images/hero-xiao-bailong.png',
  5: '/static/assets/images/hero-tang-sanzang.png',
}

export function heroImage(heroConfigId: number): string | undefined {
  return HERO_IMAGES[heroConfigId]
}

/** 神将名（hero_config_id → 名称），抽卡结果等场景使用 */
export const HERO_NAMES: Record<number, string> = {
  1: '孙悟空',
  2: '猪八戒',
  3: '沙悟净',
  4: '小白龙',
  5: '唐三藏',
}

/** 大厅默认头像（旧 mini_5 行为） */
export const DEFAULT_AVATAR_IMAGE = '/static/assets/images/hero-xiao-bailong.png'

export const IMAGE_HOME_BACKGROUND = '/static/assets/images/home-background.png'
export const IMAGE_ACTIVITY_BANNER = '/static/assets/images/activity-flame-mountain.png'
export const IMAGE_REWARD_GOLD = '/static/assets/images/reward-gold.png'
