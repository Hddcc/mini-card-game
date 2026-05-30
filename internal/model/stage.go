package model

import "time"

type StageConfig struct {
	ID             uint64    `gorm:"primaryKey;column:id"`
	Name           string    `gorm:"column:name;size:64"`
	Chapter        uint32    `gorm:"column:chapter"`
	PrevStageID    uint64    `gorm:"column:prev_stage_id"`
	StaminaCost    uint32    `gorm:"column:stamina_cost"`
	RecommendPower uint64    `gorm:"column:recommend_power"`
	RewardGold     uint64    `gorm:"column:reward_gold"`
	RewardExp      uint32    `gorm:"column:reward_exp"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (StageConfig) TableName() string {
	return "stage_config"
}

type PlayerStage struct {
	ID            uint64     `gorm:"primaryKey;column:id"`
	PlayerID      uint64     `gorm:"column:player_id"`
	StageID       uint64     `gorm:"column:stage_id"`
	Status        uint8      `gorm:"column:status"`
	BestPower     uint64     `gorm:"column:best_power"`
	FirstPassedAt *time.Time `gorm:"column:first_passed_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (PlayerStage) TableName() string {
	return "player_stage"
}
