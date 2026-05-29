package model

import "time"

type GachaPool struct {
	ID         uint64     `gorm:"primaryKey;column:id"`
	Name       string     `gorm:"column:name"`
	CostItem   string     `gorm:"column:cost_item"`
	CostOne    uint32     `gorm:"column:cost_one"`
	CostTen    uint32     `gorm:"column:cost_ten"`
	PityLimit  uint32     `gorm:"column:pity_limit"`
	Status     uint8      `gorm:"column:status"`
	StartAt    *time.Time `gorm:"column:start_at"`
	EndAt      *time.Time `gorm:"column:end_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
}

func (GachaPool) TableName() string {
	return "gacha_pool"
}

type GachaPoolItem struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	PoolID    uint64    `gorm:"column:pool_id"`
	ItemType  string    `gorm:"column:item_type"`
	ItemID    uint64    `gorm:"column:item_id"`
	ItemCount uint32    `gorm:"column:item_count"`
	Quality   uint8     `gorm:"column:quality"`
	Weight    uint32    `gorm:"column:weight"`
	IsPity    uint8     `gorm:"column:is_pity"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (GachaPoolItem) TableName() string {
	return "gacha_pool_item"
}

type PlayerGachaState struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	PlayerID     uint64    `gorm:"column:player_id"`
	PoolID       uint64    `gorm:"column:pool_id"`
	PityCounter  uint32    `gorm:"column:pity_counter"`
	TotalDraw    uint32    `gorm:"column:total_draw"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (PlayerGachaState) TableName() string {
	return "player_gacha_state"
}

type GachaRecord struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	PlayerID  uint64    `gorm:"column:player_id"`
	PoolID    uint64    `gorm:"column:pool_id"`
	DrawNo    string    `gorm:"column:draw_no"`
	ItemType  string    `gorm:"column:item_type"`
	ItemID    uint64    `gorm:"column:item_id"`
	ItemCount uint32    `gorm:"column:item_count"`
	Quality   uint8     `gorm:"column:quality"`
	IsPity    uint8     `gorm:"column:is_pity"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (GachaRecord) TableName() string {
	return "gacha_record"
}