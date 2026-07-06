package service

import (
	"errors"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type RewardService struct {
	assetRepo *repository.AssetRepository
	heroRepo  *repository.HeroRepository
}

type GrantResult struct {
	Type            string `json:"type"`
	ID              uint64 `json:"id"`
	Count           uint64 `json:"count"`
	IsDuplicate     bool   `json:"is_duplicate"`
	ConvertedShards uint32 `json:"converted_shards"`
}

func NewRewardService(assetRepo *repository.AssetRepository, heroRepo *repository.HeroRepository) *RewardService {
	return &RewardService{assetRepo: assetRepo, heroRepo: heroRepo}
}

func (s *RewardService) Grant(tx *gorm.DB, playerID uint64, rewards []model.Reward) error {
	_, err := s.GrantWithResults(tx, playerID, rewards)
	return err
}

func (s *RewardService) GrantWithResults(tx *gorm.DB, playerID uint64, rewards []model.Reward) ([]GrantResult, error) {
	results := make([]GrantResult, 0, len(rewards))
	for _, reward := range rewards {
		result := GrantResult{Type: reward.Type, ID: reward.ID, Count: reward.Count}
		switch reward.Type {
		case "gold":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return nil, err
			}
			asset.Gold += reward.Count
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return nil, err
			}
		case "diamond":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return nil, err
			}
			asset.Diamond += reward.Count
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return nil, err
			}
		case "stamina":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return nil, err
			}
			asset.Stamina += uint32(reward.Count)
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return nil, err
			}
		case "hero":
			heroResult, err := s.grantHero(tx, playerID, reward.ID)
			if err != nil {
				return nil, err
			}
			result.IsDuplicate = heroResult.IsDuplicate
			result.ConvertedShards = heroResult.ConvertedShards
		default:
			return nil, errors.New("unknown reward type")
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *RewardService) grantHero(tx *gorm.DB, playerID uint64, heroConfigID uint64) (GrantResult, error) {
	hero, err := s.heroRepo.LockPlayerHero(tx, playerID, heroConfigID)
	if err == nil {
		config, err := s.heroRepo.FindConfig(heroConfigID)
		if err != nil {
			return GrantResult{}, err
		}
		shards := DuplicateHeroShardAmount(config.Quality)
		hero.Shard += shards
		return GrantResult{
			Type:            "hero",
			ID:              heroConfigID,
			Count:           1,
			IsDuplicate:     true,
			ConvertedShards: shards,
		}, s.heroRepo.SavePlayerHero(tx, hero)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return GrantResult{}, err
	}

	if err := s.heroRepo.CreatePlayerHero(tx, &model.PlayerHero{
		PlayerID:     playerID,
		HeroConfigID: heroConfigID,
		Level:        1,
		Star:         1,
		Shard:        0,
		Locked:       0,
	}); err != nil {
		return GrantResult{}, err
	}
	return GrantResult{Type: "hero", ID: heroConfigID, Count: 1}, nil
}
