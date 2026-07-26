package model

import "time"

type PlayerProfile struct {
	ID                   uint64     `gorm:"primaryKey;column:id"`
	UserID               uint64     `gorm:"column:user_id"`
	Nickname             string     `gorm:"column:nickname;size:64"`
	Level                uint32     `gorm:"column:level"`
	Exp                  uint32     `gorm:"column:exp"`
	Avatar               string     `gorm:"column:avatar;size:128"`
	Power                uint64     `gorm:"column:power"`
	NameUpdatedAt        *time.Time `gorm:"column:name_updated_at"`
	NameChangeDate       *time.Time `gorm:"column:name_change_date;type:date"`
	NameDailyChangeCount uint32     `gorm:"column:name_daily_change_count"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (PlayerProfile) TableName() string {
	return "player_profiles"
}

type PlayerAsset struct {
	PlayerID         uint64     `gorm:"primaryKey;column:player_id"`
	Gold             uint64     `gorm:"column:gold"`
	Diamond          uint64     `gorm:"column:diamond"`
	Stamina          uint32     `gorm:"column:stamina"`
	StaminaUpdatedAt *time.Time `gorm:"column:stamina_updated_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (PlayerAsset) TableName() string {
	return "player_assets"
}
