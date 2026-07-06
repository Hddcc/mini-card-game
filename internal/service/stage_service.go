package service

import (
	"errors"
	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"
	"time"

	"gorm.io/gorm"
)

type StageService struct {
	db          *gorm.DB
	stageRepo   *repository.StageRepository
	teamRepo    *repository.TeamRepository
	heroRepo    *repository.HeroRepository
	assetRepo   *repository.AssetRepository
	playerRepo  *repository.PlayerRepository
	taskService *TaskService
}

func NewStageService(
	db *gorm.DB,
	stageRepo *repository.StageRepository,
	teamRepo *repository.TeamRepository,
	heroRepo *repository.HeroRepository,
	assetRepo *repository.AssetRepository,
	playerRepo *repository.PlayerRepository,
	taskService *TaskService,
) *StageService {
	return &StageService{
		db:          db,
		stageRepo:   stageRepo,
		teamRepo:    teamRepo,
		heroRepo:    heroRepo,
		assetRepo:   assetRepo,
		playerRepo:  playerRepo,
		taskService: taskService,
	}
}

type StageFightResult struct {
	Win                bool   `json:"win"`
	CurrentPower       uint64 `json:"current_power"`
	RecommendPower     uint64 `json:"recommend_power"`
	RewardGold         uint64 `json:"reward_gold"`
	RewardExp          uint32 `json:"reward_exp"`
	Stamina            uint32 `json:"stamina"`
	MaxStamina         uint32 `json:"max_stamina"`
	NextStaminaSeconds int64  `json:"next_stamina_seconds"`
	BestPower          uint64 `json:"best_power"`
}

type StageProgressView struct {
	StageID       uint64     `json:"stage_id"`
	Status        uint8      `json:"status"`
	BestPower     uint64     `json:"best_power"`
	FirstPassedAt *time.Time `json:"first_passed_at,omitempty"`
}

var (
	ErrStageNotFound       = errors.New("stage not found")
	ErrPrevStageNotCleared = errors.New("previous stage is not cleared")
	ErrNoTeam              = errors.New("team not found")
	ErrNotEnoughStamina    = errors.New("not enough stamina")
)

func (s *StageService) Progress(playerID uint64) ([]StageProgressView, error) {
	rows, err := s.stageRepo.ListPlayerStages(playerID)
	if err != nil {
		return nil, err
	}
	progress := make([]StageProgressView, 0, len(rows))
	for _, row := range rows {
		progress = append(progress, StageProgressView{
			StageID:       row.StageID,
			Status:        row.Status,
			BestPower:     row.BestPower,
			FirstPassedAt: row.FirstPassedAt,
		})
	}
	return progress, nil
}

func (s *StageService) Fight(playerID uint64, stageID uint64) (*StageFightResult, error) {
	config, err := s.stageRepo.FindConfig(stageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStageNotFound
		}
		return nil, err
	}

	if config.PrevStageID > 0 {
		prevStage, err := s.stageRepo.FindPlayerStage(playerID, config.PrevStageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPrevStageNotCleared
			}
			return nil, err
		}
		if prevStage.Status != 1 {
			return nil, ErrPrevStageNotCleared
		}
	}

	teamRows, err := s.teamRepo.ListByPlayer(playerID)
	if err != nil {
		return nil, err
	}
	if len(teamRows) == 0 {
		return nil, ErrNoTeam
	}

	heroes, err := s.heroRepo.ListPlayerHeroes(playerID)
	if err != nil {
		return nil, err
	}
	heroMap := make(map[uint64]model.PlayerHero)
	for _, hero := range heroes {
		heroMap[hero.ID] = hero
	}

	configIDs := make([]uint64, 0, len(teamRows))
	configIDSet := make(map[uint64]struct{})
	var totalPower uint64
	for _, team := range teamRows {
		hero, ok := heroMap[team.PlayerHeroID]
		if !ok {
			return nil, errors.New("team contains invalid hero")
		}
		if _, ok := configIDSet[hero.HeroConfigID]; !ok {
			configIDSet[hero.HeroConfigID] = struct{}{}
			configIDs = append(configIDs, hero.HeroConfigID)
		}
	}

	configs, err := s.heroRepo.ListConfigsByIDs(configIDs)
	if err != nil {
		return nil, err
	}
	cfgMap := make(map[uint64]model.HeroConfig)
	for _, cfg := range configs {
		cfgMap[cfg.ID] = cfg
	}

	for _, team := range teamRows {
		hero := heroMap[team.PlayerHeroID]
		cfg, ok := cfgMap[hero.HeroConfigID]
		if !ok {
			return nil, errors.New("hero config missing")
		}
		totalPower += CalcHeroPower(cfg, hero.Level, hero.Star)
	}

	result := &StageFightResult{
		CurrentPower:   totalPower,
		RecommendPower: config.RecommendPower,
	}

	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
		if err != nil {
			return err
		}
		SettleStamina(asset, now)
		if asset.Stamina < config.StaminaCost {
			return ErrNotEnoughStamina
		}

		asset.Stamina -= config.StaminaCost
		TouchStaminaAfterSpend(asset, now)
		if totalPower >= config.RecommendPower {
			result.Win = true
			result.RewardGold = config.RewardGold
			result.RewardExp = config.RewardExp
			asset.Gold += config.RewardGold
		}

		if err := s.assetRepo.Save(tx, asset); err != nil {
			return err
		}

		if s.taskService != nil {
			if err := s.taskService.AddProgress(tx, playerID, "stage_fight", 1); err != nil {
				return err
			}
		}

		result.Stamina = asset.Stamina
		result.MaxStamina = MaxStamina
		result.NextStaminaSeconds = NextStaminaSeconds(asset, now)

		if result.Win {
			stageRow, err := s.stageRepo.LockPlayerStage(tx, playerID, stageID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					stageRow = &model.PlayerStage{
						PlayerID:  playerID,
						StageID:   stageID,
						Status:    1,
						BestPower: totalPower,
					}
				} else {
					return err
				}
			}

			stageRow.Status = 1
			if totalPower > stageRow.BestPower {
				stageRow.BestPower = totalPower
			}
			if err := s.stageRepo.SavePlayerStage(tx, stageRow); err != nil {
				return err
			}

			if s.taskService != nil {
				if err := s.taskService.AddProgress(tx, playerID, "stage_win", 1); err != nil {
					return err
				}
			}

			if err := s.playerRepo.AddProfileExp(tx, playerID, config.RewardExp); err != nil {
				return err
			}
			result.BestPower = stageRow.BestPower
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
