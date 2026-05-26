package model

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
