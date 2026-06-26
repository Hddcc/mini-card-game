package repository

import (
	"time"

	"mini-card-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActivityLotteryRepository struct {
	db *gorm.DB
}

func NewActivityLotteryRepository(db *gorm.DB) *ActivityLotteryRepository {
	return &ActivityLotteryRepository{db: db}
}

func (r *ActivityLotteryRepository) FindActive(now time.Time) (*model.ActivityLottery, error) {
	var activity model.ActivityLottery
	err := r.db.Where("status = ? AND start_at <= ? AND end_at >= ?", 1, now, now).
		Order("id ASC").
		First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityLotteryRepository) ListActivePrizes(activityID uint64) ([]model.ActivityLotteryPrize, error) {
	var prizes []model.ActivityLotteryPrize
	err := r.db.Where("activity_id = ? AND status = ?", activityID, 1).
		Order("display_order ASC, id ASC").
		Find(&prizes).Error
	return prizes, err
}

func (r *ActivityLotteryRepository) DecrementPrizeInventory(tx *gorm.DB, prizeID uint64, count int64) (bool, error) {
	res := tx.Model(&model.ActivityLotteryPrize{}).
		Where("id = ? AND (total_num < 0 OR left_num >= ?)", prizeID, count).
		UpdateColumn("left_num", gorm.Expr("CASE WHEN total_num < 0 THEN left_num ELSE left_num - ? END", count))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *ActivityLotteryRepository) IncrementPrizeInventory(tx *gorm.DB, prizeID uint64, count int64) error {
	return tx.Model(&model.ActivityLotteryPrize{}).
		Where("id = ? AND total_num >= 0", prizeID).
		UpdateColumn("left_num", gorm.Expr("left_num + ?", count)).Error
}

func (r *ActivityLotteryRepository) CreateRecord(tx *gorm.DB, record *model.ActivityLotteryRecord) error {
	return tx.Create(record).Error
}

func (r *ActivityLotteryRepository) SaveRecord(tx *gorm.DB, record *model.ActivityLotteryRecord) error {
	return tx.Save(record).Error
}

func (r *ActivityLotteryRepository) FindRecordForUpdate(tx *gorm.DB, drawNo string) (*model.ActivityLotteryRecord, error) {
	var record model.ActivityLotteryRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("draw_no = ?", drawNo).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ActivityLotteryRepository) ListPlayerRecords(playerID uint64, limit int) ([]model.ActivityLotteryRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	var records []model.ActivityLotteryRecord
	err := r.db.Where("player_id = ?", playerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *ActivityLotteryRepository) CountPlayerRecordsOnDay(playerID uint64, activityID uint64, dayStart time.Time, dayEnd time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.ActivityLotteryRecord{}).
		Where("player_id = ? AND activity_id = ? AND created_at >= ? AND created_at < ?", playerID, activityID, dayStart, dayEnd).
		Count(&count).Error
	return count, err
}

func (r *ActivityLotteryRepository) CountIPRecordsOnDay(ip string, activityID uint64, dayStart time.Time, dayEnd time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.ActivityLotteryRecord{}).
		Where("request_ip = ? AND activity_id = ? AND created_at >= ? AND created_at < ?", ip, activityID, dayStart, dayEnd).
		Count(&count).Error
	return count, err
}

func (r *ActivityLotteryRepository) FindActiveBlacklist(targetType string, target string, now time.Time) (*model.ActivityLotteryBlacklist, error) {
	var item model.ActivityLotteryBlacklist
	err := r.db.Where("target_type = ? AND target = ? AND status = ? AND (expire_at IS NULL OR expire_at > ?)", targetType, target, 1, now).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ActivityLotteryRepository) CreateLocalMessage(tx *gorm.DB, message *model.ActivityLotteryLocalMessage) error {
	return tx.Create(message).Error
}

func (r *ActivityLotteryRepository) SaveLocalMessage(tx *gorm.DB, message *model.ActivityLotteryLocalMessage) error {
	return tx.Save(message).Error
}

func (r *ActivityLotteryRepository) FindLocalMessageForUpdate(tx *gorm.DB, businessID string) (*model.ActivityLotteryLocalMessage, error) {
	var message model.ActivityLotteryLocalMessage
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("business_id = ?", businessID).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *ActivityLotteryRepository) ListRetryableMessages(now time.Time, limit int) ([]model.ActivityLotteryLocalMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	var messages []model.ActivityLotteryLocalMessage
	err := r.db.Where("status IN ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", []int{0, 2}, now).
		Order("updated_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *ActivityLotteryRepository) UpsertReleaseState(tx *gorm.DB, prizeID uint64, windowKey string, released int64) error {
	state := model.ActivityPrizeReleaseState{
		PrizeID:   prizeID,
		WindowKey: windowKey,
		Released:  released,
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "prize_id"}, {Name: "window_key"}},
		DoNothing: true,
	}).Create(&state).Error
}

func (r *ActivityLotteryRepository) FindReleaseState(prizeID uint64, windowKey string) (*model.ActivityPrizeReleaseState, error) {
	var state model.ActivityPrizeReleaseState
	err := r.db.Where("prize_id = ? AND window_key = ?", prizeID, windowKey).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}
