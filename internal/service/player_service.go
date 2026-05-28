package service

import "mini-card-game/internal/repository"

type PlayerService struct {
	playerRepo *repository.PlayerRepository
}

func NewPlayerService(playerRepo *repository.PlayerRepository) *PlayerService {
	return &PlayerService{playerRepo: playerRepo}
}

func (s *PlayerService) Profile(playerID uint64) (interface{}, error) {
	return s.playerRepo.FindProfile(playerID)
}

func (s *PlayerService) Assets(playerID uint64) (interface{}, error) {
	return s.playerRepo.FindAsset(playerID)
}
