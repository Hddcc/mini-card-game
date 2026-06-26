package model

import "time"

type ActivityLottery struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	Code         string    `gorm:"column:code;size:64;uniqueIndex"`
	Name         string    `gorm:"column:name;size:64"`
	Description  string    `gorm:"column:description;size:255"`
	BannerImage  string    `gorm:"column:banner_image;size:255"`
	DailyLimit   uint32    `gorm:"column:daily_limit"`
	IPDailyLimit uint32    `gorm:"column:ip_daily_limit"`
	Status       uint8     `gorm:"column:status"`
	StartAt      time.Time `gorm:"column:start_at"`
	EndAt        time.Time `gorm:"column:end_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ActivityLottery) TableName() string {
	return "activity_lottery"
}

type ActivityLotteryPrize struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	ActivityID   uint64    `gorm:"column:activity_id;index"`
	Name         string    `gorm:"column:name;size:64"`
	Description  string    `gorm:"column:description;size:255"`
	Icon         string    `gorm:"column:icon;size:64"`
	RewardType   string    `gorm:"column:reward_type;size:32"`
	RewardID     uint64    `gorm:"column:reward_id"`
	RewardCount  uint64    `gorm:"column:reward_count"`
	Quality      uint8     `gorm:"column:quality"`
	Weight       uint32    `gorm:"column:weight"`
	TotalNum     int64     `gorm:"column:total_num"`
	LeftNum      int64     `gorm:"column:left_num"`
	ReleasePlan  string    `gorm:"column:release_plan;type:text"`
	DisplayOrder uint32    `gorm:"column:display_order"`
	Fallback     uint8     `gorm:"column:fallback"`
	Status       uint8     `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ActivityLotteryPrize) TableName() string {
	return "activity_lottery_prize"
}

type ActivityLotteryRecord struct {
	ID             uint64    `gorm:"primaryKey;column:id"`
	ActivityID     uint64    `gorm:"column:activity_id;index"`
	PrizeID        uint64    `gorm:"column:prize_id;index"`
	PlayerID       uint64    `gorm:"column:player_id;index"`
	DrawNo         string    `gorm:"column:draw_no;size:64;uniqueIndex"`
	RandomPoint    uint32    `gorm:"column:random_point"`
	PrizeName      string    `gorm:"column:prize_name;size:64"`
	RewardType     string    `gorm:"column:reward_type;size:32"`
	RewardID       uint64    `gorm:"column:reward_id"`
	RewardCount    uint64    `gorm:"column:reward_count"`
	DeliveryStatus uint8     `gorm:"column:delivery_status"`
	RequestIP      string    `gorm:"column:request_ip;size:64"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (ActivityLotteryRecord) TableName() string {
	return "activity_lottery_record"
}

type ActivityLotteryBlacklist struct {
	ID         uint64     `gorm:"primaryKey;column:id"`
	TargetType string     `gorm:"column:target_type;size:16;index:idx_activity_blacklist"`
	Target     string     `gorm:"column:target;size:64;index:idx_activity_blacklist"`
	Reason     string     `gorm:"column:reason;size:255"`
	ExpireAt   *time.Time `gorm:"column:expire_at"`
	Status     uint8      `gorm:"column:status"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
}

func (ActivityLotteryBlacklist) TableName() string {
	return "activity_lottery_blacklist"
}

type ActivityLotteryLocalMessage struct {
	ID          uint64     `gorm:"primaryKey;column:id"`
	BusinessID  string     `gorm:"column:business_id;size:64;uniqueIndex"`
	MessageType string     `gorm:"column:message_type;size:32"`
	Payload     string     `gorm:"column:payload;type:text"`
	Status      uint8      `gorm:"column:status;index"`
	RetryCount  uint32     `gorm:"column:retry_count"`
	NextRetryAt *time.Time `gorm:"column:next_retry_at;index"`
	LastError   string     `gorm:"column:last_error;size:255"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (ActivityLotteryLocalMessage) TableName() string {
	return "activity_lottery_local_message"
}

type ActivityPrizeReleaseState struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	PrizeID   uint64    `gorm:"column:prize_id;uniqueIndex:uk_prize_window"`
	WindowKey string    `gorm:"column:window_key;size:64;uniqueIndex:uk_prize_window"`
	Released  int64     `gorm:"column:released"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ActivityPrizeReleaseState) TableName() string {
	return "activity_prize_release_state"
}
