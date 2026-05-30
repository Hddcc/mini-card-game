package repository

import (
	"errors"

	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type GachaRepository struct {
	db *gorm.DB
}

func NewGachaRepository(db *gorm.DB) *GachaRepository {
	return &GachaRepository{db: db}
}

func (r *GachaRepository) FindPool(poolID uint64) (*model.GachaPool, error) {
	var pool model.GachaPool
	if err := r.db.Where("id = ?", poolID).First(&pool).Error; err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *GachaRepository) ListPoolItems(poolID uint64) ([]model.GachaPoolItem, error) {
	var items []model.GachaPoolItem
	err := r.db.Where("pool_id = ?", poolID).Find(&items).Error
	return items, err
}

func (r *GachaRepository) FindState(playerID uint64, poolID uint64) (*model.PlayerGachaState, error) {
	var state model.PlayerGachaState
	err := r.db.Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)}).
		Where("player_id = ? AND pool_id = ?", playerID, poolID).
		First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *GachaRepository) LockOrCreateState(tx *gorm.DB, playerID uint64, poolID uint64) (*model.PlayerGachaState, error) {
	var state model.PlayerGachaState
	err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ? AND pool_id = ?", playerID, poolID).
		First(&state).Error
	if err == nil {
		return &state, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	state = model.PlayerGachaState{
		PlayerID:    playerID,
		PoolID:      poolID,
		PityCounter: 0,
		TotalDraw:   0,
	}
	if err := tx.Create(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *GachaRepository) SaveState(tx *gorm.DB, state *model.PlayerGachaState) error {
	return tx.Save(state).Error
}

func (r *GachaRepository) CreateRecords(tx *gorm.DB, records []model.GachaRecord) error {
	return tx.Create(&records).Error
}
