package model

import "time"

type DailyTaskConfig struct {
	ID            uint64    `gorm:"primaryKey;column:id"`
	Name          string    `gorm:"column:name"`
	EventType     string    `gorm:"column:event_type"`
	TargetCount   uint32    `gorm:"column:target_count"`
	RewardGold    uint64    `gorm:"column:reward_gold"`
	RewardDiamond uint64    `gorm:"column:reward_diamond"`
	Status        uint8     `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (DailyTaskConfig) TableName() string {
	return "daily_task_config"
}

type PlayerDailyTask struct {
	ID        uint64     `gorm:"primaryKey;column:id"`
	PlayerID  uint64     `gorm:"column:player_id"`
	TaskID    uint64     `gorm:"column:task_id"`
	TaskDate  time.Time  `gorm:"column:task_date"`
	Progress  uint32     `gorm:"column:progress"`
	Status    uint8      `gorm:"column:status"`
	ClaimedAt *time.Time `gorm:"column:claimed_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (PlayerDailyTask) TableName() string {
	return "player_daily_task"
}
