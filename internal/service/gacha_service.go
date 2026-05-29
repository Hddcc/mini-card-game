package service

import (
	"errors"
	"fmt"
	"time"

	"mini-card-game/internal/model"
	"mini-card-game/internal/pkg/random"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type GachaService struct {
	db            *gorm.DB
	assetRepo     *repository.AssetRepository
	gachaRepo     *repository.GachaRepository
	rewardService *RewardService
}

func NewGachaService(db *gorm.DB, assetRepo *repository.AssetRepository, gachaRepo *repository.GachaRepository, rewardService *RewardService) *GachaService {
	return &GachaService{
		db:            db,
		assetRepo:     assetRepo,
		gachaRepo:     gachaRepo,
		rewardService: rewardService,
	}
}

type DrawResult struct {
	ItemType  string `json:"item_type"`
	ItemID    uint64 `json:"item_id"`
	ItemCount uint32 `json:"item_count"`
	Quality   uint8  `json:"quality"`
	IsPity    bool   `json:"is_pity"`
}

type DrawOutput struct {
	DrawNo      string       `json:"draw_no"`
	Results     []DrawResult `json:"results"`
	Diamond     uint64       `json:"diamond"`
	PityCounter uint32       `json:"pity_counter"`
}

func (s *GachaService) Draw(playerID uint64, poolID uint64, times int) (*DrawOutput, error) {
	if times != 1 && times != 10 {
		return nil, errors.New("times must be 1 or 10")
	}

	pool, err := s.gachaRepo.FindPool(poolID)
	if err != nil {
		return nil, err
	}
	if pool.Status != 1 {
		return nil, errors.New("gacha pool closed")
	}

	items, err := s.gachaRepo.ListPoolItems(poolID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("gacha pool item empty")
	}

	cost := uint64(pool.CostOne)
	if times == 10 {
		cost = uint64(pool.CostTen)
	}

	drawNo := fmt.Sprintf("G%d%d", time.Now().UnixNano(), playerID)
	output := &DrawOutput{DrawNo: drawNo}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
		if err != nil {
			return err
		}
		if asset.Diamond < cost {
			return errors.New("diamond not enough")
		}

		asset.Diamond -= cost
		if err := s.assetRepo.Save(tx, asset); err != nil {
			return err
		}

		state, err := s.gachaRepo.LockOrCreateState(tx, playerID, poolID)
		if err != nil {
			return err
		}

		records := make([]model.GachaRecord, 0, times)
		results := make([]DrawResult, 0, times)

		for i := 0; i < times; i++ {
			item, isPity, err := s.pickItem(items, state.PityCounter+1 >= pool.PityLimit)
			if err != nil {
				return err
			}

			if item.Quality >= 5 {
				state.PityCounter = 0
			} else {
				state.PityCounter++
			}
			state.TotalDraw++

			reward := model.Reward{
				Type:  item.ItemType,
				ID:    item.ItemID,
				Count: uint64(item.ItemCount),
			}
			if err := s.rewardService.Grant(tx, playerID, []model.Reward{reward}); err != nil {
				return err
			}

			results = append(results, DrawResult{
				ItemType:  item.ItemType,
				ItemID:    item.ItemID,
				ItemCount: item.ItemCount,
				Quality:   item.Quality,
				IsPity:    isPity,
			})

			record := model.GachaRecord{
				PlayerID:  playerID,
				PoolID:    poolID,
				DrawNo:    drawNo,
				ItemType:  item.ItemType,
				ItemID:    item.ItemID,
				ItemCount: item.ItemCount,
				Quality:   item.Quality,
			}
			if isPity {
				record.IsPity = 1
			}
			records = append(records, record)
		}

		if err := s.gachaRepo.SaveState(tx, state); err != nil {
			return err
		}
		if err := s.gachaRepo.CreateRecords(tx, records); err != nil {
			return err
		}

		output.Results = results
		output.Diamond = asset.Diamond
		output.PityCounter = state.PityCounter
		return nil
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *GachaService) pickItem(items []model.GachaPoolItem, usePity bool) (model.GachaPoolItem, bool, error) {
	candidates := items
	if usePity {
		candidates = make([]model.GachaPoolItem, 0)
		for _, item := range items {
			if item.IsPity == 1 {
				candidates = append(candidates, item)
			}
		}
	}

	weights := make([]uint32, 0, len(candidates))
	for _, item := range candidates {
		weights = append(weights, item.Weight)
	}

	index := random.WeightIndex(weights)
	if index < 0 {
		return model.GachaPoolItem{}, false, errors.New("invalid gacha weight")
	}

	return candidates[index], usePity, nil
}
