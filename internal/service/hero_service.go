package service

import (
	"errors"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type HeroService struct {
	db        *gorm.DB
	heroRepo  *repository.HeroRepository
	assetRepo *repository.AssetRepository
}

func NewHeroService(db *gorm.DB, heroRepo *repository.HeroRepository, assetRepo *repository.AssetRepository) *HeroService {
	return &HeroService{db: db, heroRepo: heroRepo, assetRepo: assetRepo}
}

type HeroView struct {
	PlayerHeroID uint64      `json:"player_hero_id"`
	ID           uint64      `json:"id"`
	HeroConfigID uint64      `json:"hero_config_id"`
	Name         string      `json:"name"`
	Quality      uint8       `json:"quality"`
	Role         string      `json:"role"`
	Level        uint32      `json:"level"`
	Star         uint32      `json:"star"`
	Shard        uint32      `json:"shard"`
	MaxStar      bool        `json:"max_star"`
	NextStarCost *StarUpCost `json:"next_star_cost,omitempty"`
}

type StarUpOutput struct {
	Hero   HeroView          `json:"hero"`
	Assets model.PlayerAsset `json:"assets"`
}

var (
	ErrHeroNotOwned        = errors.New("hero not owned")
	ErrHeroMaxStar         = errors.New("hero already max star")
	ErrHeroShardNotEnough  = errors.New("hero shard not enough")
	ErrHeroGoldNotEnough   = errors.New("gold not enough")
	ErrHeroStarCostMissing = errors.New("hero star cost missing")
)

func (s *HeroService) List(playerID uint64) ([]HeroView, error) {
	heroes, err := s.heroRepo.ListPlayerHeroes(playerID)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(heroes))
	for _, hero := range heroes {
		ids = append(ids, hero.HeroConfigID)
	}

	configs, err := s.heroRepo.ListConfigsByIDs(ids)
	if err != nil {
		return nil, err
	}

	configMap := make(map[uint64]model.HeroConfig)
	for _, cfg := range configs {
		configMap[cfg.ID] = cfg
	}

	result := make([]HeroView, 0, len(heroes))
	for _, hero := range heroes {
		result = append(result, buildHeroView(hero, configMap[hero.HeroConfigID]))
	}
	return result, nil
}

func (s *HeroService) StarUp(playerID uint64, playerHeroID uint64) (*StarUpOutput, error) {
	var output *StarUpOutput
	err := s.db.Transaction(func(tx *gorm.DB) error {
		hero, err := s.heroRepo.LockPlayerHeroByID(tx, playerID, playerHeroID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrHeroNotOwned
			}
			return err
		}
		config, err := s.heroRepo.FindConfig(hero.HeroConfigID)
		if err != nil {
			return err
		}
		cost, ok := NextStarUpCost(hero.Star)
		if !ok {
			if hero.Star >= MaxHeroStar {
				return ErrHeroMaxStar
			}
			return ErrHeroStarCostMissing
		}
		if hero.Shard < cost.Shard {
			return ErrHeroShardNotEnough
		}
		asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
		if err != nil {
			return err
		}
		if asset.Gold < cost.Gold {
			return ErrHeroGoldNotEnough
		}

		hero.Shard -= cost.Shard
		hero.Star++
		asset.Gold -= cost.Gold
		if err := s.heroRepo.SavePlayerHero(tx, hero); err != nil {
			return err
		}
		if err := s.assetRepo.Save(tx, asset); err != nil {
			return err
		}
		output = &StarUpOutput{
			Hero:   buildHeroView(*hero, *config),
			Assets: *asset,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func buildHeroView(hero model.PlayerHero, cfg model.HeroConfig) HeroView {
	view := HeroView{
		PlayerHeroID: hero.ID,
		ID:           hero.ID,
		HeroConfigID: hero.HeroConfigID,
		Name:         cfg.Name,
		Quality:      cfg.Quality,
		Role:         cfg.Role,
		Level:        hero.Level,
		Star:         hero.Star,
		Shard:        hero.Shard,
		MaxStar:      hero.Star >= MaxHeroStar,
	}
	if cost, ok := NextStarUpCost(hero.Star); ok {
		view.NextStarCost = &cost
	}
	return view
}
