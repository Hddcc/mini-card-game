package model

import "time"

type PlayerTeam struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	PlayerID     uint64    `gorm:"column:player_id"`
	Slot         uint8     `gorm:"column:slot"`
	PlayerHeroID uint64    `gorm:"column:player_hero_id"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (PlayerTeam) TableName() string {
	return "player_team"
}
