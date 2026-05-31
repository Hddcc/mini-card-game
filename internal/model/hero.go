package model

import "time"

type HeroConfig struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;size:64" json:"name"`
	Quality     uint8     `gorm:"column:quality" json:"quality"`
	Role        string    `gorm:"column:role;size:32" json:"role"`
	BaseHP      uint32    `gorm:"column:base_hp" json:"base_hp"`
	BaseATK     uint32    `gorm:"column:base_atk" json:"base_atk"`
	BaseDEF     uint32    `gorm:"column:base_def" json:"base_def"`
	PowerFactor uint32    `gorm:"column:power_factor" json:"power_factor"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (HeroConfig) TableName() string {
	return "hero_config"
}

type PlayerHero struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	PlayerID     uint64    `gorm:"column:player_id" json:"player_id"`
	HeroConfigID uint64    `gorm:"column:hero_config_id" json:"hero_config_id"`
	Level        uint32    `gorm:"column:level" json:"level"`
	Star         uint32    `gorm:"column:star" json:"star"`
	Shard        uint32    `gorm:"column:shard" json:"shard"`
	Locked       uint8     `gorm:"column:locked" json:"locked"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PlayerHero) TableName() string {
	return "player_hero"
}
