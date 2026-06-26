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

func NewRewardService(assetRepo *repository.AssetRepository, heroRepo *repository.HeroRepository) *RewardService {
	return &RewardService{assetRepo: assetRepo, heroRepo: heroRepo}
}

func (s *RewardService) Grant(tx *gorm.DB, playerID uint64, rewards []model.Reward) error {
	for _, reward := range rewards {
		switch reward.Type {
		case "gold":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return err
			}
			asset.Gold += reward.Count
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return err
			}
		case "diamond":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return err
			}
			asset.Diamond += reward.Count
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return err
			}
		case "stamina":
			asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
			if err != nil {
				return err
			}
			asset.Stamina += uint32(reward.Count)
			if err := s.assetRepo.Save(tx, asset); err != nil {
				return err
			}
		case "hero":
			if err := s.grantHero(tx, playerID, reward.ID); err != nil {
				return err
			}
		default:
			return errors.New("unknown reward type")
		}
	}
	return nil
}

func (s *RewardService) grantHero(tx *gorm.DB, playerID uint64, heroConfigID uint64) error {
	hero, err := s.heroRepo.LockPlayerHero(tx, playerID, heroConfigID)
	if err == nil {
		hero.Shard += 10
		return s.heroRepo.SavePlayerHero(tx, hero)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.heroRepo.CreatePlayerHero(tx, &model.PlayerHero{
		PlayerID:     playerID,
		HeroConfigID: heroConfigID,
		Level:        1,
		Star:         1,
		Shard:        0,
		Locked:       0,
	})
}
