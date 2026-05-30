package service

import (
	"time"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"
)

const (
	MaxStamina            uint32 = 120
	StaminaRecoverSeconds        = 5 * 60
)

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
	asset, err := s.playerRepo.FindAsset(playerID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	changed := SettleStamina(asset, now)
	if changed {
		if err := s.playerRepo.SaveAsset(asset); err != nil {
			return nil, err
		}
	}
	return BuildAssetView(asset, now), nil
}

type AssetView struct {
	PlayerID            uint64 `json:"player_id"`
	Gold                uint64 `json:"gold"`
	Diamond             uint64 `json:"diamond"`
	Stamina             uint32 `json:"stamina"`
	MaxStamina          uint32 `json:"max_stamina"`
	NextStaminaSeconds  int64  `json:"next_stamina_seconds"`
	StaminaRecoverEvery int64  `json:"stamina_recover_every"`
}

func BuildAssetView(asset *model.PlayerAsset, now time.Time) AssetView {
	return AssetView{
		PlayerID:            asset.PlayerID,
		Gold:                asset.Gold,
		Diamond:             asset.Diamond,
		Stamina:             asset.Stamina,
		MaxStamina:          MaxStamina,
		NextStaminaSeconds:  NextStaminaSeconds(asset, now),
		StaminaRecoverEvery: StaminaRecoverSeconds,
	}
}

func SettleStamina(asset *model.PlayerAsset, now time.Time) bool {
	if asset.Stamina >= MaxStamina {
		if asset.Stamina > MaxStamina {
			asset.Stamina = MaxStamina
			return true
		}
		return false
	}
	if asset.StaminaUpdatedAt == nil {
		asset.StaminaUpdatedAt = &now
		return true
	}
	elapsed := int64(now.Sub(*asset.StaminaUpdatedAt).Seconds())
	if elapsed < StaminaRecoverSeconds {
		return false
	}
	recovered := uint32(elapsed / StaminaRecoverSeconds)
	if recovered == 0 {
		return false
	}
	asset.Stamina += recovered
	if asset.Stamina >= MaxStamina {
		asset.Stamina = MaxStamina
		asset.StaminaUpdatedAt = nil
		return true
	}
	next := asset.StaminaUpdatedAt.Add(time.Duration(recovered*StaminaRecoverSeconds) * time.Second)
	asset.StaminaUpdatedAt = &next
	return true
}

func TouchStaminaAfterSpend(asset *model.PlayerAsset, now time.Time) {
	if asset.Stamina < MaxStamina && asset.StaminaUpdatedAt == nil {
		asset.StaminaUpdatedAt = &now
	}
	if asset.Stamina >= MaxStamina {
		asset.StaminaUpdatedAt = nil
	}
}

func NextStaminaSeconds(asset *model.PlayerAsset, now time.Time) int64 {
	if asset.Stamina >= MaxStamina || asset.StaminaUpdatedAt == nil {
		return 0
	}
	nextAt := asset.StaminaUpdatedAt.Add(StaminaRecoverSeconds * time.Second)
	seconds := int64(nextAt.Sub(now).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
