package repository

import (
	"mini-card-game/internal/model"

	"gorm.io/gorm"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) CreateProfile(tx *gorm.DB, profile *model.PlayerProfile) error {
	return tx.Create(profile).Error
}

func (r *PlayerRepository) CreateAsset(tx *gorm.DB, asset *model.PlayerAsset) error {
	return tx.Create(asset).Error
}

func (r *PlayerRepository) FindProfileByUserID(userID uint64) (*model.PlayerProfile, error) {
	var profile model.PlayerProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *PlayerRepository) FindProfile(playerID uint64) (*model.PlayerProfile, error) {
	var profile model.PlayerProfile
	if err := r.db.Where("id = ?", playerID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *PlayerRepository) FindAsset(playerID uint64) (*model.PlayerAsset, error) {
	var asset model.PlayerAsset
	if err := r.db.Where("player_id = ?", playerID).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *PlayerRepository) SaveAsset(asset *model.PlayerAsset) error {
	return r.db.Save(asset).Error
}

func (r *PlayerRepository) UpdateProfilePower(tx *gorm.DB, playerID uint64, power uint64) error {
	return tx.Model(&model.PlayerProfile{}).Where("id = ?", playerID).Update("power", power).Error
}
func (r *PlayerRepository) AddProfileExp(tx *gorm.DB, playerID uint64, exp uint32) error {
	return tx.Model(&model.PlayerProfile{}).
		Where("id = ?", playerID).
		Update("exp", gorm.Expr("exp + ?", exp)).Error
}
