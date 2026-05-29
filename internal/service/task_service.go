package service

import (
	"errors"
	"time"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type TaskService struct {
	db            *gorm.DB
	taskRepo      *repository.TaskRepository
	rewardService *RewardService
}

type TaskView struct {
	TaskID        uint64     `json:"task_id"`
	Name          string     `json:"name"`
	EventType     string     `json:"event_type"`
	Progress      uint32     `json:"progress"`
	TargetCount   uint32     `json:"target_count"`
	Status        uint8      `json:"status"`
	RewardGold    uint64     `json:"reward_gold"`
	RewardDiamond uint64     `json:"reward_diamond"`
	ClaimedAt     *time.Time `json:"claimed_at"`
}

func NewTaskService(db *gorm.DB, taskRepo *repository.TaskRepository, rewardService *RewardService) *TaskService {
	return &TaskService{
		db:            db,
		taskRepo:      taskRepo,
		rewardService: rewardService,
	}
}

func todayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func (s *TaskService) ListDaily(playerID uint64) ([]TaskView, error) {
	configs, err := s.taskRepo.ListActiveConfigs()
	if err != nil {
		return nil, err
	}

	today := todayDate()
	tasks, err := s.taskRepo.ListPlayerDailyTasks(playerID, today)
	if err != nil {
		return nil, err
	}

	existing := make(map[uint64]model.PlayerDailyTask)
	for _, task := range tasks {
		existing[task.TaskID] = task
	}

	if len(existing) < len(configs) {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			tasks, err = s.taskRepo.ListPlayerDailyTasks(playerID, today)
			if err != nil {
				return err
			}
			existing = make(map[uint64]model.PlayerDailyTask)
			for _, task := range tasks {
				existing[task.TaskID] = task
			}
			for _, cfg := range configs {
				if _, ok := existing[cfg.ID]; ok {
					continue
				}
				task := &model.PlayerDailyTask{
					PlayerID: playerID,
					TaskID:   cfg.ID,
					TaskDate: today,
					Progress: 0,
					Status:   0,
				}
				if err := s.taskRepo.CreatePlayerDailyTask(tx, task); err != nil {
					return err
				}
				existing[cfg.ID] = *task
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		tasks, err = s.taskRepo.ListPlayerDailyTasks(playerID, today)
		if err != nil {
			return nil, err
		}
	}

	return s.buildTaskViews(tasks, configs), nil
}

func (s *TaskService) buildTaskViews(tasks []model.PlayerDailyTask, configs []model.DailyTaskConfig) []TaskView {
	cfgMap := make(map[uint64]model.DailyTaskConfig)
	for _, cfg := range configs {
		cfgMap[cfg.ID] = cfg
	}

	views := make([]TaskView, 0, len(tasks))
	for _, task := range tasks {
		cfg, ok := cfgMap[task.TaskID]
		if !ok {
			continue
		}
		views = append(views, TaskView{
			TaskID:        task.TaskID,
			Name:          cfg.Name,
			EventType:     cfg.EventType,
			Progress:      task.Progress,
			TargetCount:   cfg.TargetCount,
			Status:        task.Status,
			RewardGold:    cfg.RewardGold,
			RewardDiamond: cfg.RewardDiamond,
			ClaimedAt:     task.ClaimedAt,
		})
	}
	return views
}

func (s *TaskService) AddProgress(tx *gorm.DB, playerID uint64, eventType string, add uint32) error {
	if add == 0 {
		return nil
	}

	today := todayDate()
	configs, err := s.taskRepo.ListConfigsByEventType(eventType)
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		return nil
	}

	tasks, err := s.taskRepo.ListPlayerDailyTasksByEventType(playerID, today, eventType)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		if _, err := s.ListDaily(playerID); err != nil {
			return err
		}
		tasks, err = s.taskRepo.ListPlayerDailyTasksByEventType(playerID, today, eventType)
		if err != nil {
			return err
		}
	}

	cfgMap := make(map[uint64]model.DailyTaskConfig)
	for _, cfg := range configs {
		cfgMap[cfg.ID] = cfg
	}

	for _, task := range tasks {
		if task.Status == 2 {
			continue
		}
		cfg, ok := cfgMap[task.TaskID]
		if !ok {
			continue
		}
		if task.Progress >= cfg.TargetCount {
			if task.Status != 1 {
				task.Status = 1
				if err := s.taskRepo.SavePlayerDailyTask(tx, &task); err != nil {
					return err
				}
			}
			continue
		}

		progress := task.Progress + add
		if progress >= cfg.TargetCount {
			progress = cfg.TargetCount
			task.Status = 1
		}
		task.Progress = progress
		if err := s.taskRepo.SavePlayerDailyTask(tx, &task); err != nil {
			return err
		}
	}

	return nil
}

func (s *TaskService) Claim(playerID uint64, taskID uint64) ([]model.Reward, error) {
	var rewards []model.Reward
	today := todayDate()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		task, err := s.taskRepo.FindPlayerDailyTask(tx, playerID, taskID, today)
		if err != nil {
			return err
		}
		if task.Status == 2 {
			return errors.New("task already claimed")
		}

		cfg, err := s.taskRepo.FindConfig(task.TaskID)
		if err != nil {
			return err
		}
		if task.Progress < cfg.TargetCount {
			return errors.New("task not completed")
		}

		if cfg.RewardGold > 0 {
			rewards = append(rewards, model.Reward{Type: "gold", Count: cfg.RewardGold})
		}
		if cfg.RewardDiamond > 0 {
			rewards = append(rewards, model.Reward{Type: "diamond", Count: cfg.RewardDiamond})
		}

		if len(rewards) > 0 {
			if err := s.rewardService.Grant(tx, playerID, rewards); err != nil {
				return err
			}
		}

		now := time.Now()
		task.Status = 2
		task.ClaimedAt = &now
		if err := s.taskRepo.SavePlayerDailyTask(tx, task); err != nil {
			return err
		}
		return nil
	})

	return rewards, err
}
