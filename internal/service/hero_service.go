package service

import "mini-card-game/internal/repository"

type HeroService struct {
	heroRepo *repository.HeroRepository
}

func NewHeroService(heroRepo *repository.HeroRepository) *HeroService {
	return &HeroService{heroRepo: heroRepo}
}

type HeroView struct {
	PlayerHeroID uint64 `json:"player_hero_id"`
	ID           uint64 `json:"id"`
	HeroConfigID uint64 `json:"hero_config_id"`
	Name         string `json:"name"`
	Quality      uint8  `json:"quality"`
	Level        uint32 `json:"level"`
	Star         uint32 `json:"star"`
	Shard        uint32 `json:"shard"`
}

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

	configMap := make(map[uint64]string)
	qualityMap := make(map[uint64]uint8)
	for _, cfg := range configs {
		configMap[cfg.ID] = cfg.Name
		qualityMap[cfg.ID] = cfg.Quality
	}

	result := make([]HeroView, 0, len(heroes))
	for _, hero := range heroes {
		result = append(result, HeroView{
			PlayerHeroID: hero.ID,
			ID:           hero.ID,
			HeroConfigID: hero.HeroConfigID,
			Name:         configMap[hero.HeroConfigID],
			Quality:      qualityMap[hero.HeroConfigID],
			Level:        hero.Level,
			Star:         hero.Star,
			Shard:        hero.Shard,
		})
	}
	return result, nil
}
