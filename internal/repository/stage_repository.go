package repository

import (
	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StageRepository struct {
	db *gorm.DB
}

func NewStageRepository(db *gorm.DB) *StageRepository {
	return &StageRepository{db: db}
}

func (r *StageRepository) FindConfig(stageID uint64) (*model.StageConfig, error) {
	var config model.StageConfig
	if err := r.db.Where("id = ?", stageID).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *StageRepository) FindPlayerStage(playerID uint64, stageID uint64) (*model.PlayerStage, error) {
	var playerStage model.PlayerStage
	if err := r.db.Where("player_id = ? AND stage_id = ?", playerID, stageID).First(&playerStage).Error; err != nil {
		return nil, err
	}
	return &playerStage, nil
}

func (r *StageRepository) ListPlayerStages(playerID uint64) ([]model.PlayerStage, error) {
	var rows []model.PlayerStage
	err := r.db.Where("player_id = ?", playerID).Order("stage_id ASC").Find(&rows).Error
	return rows, err
}

func (r *StageRepository) LockPlayerStage(tx *gorm.DB, playerID uint64, stageID uint64) (*model.PlayerStage, error) {
	var playerStage model.PlayerStage
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ? AND stage_id = ?", playerID, stageID).
		First(&playerStage).Error
	if err != nil {
		return nil, err
	}
	return &playerStage, nil
}

func (r *StageRepository) SavePlayerStage(tx *gorm.DB, playerStage *model.PlayerStage) error {
	return tx.Save(playerStage).Error
}
