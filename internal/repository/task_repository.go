package repository

import (
	"time"

	"mini-card-game/internal/model"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) ListActiveConfigs() ([]model.DailyTaskConfig, error) {
	var configs []model.DailyTaskConfig
	err := r.db.Where("status = ?", 1).Find(&configs).Error
	return configs, err
}

func (r *TaskRepository) ListConfigsByEventType(eventType string) ([]model.DailyTaskConfig, error) {
	var configs []model.DailyTaskConfig
	err := r.db.Where("event_type = ? AND status = ?", eventType, 1).Find(&configs).Error
	return configs, err
}

func (r *TaskRepository) ListPlayerDailyTasks(playerID uint64, date time.Time) ([]model.PlayerDailyTask, error) {
	var tasks []model.PlayerDailyTask
	err := r.db.Where("player_id = ? AND DATE(task_date) = ?", playerID, date.Format("2006-01-02")).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) ListPlayerDailyTasksByEventType(playerID uint64, date time.Time, eventType string) ([]model.PlayerDailyTask, error) {
	var tasks []model.PlayerDailyTask
	err := r.db.Table("player_daily_task AS pdt").
		Select("pdt.*").
		Joins("JOIN daily_task_config dtc ON dtc.id = pdt.task_id").
		Where("pdt.player_id = ? AND DATE(pdt.task_date) = ? AND dtc.event_type = ? AND dtc.status = ?", playerID, date.Format("2006-01-02"), eventType, 1).
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) FindPlayerDailyTask(tx *gorm.DB, playerID uint64, taskID uint64, date time.Time) (*model.PlayerDailyTask, error) {
	var task model.PlayerDailyTask
	err := tx.Where("player_id = ? AND task_id = ? AND DATE(task_date) = ?", playerID, taskID, date.Format("2006-01-02")).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindConfig(taskID uint64) (*model.DailyTaskConfig, error) {
	var cfg model.DailyTaskConfig
	if err := r.db.Where("id = ?", taskID).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *TaskRepository) CreatePlayerDailyTask(tx *gorm.DB, task *model.PlayerDailyTask) error {
	return tx.Create(task).Error
}

func (r *TaskRepository) SavePlayerDailyTask(tx *gorm.DB, task *model.PlayerDailyTask) error {
	return tx.Save(task).Error
}
