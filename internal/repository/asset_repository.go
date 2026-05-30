package repository

import (
	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) LockByPlayerID(tx *gorm.DB, playerID uint64) (*model.PlayerAsset, error) {
	var asset model.PlayerAsset
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ?", playerID).
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *AssetRepository) FindByPlayerID(playerID uint64) (*model.PlayerAsset, error) {
	var asset model.PlayerAsset
	if err := r.db.Where("player_id = ?", playerID).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *AssetRepository) Save(tx *gorm.DB, asset *model.PlayerAsset) error {
	return tx.Save(asset).Error
}
