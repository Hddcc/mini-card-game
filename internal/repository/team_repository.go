package repository

import (
	"mini-card-game/internal/model"

	"gorm.io/gorm"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *TeamRepository) DeleteByPlayer(tx *gorm.DB, playerID uint64) error {
	return tx.Where("player_id = ?", playerID).Delete(&model.PlayerTeam{}).Error
}

func (r *TeamRepository) CreatePlayerTeam(tx *gorm.DB, team *model.PlayerTeam) error {
	return tx.Create(team).Error
}

func (r *TeamRepository) ListByPlayer(playerID uint64) ([]model.PlayerTeam, error) {
	var teams []model.PlayerTeam
	err := r.db.Where("player_id = ?", playerID).Order("slot ASC").Find(&teams).Error
	return teams, err
}
