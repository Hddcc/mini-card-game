/** internal/service/activity_lottery_service.go */

export interface ActivityView {
  id: number
  code: string
  name: string
  description: string
  banner_image: string
  start_at: string
  end_at: string
}

export interface ActivityQuotaView {
  daily_limit: number
  used_today: number
  remaining_today: number
  ip_daily_limit: number
  ip_used_today: number
}

export interface ActivityPrizeView {
  id: number
  name: string
  description: string
  icon: string
  /** gold | diamond | stamina | hero */
  reward_type: string
  reward_id: number
  reward_count: number
  quality: number
  left_num: number
  unlimited: boolean
  display_order: number
}

export interface ActivityRecordView {
  draw_no: string
  prize_name: string
  reward_type: string
  reward_id: number
  reward_count: number
  /** 0 待发放 / 1 成功 / 2 失败（异步发奖） */
  delivery_status: number
  created_at: string
}

export interface ActivityStateView {
  active: boolean
  activity?: ActivityView | null
  quota?: ActivityQuotaView | null
  eligible: boolean
  reason: string
  prizes: ActivityPrizeView[] | null
  records: ActivityRecordView[] | null
}

export interface ActivityDrawOutput {
  draw_no: string
  prize: ActivityPrizeView
  quota: ActivityQuotaView
  delivery_status: number
}
