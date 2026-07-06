package service

import (
	"errors"
	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type TeamService struct {
	teamRepo   *repository.TeamRepository
	heroRepo   *repository.HeroRepository
	playerRepo *repository.PlayerRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, heroRepo *repository.HeroRepository, playerRepo *repository.PlayerRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, heroRepo: heroRepo, playerRepo: playerRepo}
}

type TeamSlot struct {
	Slot         uint8  `json:"slot"`
	PlayerHeroID uint64 `json:"player_hero_id"`
}

func (s *TeamService) Save(playerID uint64, slots []TeamSlot) error {
	// 1. Validate count
	if len(slots) < 1 || len(slots) > 5 {
		return errors.New("team size must be between 1 and 5")
	}

	slotSeen := make(map[uint8]struct{})
	heroSeen := make(map[uint64]struct{})
	for _, t := range slots {
		if t.Slot < 1 || t.Slot > 5 {
			return errors.New("slot out of range")
		}
		if _, ok := slotSeen[t.Slot]; ok {
			return errors.New("duplicate slot")
		}
		slotSeen[t.Slot] = struct{}{}
		if t.PlayerHeroID == 0 {
			return errors.New("invalid player_hero_id")
		}
		if _, ok := heroSeen[t.PlayerHeroID]; ok {
			return errors.New("duplicate player_hero_id")
		}
		heroSeen[t.PlayerHeroID] = struct{}{}
	}

	// 5. verify ownership: list all player's heroes and ensure provided ids belong to player
	heroes, err := s.heroRepo.ListPlayerHeroes(playerID)
	if err != nil {
		return err
	}
	playerHeroMap := make(map[uint64]model.PlayerHero)
	for _, h := range heroes {
		playerHeroMap[h.ID] = h
	}
	usedHeroIDs := make([]uint64, 0, len(slots))
	for _, t := range slots {
		if _, ok := playerHeroMap[t.PlayerHeroID]; !ok {
			return errors.New("one or more heroes do not belong to player")
		}
		usedHeroIDs = append(usedHeroIDs, playerHeroMap[t.PlayerHeroID].HeroConfigID)
	}

	// collect unique config ids
	cfgIDMap := make(map[uint64]struct{})
	cfgIDs := make([]uint64, 0, len(usedHeroIDs))
	for _, id := range usedHeroIDs {
		if _, ok := cfgIDMap[id]; !ok {
			cfgIDMap[id] = struct{}{}
			cfgIDs = append(cfgIDs, id)
		}
	}

	// fetch configs
	configs, err := s.heroRepo.ListConfigsByIDs(cfgIDs)
	if err != nil {
		return err
	}
	cfgMap := make(map[uint64]model.HeroConfig)
	for _, c := range configs {
		cfgMap[c.ID] = c
	}

	// perform DB transaction: delete old team, insert new team, update power
	err = s.teamRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.teamRepo.DeleteByPlayer(tx, playerID); err != nil {
			return err
		}

		var totalPower uint64 = 0
		for _, t := range slots {
			// insert team row
			team := &model.PlayerTeam{
				PlayerID:     playerID,
				Slot:         t.Slot,
				PlayerHeroID: t.PlayerHeroID,
			}
			if err := s.teamRepo.CreatePlayerTeam(tx, team); err != nil {
				return err
			}

			// calc power for this hero
			ph := playerHeroMap[t.PlayerHeroID]
			cfg, ok := cfgMap[ph.HeroConfigID]
			if !ok {
				// should not happen because we fetched configs for used heroes
				return errors.New("hero config missing")
			}
			p := CalcHeroPower(cfg, ph.Level, ph.Star)
			totalPower += p
		}

		if err := s.playerRepo.UpdateProfilePower(tx, playerID, totalPower); err != nil {
			return err
		}

		return nil
	})
	return err
}

type TeamView struct {
	Slot         uint8  `json:"slot"`
	PlayerHeroID uint64 `json:"player_hero_id"`
	ID           uint64 `json:"id"`
	HeroConfigID uint64 `json:"hero_config_id"`
	Name         string `json:"name"`
	Quality      uint8  `json:"quality"`
	Level        uint32 `json:"level"`
	Star         uint32 `json:"star"`
}

func (s *TeamService) Get(playerID uint64) ([]TeamView, error) {
	teams, err := s.teamRepo.ListByPlayer(playerID)
	if err != nil {
		return nil, err
	}

	heroes, err := s.heroRepo.ListPlayerHeroes(playerID)
	if err != nil {
		return nil, err
	}
	heroMap := make(map[uint64]model.PlayerHero)
	configIDs := make([]uint64, 0, len(heroes))
	configIDSet := make(map[uint64]struct{})
	for _, hero := range heroes {
		heroMap[hero.ID] = hero
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

	views := make([]TeamView, 0, len(teams))
	for _, team := range teams {
		hero, ok := heroMap[team.PlayerHeroID]
		if !ok {
			return nil, errors.New("team contains invalid hero")
		}
		cfg, ok := cfgMap[hero.HeroConfigID]
		if !ok {
			return nil, errors.New("hero config missing")
		}
		views = append(views, TeamView{
			Slot:         team.Slot,
			PlayerHeroID: team.PlayerHeroID,
			ID:           team.PlayerHeroID,
			HeroConfigID: hero.HeroConfigID,
			Name:         cfg.Name,
			Quality:      cfg.Quality,
			Level:        hero.Level,
			Star:         hero.Star,
		})
	}
	return views, nil
}
