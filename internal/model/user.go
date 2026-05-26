package model

import "time"

type User struct {
	ID           uint64     `gorm:"primaryKey;column:id"`
	Username     string     `gorm:"column:username"`
	PasswordHash string     `gorm:"column:password_hash"`
	Status       int        `gorm:"column:status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
