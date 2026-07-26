package repository

import (
	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HeroRepository struct {
	db *gorm.DB
}

func NewHeroRepository(db *gorm.DB) *HeroRepository {
	return &HeroRepository{db: db}
}

func (r *HeroRepository) ListPlayerHeroes(playerID uint64) ([]model.PlayerHero, error) {
	var heroes []model.PlayerHero
	err := r.db.Where("player_id = ?", playerID).Find(&heroes).Error
	return heroes, err
}

func (r *HeroRepository) ListConfigsByIDs(ids []uint64) ([]model.HeroConfig, error) {
	var configs []model.HeroConfig
	if len(ids) == 0 {
		return configs, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&configs).Error
	return configs, err
}

func (r *HeroRepository) FindConfig(id uint64) (*model.HeroConfig, error) {
	var config model.HeroConfig
	if err := r.db.Where("id = ?", id).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *HeroRepository) LockPlayerHero(tx *gorm.DB, playerID uint64, heroConfigID uint64) (*model.PlayerHero, error) {
	var hero model.PlayerHero
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ? AND hero_config_id = ?", playerID, heroConfigID).
		First(&hero).Error
	if err != nil {
		return nil, err
	}
	return &hero, nil
}

func (r *HeroRepository) LockPlayerHeroByID(tx *gorm.DB, playerID uint64, playerHeroID uint64) (*model.PlayerHero, error) {
	var hero model.PlayerHero
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ? AND id = ?", playerID, playerHeroID).
		First(&hero).Error
	if err != nil {
		return nil, err
	}
	return &hero, nil
}

func (r *HeroRepository) CreatePlayerHero(tx *gorm.DB, hero *model.PlayerHero) error {
	return tx.Create(hero).Error
}

func (r *HeroRepository) CreatePlayerHeroIfMissing(tx *gorm.DB, hero *model.PlayerHero) error {
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(hero).Error
}

func (r *HeroRepository) SavePlayerHero(tx *gorm.DB, hero *model.PlayerHero) error {
	return tx.Save(hero).Error
}
