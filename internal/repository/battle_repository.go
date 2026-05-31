package repository

import (
	"time"

	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BattleRepository struct {
	db *gorm.DB
}

func NewBattleRepository(db *gorm.DB) *BattleRepository {
	return &BattleRepository{db: db}
}

func (r *BattleRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *BattleRepository) ListStageEnemies(stageID uint64) ([]model.StageEnemyConfig, error) {
	var rows []model.StageEnemyConfig
	err := r.db.Where("stage_id = ?", stageID).Order("slot ASC").Find(&rows).Error
	return rows, err
}

func (r *BattleRepository) ListEnemyConfigsByIDs(ids []uint64) ([]model.EnemyConfig, error) {
	var configs []model.EnemyConfig
	if len(ids) == 0 {
		return configs, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&configs).Error
	return configs, err
}

func (r *BattleRepository) ListSkillConfigsByIDs(ids []uint64) ([]model.SkillConfig, error) {
	var configs []model.SkillConfig
	if len(ids) == 0 {
		return configs, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&configs).Error
	return configs, err
}

func (r *BattleRepository) ListCardSkins(ownerType string, ownerIDs []uint64) ([]model.CardSkinConfig, error) {
	var skins []model.CardSkinConfig
	if len(ownerIDs) == 0 {
		return skins, nil
	}
	err := r.db.Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIDs).Find(&skins).Error
	return skins, err
}

func (r *BattleRepository) ListEncounterVariants(stageID uint64) ([]model.StageEncounterVariant, error) {
	var variants []model.StageEncounterVariant
	err := r.db.Where("stage_id = ? AND status = ?", stageID, 1).Order("id ASC").Find(&variants).Error
	return variants, err
}

func (r *BattleRepository) ListEncounterEnemies(variantID uint64) ([]model.StageEncounterEnemy, error) {
	var rows []model.StageEncounterEnemy
	err := r.db.Where("variant_id = ?", variantID).Order("slot ASC").Find(&rows).Error
	return rows, err
}

func (r *BattleRepository) CreateSession(tx *gorm.DB, session *model.PlayerBattleSession) error {
	return tx.Create(session).Error
}

func (r *BattleRepository) SaveSession(tx *gorm.DB, session *model.PlayerBattleSession) error {
	return tx.Save(session).Error
}

func (r *BattleRepository) LockSession(tx *gorm.DB, playerID uint64, sessionID uint64) (*model.PlayerBattleSession, error) {
	var session model.PlayerBattleSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND player_id = ?", sessionID, playerID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *BattleRepository) LockActiveSession(tx *gorm.DB, playerID uint64, now time.Time) (*model.PlayerBattleSession, error) {
	var session model.PlayerBattleSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("player_id = ? AND status = ? AND expires_at > ?", playerID, model.BattleStatusActive, now).
		Order("id DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *BattleRepository) AbandonExpiredActiveSessions(tx *gorm.DB, playerID uint64, now time.Time) error {
	return tx.Model(&model.PlayerBattleSession{}).
		Where("player_id = ? AND status = ? AND expires_at <= ?", playerID, model.BattleStatusActive, now).
		Updates(map[string]interface{}{"status": model.BattleStatusAbandoned}).Error
}

func (r *BattleRepository) HasActiveSession(tx *gorm.DB, playerID uint64, now time.Time) (bool, error) {
	var count int64
	err := tx.Model(&model.PlayerBattleSession{}).
		Where("player_id = ? AND status = ? AND expires_at > ?", playerID, model.BattleStatusActive, now).
		Count(&count).Error
	return count > 0, err
}
