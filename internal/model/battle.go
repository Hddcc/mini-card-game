package model

import "time"

const (
	BattleStatusActive    = "active"
	BattleStatusWon       = "won"
	BattleStatusLost      = "lost"
	BattleStatusAbandoned = "abandoned"
)

type EnemyConfig struct {
	ID              uint64    `gorm:"primaryKey;column:id" json:"id"`
	Name            string    `gorm:"column:name;size:64" json:"name"`
	Role            string    `gorm:"column:role;size:32" json:"role"`
	BaseHP          uint32    `gorm:"column:base_hp" json:"base_hp"`
	BaseATK         uint32    `gorm:"column:base_atk" json:"base_atk"`
	BaseDEF         uint32    `gorm:"column:base_def" json:"base_def"`
	SkillID         uint64    `gorm:"column:skill_id" json:"skill_id"`
	CardArt         string    `gorm:"column:card_art;size:255" json:"card_art"`
	PortraitArt     string    `gorm:"column:portrait_art;size:255" json:"portrait_art"`
	AttackAnimation string    `gorm:"column:attack_animation;size:64" json:"attack_animation"`
	SkillAnimation  string    `gorm:"column:skill_animation;size:64" json:"skill_animation"`
	HitAnimation    string    `gorm:"column:hit_animation;size:64" json:"hit_animation"`
	DefeatAnimation string    `gorm:"column:defeat_animation;size:64" json:"defeat_animation"`
	IdleAnimation   string    `gorm:"column:idle_animation;size:64" json:"idle_animation"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (EnemyConfig) TableName() string {
	return "enemy_config"
}

type StageEnemyConfig struct {
	ID            uint64    `gorm:"primaryKey;column:id" json:"id"`
	StageID       uint64    `gorm:"column:stage_id" json:"stage_id"`
	EnemyConfigID uint64    `gorm:"column:enemy_config_id" json:"enemy_config_id"`
	Slot          uint8     `gorm:"column:slot" json:"slot"`
	Level         uint32    `gorm:"column:level" json:"level"`
	Count         uint32    `gorm:"column:count" json:"count"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (StageEnemyConfig) TableName() string {
	return "stage_enemy_config"
}

type SkillConfig struct {
	ID             uint64    `gorm:"primaryKey;column:id" json:"id"`
	Code           string    `gorm:"column:code;size:64" json:"code"`
	Name           string    `gorm:"column:name;size:64" json:"name"`
	TargetType     string    `gorm:"column:target_type;size:32" json:"target_type"`
	EffectType     string    `gorm:"column:effect_type;size:32" json:"effect_type"`
	Multiplier     uint32    `gorm:"column:multiplier" json:"multiplier"`
	CostRage       uint32    `gorm:"column:cost_rage" json:"cost_rage"`
	Cooldown       uint32    `gorm:"column:cooldown" json:"cooldown"`
	DurationRounds uint32    `gorm:"column:duration_rounds" json:"duration_rounds"`
	StatDelta      int32     `gorm:"column:stat_delta" json:"stat_delta"`
	EffectKey      string    `gorm:"column:effect_key;size:64" json:"effect_key"`
	AnimationKey   string    `gorm:"column:animation_key;size:64" json:"animation_key"`
	Description    string    `gorm:"column:description;size:255" json:"description"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SkillConfig) TableName() string {
	return "skill_config"
}

type PlayerBattleSession struct {
	ID         uint64     `gorm:"primaryKey;column:id" json:"id"`
	PlayerID   uint64     `gorm:"column:player_id" json:"player_id"`
	StageID    uint64     `gorm:"column:stage_id" json:"stage_id"`
	Status     string     `gorm:"column:status;size:24" json:"status"`
	Round      uint32     `gorm:"column:round" json:"round"`
	StateJSON  string     `gorm:"column:state_json;type:longtext" json:"state_json"`
	ResultJSON string     `gorm:"column:result_json;type:longtext" json:"result_json"`
	ExpiresAt  time.Time  `gorm:"column:expires_at" json:"expires_at"`
	SettledAt  *time.Time `gorm:"column:settled_at" json:"settled_at"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (PlayerBattleSession) TableName() string {
	return "player_battle_session"
}

type CardSkinConfig struct {
	ID              uint64    `gorm:"primaryKey;column:id" json:"id"`
	OwnerType       string    `gorm:"column:owner_type;size:16" json:"owner_type"`
	OwnerID         uint64    `gorm:"column:owner_id" json:"owner_id"`
	CardArt         string    `gorm:"column:card_art;size:255" json:"card_art"`
	PortraitArt     string    `gorm:"column:portrait_art;size:255" json:"portrait_art"`
	AttackAnimation string    `gorm:"column:attack_animation;size:64" json:"attack_animation"`
	SkillAnimation  string    `gorm:"column:skill_animation;size:64" json:"skill_animation"`
	HitAnimation    string    `gorm:"column:hit_animation;size:64" json:"hit_animation"`
	DefeatAnimation string    `gorm:"column:defeat_animation;size:64" json:"defeat_animation"`
	IdleAnimation   string    `gorm:"column:idle_animation;size:64" json:"idle_animation"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CardSkinConfig) TableName() string {
	return "card_skin_config"
}

type StageEncounterVariant struct {
	ID             uint64    `gorm:"primaryKey;column:id" json:"id"`
	StageID        uint64    `gorm:"column:stage_id" json:"stage_id"`
	Name           string    `gorm:"column:name;size:64" json:"name"`
	MinPower       uint64    `gorm:"column:min_power" json:"min_power"`
	MaxPower       uint64    `gorm:"column:max_power" json:"max_power"`
	EstimatedPower uint64    `gorm:"column:estimated_power" json:"estimated_power"`
	Weight         uint32    `gorm:"column:weight" json:"weight"`
	Status         uint8     `gorm:"column:status" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (StageEncounterVariant) TableName() string {
	return "stage_encounter_variant"
}

type StageEncounterEnemy struct {
	ID            uint64    `gorm:"primaryKey;column:id" json:"id"`
	VariantID     uint64    `gorm:"column:variant_id" json:"variant_id"`
	EnemyConfigID uint64    `gorm:"column:enemy_config_id" json:"enemy_config_id"`
	Slot          uint8     `gorm:"column:slot" json:"slot"`
	Level         uint32    `gorm:"column:level" json:"level"`
	Count         uint32    `gorm:"column:count" json:"count"`
	SkillID       uint64    `gorm:"column:skill_id" json:"skill_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (StageEncounterEnemy) TableName() string {
	return "stage_encounter_enemy"
}
