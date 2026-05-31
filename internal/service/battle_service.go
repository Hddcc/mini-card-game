package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

const (
	BattleSessionTTL = 30 * time.Minute
	MaxBattleLogs    = 30

	ActionNormal = "normal_attack"
	ActionSkill  = "skill"
	ActionDefend = "defend"

	EffectDamage      = "damage"
	EffectHeal        = "heal"
	EffectAttackBuff  = "attack_buff"
	EffectDefenseBuff = "defense_buff"
	EffectDefend      = "defend"
)

var (
	ErrActiveBattleExists  = errors.New("active battle already exists")
	ErrBattleNotFound      = errors.New("battle session not found")
	ErrBattleFinished      = errors.New("battle session finished")
	ErrBattleExpired       = errors.New("battle session expired")
	ErrInvalidBattleAction = errors.New("invalid battle action")
	ErrBattleConfigMissing = errors.New("battle config missing")
)

type BattleService struct {
	db          *gorm.DB
	battleRepo  *repository.BattleRepository
	stageRepo   *repository.StageRepository
	teamRepo    *repository.TeamRepository
	heroRepo    *repository.HeroRepository
	assetRepo   *repository.AssetRepository
	playerRepo  *repository.PlayerRepository
	taskService *TaskService
}

func NewBattleService(
	db *gorm.DB,
	battleRepo *repository.BattleRepository,
	stageRepo *repository.StageRepository,
	teamRepo *repository.TeamRepository,
	heroRepo *repository.HeroRepository,
	assetRepo *repository.AssetRepository,
	playerRepo *repository.PlayerRepository,
	taskService *TaskService,
) *BattleService {
	return &BattleService{
		db:          db,
		battleRepo:  battleRepo,
		stageRepo:   stageRepo,
		teamRepo:    teamRepo,
		heroRepo:    heroRepo,
		assetRepo:   assetRepo,
		playerRepo:  playerRepo,
		taskService: taskService,
	}
}

type BattleAnimationKeys struct {
	Attack string `json:"attack"`
	Skill  string `json:"skill"`
	Hit    string `json:"hit"`
	Defeat string `json:"defeat"`
	Idle   string `json:"idle"`
}

type BattleBuff struct {
	Stat            string `json:"stat"`
	Amount          int32  `json:"amount"`
	RemainingRounds uint32 `json:"remaining_rounds"`
	SourceSkillID   uint64 `json:"source_skill_id"`
}

type BattleUnit struct {
	ID              string              `json:"id"`
	Side            string              `json:"side"`
	BoardSide       string              `json:"board_side"`
	Slot            uint8               `json:"slot"`
	BoardSlot       uint8               `json:"board_slot"`
	Name            string              `json:"name"`
	Role            string              `json:"role"`
	SourceID        uint64              `json:"source_id"`
	PlayerHeroID    uint64              `json:"player_hero_id,omitempty"`
	ConfigID        uint64              `json:"config_id"`
	Level           uint32              `json:"level"`
	Star            uint32              `json:"star"`
	MaxHP           uint32              `json:"max_hp"`
	HP              uint32              `json:"hp"`
	ATK             uint32              `json:"atk"`
	DEF             uint32              `json:"def"`
	Rage            uint32              `json:"rage"`
	SkillID         uint64              `json:"skill_id"`
	SkillName       string              `json:"skill_name"`
	SkillTarget     string              `json:"skill_target"`
	SkillEffect     string              `json:"skill_effect"`
	SkillMultiplier uint32              `json:"skill_multiplier"`
	SkillDuration   uint32              `json:"skill_duration"`
	SkillStatDelta  int32               `json:"skill_stat_delta"`
	SkillCostRage   uint32              `json:"skill_cost_rage"`
	SkillCooldown   uint32              `json:"skill_cooldown"`
	SkillDesc       string              `json:"skill_description"`
	SkillEffectKey  string              `json:"skill_effect_key"`
	SkillAnimation  string              `json:"skill_animation"`
	CardArt         string              `json:"card_art"`
	PortraitArt     string              `json:"portrait_art"`
	Animations      BattleAnimationKeys `json:"animation_keys"`
	CooldownLeft    uint32              `json:"cooldown_left"`
	Defending       bool                `json:"defending"`
	Alive           bool                `json:"alive"`
	Buffs           []BattleBuff        `json:"buffs,omitempty"`
	FailureHintTag  string              `json:"failure_hint_tag,omitempty"`
}

type BattleLog struct {
	Round        uint32 `json:"round"`
	ActorID      string `json:"actor_id"`
	ActorName    string `json:"actor_name"`
	Action       string `json:"action"`
	TargetID     string `json:"target_id,omitempty"`
	TargetName   string `json:"target_name,omitempty"`
	Damage       uint32 `json:"damage,omitempty"`
	Heal         uint32 `json:"heal,omitempty"`
	BuffStat     string `json:"buff_stat,omitempty"`
	BuffAmount   int32  `json:"buff_amount,omitempty"`
	EffectKey    string `json:"effect_key,omitempty"`
	AnimationKey string `json:"animation_key,omitempty"`
	Text         string `json:"text"`
}

type BattleActionView struct {
	Type         string   `json:"type"`
	Label        string   `json:"label"`
	SkillID      uint64   `json:"skill_id,omitempty"`
	Enabled      bool     `json:"enabled"`
	TargetType   string   `json:"target_type,omitempty"`
	EffectType   string   `json:"effect_type,omitempty"`
	EffectKey    string   `json:"effect_key,omitempty"`
	Animation    string   `json:"animation_key,omitempty"`
	ValidTargets []string `json:"valid_targets,omitempty"`
	Description  string   `json:"description,omitempty"`
}

type BattleState struct {
	Version          uint32             `json:"version"`
	Round            uint32             `json:"round"`
	Status           string             `json:"status"`
	EncounterID      uint64             `json:"encounter_id,omitempty"`
	EncounterName    string             `json:"encounter_name,omitempty"`
	CurrentActorID   string             `json:"current_actor_id"`
	SelectableActors []string           `json:"selectable_actors"`
	PlayerUnits      []BattleUnit       `json:"player_units"`
	EnemyUnits       []BattleUnit       `json:"enemy_units"`
	AvailableActions []BattleActionView `json:"available_actions"`
	SelectedTargets  []string           `json:"selected_targets"`
	Logs             []BattleLog        `json:"logs"`
	FailureHint      string             `json:"failure_hint,omitempty"`
}

type BattleResult struct {
	Win                bool   `json:"win"`
	RewardGold         uint64 `json:"reward_gold"`
	RewardExp          uint32 `json:"reward_exp"`
	Stamina            uint32 `json:"stamina"`
	MaxStamina         uint32 `json:"max_stamina"`
	NextStaminaSeconds int64  `json:"next_stamina_seconds"`
	BestPower          uint64 `json:"best_power"`
	FailureHint        string `json:"failure_hint,omitempty"`
}

type BattleResponse struct {
	SessionID uint64        `json:"session_id"`
	StageID   uint64        `json:"stage_id"`
	Status    string        `json:"status"`
	ExpiresAt time.Time     `json:"expires_at"`
	State     BattleState   `json:"state"`
	Result    *BattleResult `json:"result,omitempty"`
}

type BattleActionInput struct {
	SessionID uint64 `json:"session_id"`
	Action    string `json:"action"`
	ActorID   string `json:"actor_id"`
	TargetID  string `json:"target_id"`
	SkillID   uint64 `json:"skill_id"`
}

func (s *BattleService) Start(playerID uint64, stageID uint64) (*BattleResponse, error) {
	now := time.Now()
	active, err := s.resumeActiveBattle(playerID, now)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return active, nil
	}

	config, err := s.stageRepo.FindConfig(stageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStageNotFound
		}
		return nil, err
	}
	if err := s.ensureStageUnlocked(playerID, config); err != nil {
		return nil, err
	}

	teamRows, heroes, heroConfigs, _, err := s.loadTeamSnapshot(playerID)
	if err != nil {
		return nil, err
	}
	if len(teamRows) == 0 {
		return nil, ErrNoTeam
	}

	stageEnemies, encounter, enemyConfigs, enemySkills, err := s.loadEnemySnapshot(config)
	if err != nil {
		return nil, err
	}
	playerSkills, err := s.loadPlayerSkills(heroConfigs)
	if err != nil {
		return nil, err
	}
	heroSkins, err := s.loadCardSkins("hero", heroConfigIDs(heroConfigs))
	if err != nil {
		return nil, err
	}
	enemySkins, err := s.loadCardSkins("enemy", enemyConfigIDs(enemyConfigs))
	if err != nil {
		return nil, err
	}

	state, err := buildInitialBattleState(teamRows, heroes, heroConfigs, playerSkills, heroSkins, stageEnemies, encounter, enemyConfigs, enemySkills, enemySkins)
	if err != nil {
		return nil, err
	}
	state.appendLog(BattleLog{Round: 1, Action: "start", EffectKey: "effect-start", AnimationKey: "fx-idle-glow", Text: "战斗开始，选择一张我方卡牌行动。"})
	state.refreshTurn()

	expiresAt := now.Add(BattleSessionTTL)
	stateJSON, err := encodeBattleState(state)
	if err != nil {
		return nil, err
	}

	var created *model.PlayerBattleSession
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.battleRepo.AbandonExpiredActiveSessions(tx, playerID, now); err != nil {
			return err
		}
		hasActive, err := s.battleRepo.HasActiveSession(tx, playerID, now)
		if err != nil {
			return err
		}
		if hasActive {
			return ErrActiveBattleExists
		}

		asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
		if err != nil {
			return err
		}
		if err := spendBattleStamina(asset, config.StaminaCost, now); err != nil {
			return err
		}
		if err := s.assetRepo.Save(tx, asset); err != nil {
			return err
		}

		session := &model.PlayerBattleSession{
			PlayerID:  playerID,
			StageID:   stageID,
			Status:    model.BattleStatusActive,
			Round:     state.Round,
			StateJSON: stateJSON,
			ExpiresAt: expiresAt,
		}
		if err := s.battleRepo.CreateSession(tx, session); err != nil {
			return err
		}
		created = session
		return nil
	})
	if err != nil {
		return nil, err
	}

	return buildBattleResponse(created, state, nil), nil
}

func (s *BattleService) resumeActiveBattle(playerID uint64, now time.Time) (*BattleResponse, error) {
	var response *BattleResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.battleRepo.AbandonExpiredActiveSessions(tx, playerID, now); err != nil {
			return err
		}
		session, err := s.battleRepo.LockActiveSession(tx, playerID, now)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		state, err := decodeBattleState(session.StateJSON)
		if err != nil {
			return err
		}
		if compactBattleStateEnemies(&state, 4) {
			session.StateJSON = mustEncodeBattleState(state)
			if err := s.battleRepo.SaveSession(tx, session); err != nil {
				return err
			}
		}
		response = buildBattleResponse(session, state, nil)
		return nil
	})
	return response, err
}

func (s *BattleService) Action(playerID uint64, input BattleActionInput) (*BattleResponse, error) {
	if input.SessionID == 0 || input.Action == "" {
		return nil, ErrInvalidBattleAction
	}

	now := time.Now()
	var response *BattleResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		session, err := s.battleRepo.LockSession(tx, playerID, input.SessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBattleNotFound
			}
			return err
		}
		if err := ensureActionableSession(session, now); err != nil {
			if errors.Is(err, ErrBattleExpired) {
				session.Status = model.BattleStatusAbandoned
				if saveErr := s.battleRepo.SaveSession(tx, session); saveErr != nil {
					return saveErr
				}
			}
			return err
		}

		state, err := decodeBattleState(session.StateJSON)
		if err != nil {
			return err
		}
		if err := applyPlayerAction(&state, input); err != nil {
			return err
		}
		if allDefeated(state.EnemyUnits) {
			result, err := s.settleVictory(tx, playerID, session.StageID, &state)
			if err != nil {
				return err
			}
			now := time.Now()
			session.Status = model.BattleStatusWon
			session.SettledAt = &now
			state.Status = model.BattleStatusWon
			session.ResultJSON = mustEncodeBattleResult(result)
			session.StateJSON = mustEncodeBattleState(state)
			session.Round = state.Round
			if err := s.battleRepo.SaveSession(tx, session); err != nil {
				return err
			}
			response = buildBattleResponse(session, state, result)
			return nil
		}

		resolveEnemyTurn(&state)
		if allDefeated(state.PlayerUnits) {
			result, err := s.settleDefeat(tx, playerID, &state)
			if err != nil {
				return err
			}
			now := time.Now()
			session.Status = model.BattleStatusLost
			session.SettledAt = &now
			state.Status = model.BattleStatusLost
			state.FailureHint = result.FailureHint
			session.ResultJSON = mustEncodeBattleResult(result)
			session.StateJSON = mustEncodeBattleState(state)
			session.Round = state.Round
			if err := s.battleRepo.SaveSession(tx, session); err != nil {
				return err
			}
			response = buildBattleResponse(session, state, result)
			return nil
		}

		state.Round++
		state.refreshTurn()
		session.StateJSON = mustEncodeBattleState(state)
		session.Round = state.Round
		if err := s.battleRepo.SaveSession(tx, session); err != nil {
			return err
		}
		response = buildBattleResponse(session, state, nil)
		return nil
	})
	return response, err
}

func (s *BattleService) ensureStageUnlocked(playerID uint64, config *model.StageConfig) error {
	if config.PrevStageID == 0 {
		return nil
	}
	prevStage, err := s.stageRepo.FindPlayerStage(playerID, config.PrevStageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPrevStageNotCleared
		}
		return err
	}
	if prevStage.Status != 1 {
		return ErrPrevStageNotCleared
	}
	return nil
}

func (s *BattleService) loadTeamSnapshot(playerID uint64) ([]model.PlayerTeam, map[uint64]model.PlayerHero, map[uint64]model.HeroConfig, uint64, error) {
	teamRows, err := s.teamRepo.ListByPlayer(playerID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	if len(teamRows) == 0 {
		return teamRows, nil, nil, 0, nil
	}

	heroes, err := s.heroRepo.ListPlayerHeroes(playerID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	heroMap := make(map[uint64]model.PlayerHero)
	configIDSet := make(map[uint64]struct{})
	configIDs := make([]uint64, 0, len(heroes))
	for _, hero := range heroes {
		heroMap[hero.ID] = hero
		if _, ok := configIDSet[hero.HeroConfigID]; !ok {
			configIDSet[hero.HeroConfigID] = struct{}{}
			configIDs = append(configIDs, hero.HeroConfigID)
		}
	}

	configs, err := s.heroRepo.ListConfigsByIDs(configIDs)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	configMap := make(map[uint64]model.HeroConfig)
	for _, cfg := range configs {
		configMap[cfg.ID] = cfg
	}

	var totalPower uint64
	for _, team := range teamRows {
		hero, ok := heroMap[team.PlayerHeroID]
		if !ok {
			return nil, nil, nil, 0, errors.New("team contains invalid hero")
		}
		cfg, ok := configMap[hero.HeroConfigID]
		if !ok {
			return nil, nil, nil, 0, errors.New("hero config missing")
		}
		totalPower += CalcHeroPower(cfg.BaseATK, cfg.BaseHP, cfg.BaseDEF, hero.Level, hero.Star, cfg.PowerFactor)
	}
	return teamRows, heroMap, configMap, totalPower, nil
}

func (s *BattleService) loadEnemySnapshot(config *model.StageConfig) ([]model.StageEncounterEnemy, *model.StageEncounterVariant, map[uint64]model.EnemyConfig, map[uint64]model.SkillConfig, error) {
	var selected *model.StageEncounterVariant
	var stageEnemies []model.StageEncounterEnemy

	variants, err := s.battleRepo.ListEncounterVariants(config.ID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if len(variants) > 0 {
		selected = chooseEncounterVariant(variants, config.RecommendPower, time.Now())
		if selected != nil {
			stageEnemies, err = s.battleRepo.ListEncounterEnemies(selected.ID)
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}
	if len(stageEnemies) == 0 {
		legacyRows, err := s.battleRepo.ListStageEnemies(config.ID)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		for _, row := range legacyRows {
			stageEnemies = append(stageEnemies, model.StageEncounterEnemy{
				EnemyConfigID: row.EnemyConfigID,
				Slot:          row.Slot,
				Level:         row.Level,
				Count:         row.Count,
			})
		}
	}
	if len(stageEnemies) == 0 {
		return nil, nil, nil, nil, ErrBattleConfigMissing
	}
	stageEnemies = compactEncounterEnemies(stageEnemies, 4)

	enemyIDSet := make(map[uint64]struct{})
	enemyIDs := make([]uint64, 0, len(stageEnemies))
	skillIDSet := make(map[uint64]struct{})
	for _, row := range stageEnemies {
		if _, ok := enemyIDSet[row.EnemyConfigID]; !ok {
			enemyIDSet[row.EnemyConfigID] = struct{}{}
			enemyIDs = append(enemyIDs, row.EnemyConfigID)
		}
		if row.SkillID > 0 {
			skillIDSet[row.SkillID] = struct{}{}
		}
	}
	enemies, err := s.battleRepo.ListEnemyConfigsByIDs(enemyIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	enemyMap := make(map[uint64]model.EnemyConfig)
	for _, enemy := range enemies {
		enemyMap[enemy.ID] = enemy
		if enemy.SkillID > 0 {
			skillIDSet[enemy.SkillID] = struct{}{}
		}
	}
	if len(enemyMap) != len(enemyIDs) {
		return nil, nil, nil, nil, ErrBattleConfigMissing
	}

	skillIDs := make([]uint64, 0, len(skillIDSet))
	for id := range skillIDSet {
		skillIDs = append(skillIDs, id)
	}
	skills, err := s.battleRepo.ListSkillConfigsByIDs(skillIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	skillMap := make(map[uint64]model.SkillConfig)
	for _, skill := range skills {
		if err := validateSkillConfig(skill); err != nil {
			return nil, nil, nil, nil, err
		}
		skillMap[skill.ID] = skill
	}
	return stageEnemies, selected, enemyMap, skillMap, nil
}

func (s *BattleService) loadPlayerSkills(configs map[uint64]model.HeroConfig) (map[uint64]model.SkillConfig, error) {
	skillIDSet := make(map[uint64]struct{})
	skillIDs := make([]uint64, 0, len(configs))
	for _, cfg := range configs {
		skillID := skillIDForHero(cfg)
		if _, ok := skillIDSet[skillID]; !ok {
			skillIDSet[skillID] = struct{}{}
			skillIDs = append(skillIDs, skillID)
		}
	}
	skills, err := s.battleRepo.ListSkillConfigsByIDs(skillIDs)
	if err != nil {
		return nil, err
	}
	skillMap := make(map[uint64]model.SkillConfig)
	for _, skill := range skills {
		if err := validateSkillConfig(skill); err != nil {
			return nil, err
		}
		skillMap[skill.ID] = skill
	}
	if len(skillMap) != len(skillIDs) {
		return nil, ErrBattleConfigMissing
	}
	return skillMap, nil
}

func compactEncounterEnemies(rows []model.StageEncounterEnemy, maxUnits uint32) []model.StageEncounterEnemy {
	if maxUnits == 0 || len(rows) == 0 {
		return rows
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Slot == rows[j].Slot {
			return rows[i].EnemyConfigID < rows[j].EnemyConfigID
		}
		return rows[i].Slot < rows[j].Slot
	})

	seen := make(map[string]struct{})
	result := make([]model.StageEncounterEnemy, 0, len(rows))
	var total uint32
	for _, row := range rows {
		if row.Count == 0 {
			continue
		}
		key := fmt.Sprintf("%d:%d:%d:%d", row.Slot, row.EnemyConfigID, row.Level, row.SkillID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if total >= maxUnits {
			break
		}
		if total+row.Count > maxUnits {
			row.Count = maxUnits - total
		}
		total += row.Count
		result = append(result, row)
	}
	return result
}

func compactBattleStateEnemies(state *BattleState, maxUnits int) bool {
	if state == nil || maxUnits <= 0 || len(state.EnemyUnits) <= maxUnits {
		return false
	}
	sort.SliceStable(state.EnemyUnits, func(i, j int) bool {
		if state.EnemyUnits[i].Alive != state.EnemyUnits[j].Alive {
			return state.EnemyUnits[i].Alive
		}
		if state.EnemyUnits[i].BoardSlot == state.EnemyUnits[j].BoardSlot {
			return state.EnemyUnits[i].ID < state.EnemyUnits[j].ID
		}
		return state.EnemyUnits[i].BoardSlot < state.EnemyUnits[j].BoardSlot
	})
	state.EnemyUnits = append([]BattleUnit(nil), state.EnemyUnits[:maxUnits]...)
	for idx := range state.EnemyUnits {
		state.EnemyUnits[idx].BoardSlot = uint8(idx + 1)
		state.EnemyUnits[idx].Slot = uint8(idx + 1)
	}
	state.SelectedTargets = nil
	state.refreshTurn()
	return true
}

func (s *BattleService) loadCardSkins(ownerType string, ownerIDs []uint64) (map[uint64]model.CardSkinConfig, error) {
	skins, err := s.battleRepo.ListCardSkins(ownerType, ownerIDs)
	if err != nil {
		return nil, err
	}
	skinMap := make(map[uint64]model.CardSkinConfig)
	for _, skin := range skins {
		skinMap[skin.OwnerID] = skin
	}
	return skinMap, nil
}

func (s *BattleService) settleVictory(tx *gorm.DB, playerID uint64, stageID uint64, state *BattleState) (*BattleResult, error) {
	config, err := s.stageRepo.FindConfig(stageID)
	if err != nil {
		return nil, err
	}
	asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	SettleStamina(asset, now)
	asset.Gold += config.RewardGold
	if err := s.assetRepo.Save(tx, asset); err != nil {
		return nil, err
	}

	totalPower := totalBattlePower(state.PlayerUnits)
	stageRow, err := s.stageRepo.LockPlayerStage(tx, playerID, stageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stageRow = &model.PlayerStage{PlayerID: playerID, StageID: stageID, Status: 1, BestPower: totalPower, FirstPassedAt: &now}
		} else {
			return nil, err
		}
	}
	stageRow.Status = 1
	if stageRow.FirstPassedAt == nil {
		stageRow.FirstPassedAt = &now
	}
	if totalPower > stageRow.BestPower {
		stageRow.BestPower = totalPower
	}
	if err := s.stageRepo.SavePlayerStage(tx, stageRow); err != nil {
		return nil, err
	}
	if err := s.playerRepo.AddProfileExp(tx, playerID, config.RewardExp); err != nil {
		return nil, err
	}
	if s.taskService != nil {
		if err := s.taskService.AddProgress(tx, playerID, "stage_fight", 1); err != nil {
			return nil, err
		}
		if err := s.taskService.AddProgress(tx, playerID, "stage_win", 1); err != nil {
			return nil, err
		}
	}
	state.appendLog(BattleLog{Round: state.Round, Action: "victory", EffectKey: "effect-victory", AnimationKey: "fx-gold-burst", Text: "妖魔退散，关卡通关！"})
	return &BattleResult{Win: true, RewardGold: config.RewardGold, RewardExp: config.RewardExp, Stamina: asset.Stamina, MaxStamina: MaxStamina, NextStaminaSeconds: NextStaminaSeconds(asset, now), BestPower: stageRow.BestPower}, nil
}

func (s *BattleService) settleDefeat(tx *gorm.DB, playerID uint64, state *BattleState) (*BattleResult, error) {
	asset, err := s.assetRepo.LockByPlayerID(tx, playerID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	SettleStamina(asset, now)
	if err := s.assetRepo.Save(tx, asset); err != nil {
		return nil, err
	}
	if s.taskService != nil {
		if err := s.taskService.AddProgress(tx, playerID, "stage_fight", 1); err != nil {
			return nil, err
		}
	}
	hint := buildFailureHint(*state)
	state.appendLog(BattleLog{Round: state.Round, Action: "defeat", EffectKey: "effect-defeat", AnimationKey: "fx-defeat-smoke", Text: hint})
	return &BattleResult{Win: false, Stamina: asset.Stamina, MaxStamina: MaxStamina, NextStaminaSeconds: NextStaminaSeconds(asset, now), FailureHint: hint}, nil
}

func spendBattleStamina(asset *model.PlayerAsset, cost uint32, now time.Time) error {
	SettleStamina(asset, now)
	if asset.Stamina < cost {
		return ErrNotEnoughStamina
	}
	asset.Stamina -= cost
	TouchStaminaAfterSpend(asset, now)
	return nil
}

func ensureActionableSession(session *model.PlayerBattleSession, now time.Time) error {
	if session.Status != model.BattleStatusActive {
		return ErrBattleFinished
	}
	if !session.ExpiresAt.After(now) {
		return ErrBattleExpired
	}
	return nil
}

func buildInitialBattleState(
	teamRows []model.PlayerTeam,
	heroes map[uint64]model.PlayerHero,
	heroConfigs map[uint64]model.HeroConfig,
	playerSkills map[uint64]model.SkillConfig,
	heroSkins map[uint64]model.CardSkinConfig,
	stageEnemies []model.StageEncounterEnemy,
	encounter *model.StageEncounterVariant,
	enemyConfigs map[uint64]model.EnemyConfig,
	enemySkills map[uint64]model.SkillConfig,
	enemySkins map[uint64]model.CardSkinConfig,
) (BattleState, error) {
	state := BattleState{Version: 2, Round: 1, Status: model.BattleStatusActive}
	if encounter != nil {
		state.EncounterID = encounter.ID
		state.EncounterName = encounter.Name
	}

	for _, team := range teamRows {
		hero := heroes[team.PlayerHeroID]
		cfg := heroConfigs[hero.HeroConfigID]
		skill := playerSkills[skillIDForHero(cfg)]
		skin := skinWithDefaults(heroSkins[cfg.ID], "hero", cfg.ID)
		unit := BattleUnit{
			ID:              fmt.Sprintf("p%d", team.Slot),
			Side:            "player",
			BoardSide:       "bottom",
			Slot:            team.Slot,
			BoardSlot:       team.Slot,
			Name:            cfg.Name,
			Role:            cfg.Role,
			SourceID:        hero.ID,
			PlayerHeroID:    hero.ID,
			ConfigID:        cfg.ID,
			Level:           hero.Level,
			Star:            hero.Star,
			MaxHP:           cfg.BaseHP + hero.Level*80 + hero.Star*120,
			ATK:             cfg.BaseATK + hero.Level*8 + hero.Star*20,
			DEF:             cfg.BaseDEF + hero.Level*4 + hero.Star*10,
			Rage:            50,
			SkillID:         skill.ID,
			SkillName:       skill.Name,
			SkillTarget:     skill.TargetType,
			SkillEffect:     skill.EffectType,
			SkillMultiplier: skill.Multiplier,
			SkillDuration:   skill.DurationRounds,
			SkillStatDelta:  skill.StatDelta,
			SkillCostRage:   skill.CostRage,
			SkillCooldown:   skill.Cooldown,
			SkillDesc:       skill.Description,
			SkillEffectKey:  effectKey(skill, "effect-skill"),
			SkillAnimation:  animationKey(skill, skin.SkillAnimation),
			CardArt:         skin.CardArt,
			PortraitArt:     skin.PortraitArt,
			Animations:      animationsFromSkin(skin),
			Alive:           true,
		}
		unit.HP = unit.MaxHP
		state.PlayerUnits = append(state.PlayerUnits, unit)
	}

	for _, row := range stageEnemies {
		cfg, ok := enemyConfigs[row.EnemyConfigID]
		if !ok {
			return state, ErrBattleConfigMissing
		}
		skillID := cfg.SkillID
		if row.SkillID > 0 {
			skillID = row.SkillID
		}
		skill, ok := enemySkills[skillID]
		if !ok && skillID > 0 {
			return state, ErrBattleConfigMissing
		}
		if err := validateSkillConfig(skill); err != nil {
			return state, err
		}
		skin := skinWithDefaults(enemySkins[cfg.ID], "enemy", cfg.ID)
		count := row.Count
		if count == 0 {
			count = 1
		}
		for i := uint32(0); i < count; i++ {
			slot := row.Slot + uint8(i)
			unit := BattleUnit{
				ID:              fmt.Sprintf("e%d_%d", row.Slot, i+1),
				Side:            "enemy",
				BoardSide:       "top",
				Slot:            slot,
				BoardSlot:       slot,
				Name:            cfg.Name,
				Role:            cfg.Role,
				SourceID:        cfg.ID,
				ConfigID:        cfg.ID,
				Level:           row.Level,
				Star:            1,
				MaxHP:           cfg.BaseHP + row.Level*100,
				ATK:             cfg.BaseATK + row.Level*10,
				DEF:             cfg.BaseDEF + row.Level*5,
				SkillID:         skill.ID,
				SkillName:       skill.Name,
				SkillTarget:     skill.TargetType,
				SkillEffect:     skill.EffectType,
				SkillMultiplier: skill.Multiplier,
				SkillDuration:   skill.DurationRounds,
				SkillStatDelta:  skill.StatDelta,
				SkillCostRage:   skill.CostRage,
				SkillCooldown:   skill.Cooldown,
				SkillDesc:       skill.Description,
				SkillEffectKey:  effectKey(skill, "effect-claw"),
				SkillAnimation:  animationKey(skill, skin.SkillAnimation),
				CardArt:         firstNonEmpty(cfg.CardArt, skin.CardArt),
				PortraitArt:     firstNonEmpty(cfg.PortraitArt, skin.PortraitArt),
				Animations:      animationsFromEnemy(cfg, skin),
				Alive:           true,
			}
			unit.HP = unit.MaxHP
			state.EnemyUnits = append(state.EnemyUnits, unit)
		}
	}

	sort.Slice(state.PlayerUnits, func(i, j int) bool { return state.PlayerUnits[i].BoardSlot < state.PlayerUnits[j].BoardSlot })
	sort.Slice(state.EnemyUnits, func(i, j int) bool { return state.EnemyUnits[i].BoardSlot < state.EnemyUnits[j].BoardSlot })
	return state, nil
}

func applyPlayerAction(state *BattleState, input BattleActionInput) error {
	state.refreshTurn()
	actor := findUnit(state.PlayerUnits, input.ActorID)
	if actor == nil || !actor.Alive || !containsString(state.SelectableActors, actor.ID) {
		return ErrInvalidBattleAction
	}
	actor.Defending = false

	switch input.Action {
	case ActionNormal:
		target := findUnit(state.EnemyUnits, input.TargetID)
		if target == nil || !target.Alive {
			return ErrInvalidBattleAction
		}
		damage := applyDamage(actor, target, 100)
		actor.Rage = min32(100, actor.Rage+25)
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionNormal, TargetID: target.ID, TargetName: target.Name, Damage: damage, EffectKey: "effect-attack", AnimationKey: actor.Animations.Attack, Text: fmt.Sprintf("%s 攻击 %s，造成 %d 点伤害。", actor.Name, target.Name, damage)})
	case ActionSkill:
		if actor.SkillID == 0 || actor.Rage < actor.SkillCostRage || actor.CooldownLeft > 0 {
			return ErrInvalidBattleAction
		}
		if input.SkillID > 0 && input.SkillID != actor.SkillID {
			return ErrInvalidBattleAction
		}
		if err := applySkill(state, actor, input.TargetID); err != nil {
			return err
		}
		if actor.SkillCostRage > actor.Rage {
			actor.Rage = 0
		} else {
			actor.Rage -= actor.SkillCostRage
		}
		actor.CooldownLeft = actor.SkillCooldown
	case ActionDefend:
		actor.Defending = true
		actor.Rage = min32(100, actor.Rage+35)
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionDefend, EffectKey: "effect-shield", AnimationKey: actor.Animations.Skill, Text: fmt.Sprintf("%s 进入防御姿态。", actor.Name)})
	default:
		return ErrInvalidBattleAction
	}
	return nil
}

func applySkill(state *BattleState, actor *BattleUnit, targetID string) error {
	switch actor.SkillEffect {
	case EffectDamage:
		target := findUnit(state.EnemyUnits, targetID)
		if target == nil || !target.Alive {
			return ErrInvalidBattleAction
		}
		multiplier := actorSkillMultiplier(actor)
		damage := applyDamage(actor, target, multiplier)
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionSkill, TargetID: target.ID, TargetName: target.Name, Damage: damage, EffectKey: actor.SkillEffectKey, AnimationKey: actor.SkillAnimation, Text: fmt.Sprintf("%s 施展 %s，命中 %s 造成 %d 点伤害。", actor.Name, actor.SkillName, target.Name, damage)})
	case EffectHeal:
		target := selectFriendlyTarget(state, actor.SkillTarget, targetID)
		if target == nil {
			return ErrInvalidBattleAction
		}
		heal := effectiveATK(actor) * actorSkillMultiplier(actor) / 100
		before := target.HP
		target.HP = min32(target.MaxHP, target.HP+heal)
		actual := target.HP - before
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionSkill, TargetID: target.ID, TargetName: target.Name, Heal: actual, EffectKey: actor.SkillEffectKey, AnimationKey: actor.SkillAnimation, Text: fmt.Sprintf("%s 施展 %s，治疗 %s %d 点生命。", actor.Name, actor.SkillName, target.Name, actual)})
	case EffectAttackBuff, EffectDefenseBuff:
		target := selectFriendlyTarget(state, actor.SkillTarget, targetID)
		if target == nil {
			return ErrInvalidBattleAction
		}
		stat := "atk"
		if actor.SkillEffect == EffectDefenseBuff {
			stat = "def"
		}
		amount := actorSkillStatDelta(actor)
		duration := actorSkillDuration(actor)
		target.Buffs = append(target.Buffs, BattleBuff{Stat: stat, Amount: amount, RemainingRounds: duration, SourceSkillID: actor.SkillID})
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionSkill, TargetID: target.ID, TargetName: target.Name, BuffStat: stat, BuffAmount: amount, EffectKey: actor.SkillEffectKey, AnimationKey: actor.SkillAnimation, Text: fmt.Sprintf("%s 施展 %s，强化 %s 的%s。", actor.Name, actor.SkillName, target.Name, battleStatName(stat))})
	case EffectDefend:
		actor.Defending = true
		state.appendLog(BattleLog{Round: state.Round, ActorID: actor.ID, ActorName: actor.Name, Action: ActionSkill, EffectKey: actor.SkillEffectKey, AnimationKey: actor.SkillAnimation, Text: fmt.Sprintf("%s 施展 %s，守住阵线。", actor.Name, actor.SkillName)})
	default:
		return ErrBattleConfigMissing
	}
	return nil
}

func resolveEnemyTurn(state *BattleState) {
	for i := range state.EnemyUnits {
		enemy := &state.EnemyUnits[i]
		if !enemy.Alive {
			continue
		}
		target := lowestHPUnit(state.PlayerUnits)
		if target == nil {
			return
		}
		multiplier := uint32(100)
		action := ActionNormal
		effectKey := "effect-attack"
		animation := enemy.Animations.Attack
		if enemy.SkillID > 0 && enemy.CooldownLeft == 0 {
			multiplier = actorSkillMultiplier(enemy)
			action = ActionSkill
			effectKey = enemy.SkillEffectKey
			animation = enemy.SkillAnimation
			enemy.CooldownLeft = enemy.SkillCooldown
		}
		damage := applyDamage(enemy, target, multiplier)
		enemy.Rage = min32(100, enemy.Rage+20)
		state.appendLog(BattleLog{Round: state.Round, ActorID: enemy.ID, ActorName: enemy.Name, Action: action, TargetID: target.ID, TargetName: target.Name, Damage: damage, EffectKey: effectKey, AnimationKey: animation, Text: fmt.Sprintf("%s 袭击 %s，造成 %d 点伤害。", enemy.Name, target.Name, damage)})
		if allDefeated(state.PlayerUnits) {
			return
		}
	}
	endRoundTick(state.PlayerUnits)
	endRoundTick(state.EnemyUnits)
}

func (s *BattleState) refreshTurn() {
	s.Status = model.BattleStatusActive
	s.SelectableActors = aliveUnitIDs(s.PlayerUnits)
	s.CurrentActorID = ""
	if len(s.SelectableActors) > 0 {
		s.CurrentActorID = s.SelectableActors[0]
	}
	s.SelectedTargets = aliveUnitIDs(s.EnemyUnits)
	s.AvailableActions = nil
	actor := findUnit(s.PlayerUnits, s.CurrentActorID)
	if actor == nil {
		return
	}
	s.AvailableActions = buildActionsForActor(*actor, *s)
}

func (s *BattleState) appendLog(log BattleLog) {
	s.Logs = append(s.Logs, log)
	if len(s.Logs) > MaxBattleLogs {
		s.Logs = s.Logs[len(s.Logs)-MaxBattleLogs:]
	}
}

func buildActionsForActor(actor BattleUnit, state BattleState) []BattleActionView {
	enemyTargets := aliveUnitIDs(state.EnemyUnits)
	actions := []BattleActionView{
		{Type: ActionNormal, Label: "普攻", Enabled: len(enemyTargets) > 0, TargetType: "enemy", EffectType: EffectDamage, EffectKey: "effect-attack", Animation: actor.Animations.Attack, ValidTargets: enemyTargets, Description: "攻击一个敌人并获得怒气"},
		{Type: ActionDefend, Label: "防御", Enabled: true, TargetType: "self", EffectType: EffectDefend, EffectKey: "effect-shield", Animation: actor.Animations.Skill, ValidTargets: []string{actor.ID}, Description: "降低本轮受到的伤害并获得怒气"},
	}
	targets := targetsForSkill(actor, state)
	actions = append(actions, BattleActionView{Type: ActionSkill, Label: actor.SkillName, SkillID: actor.SkillID, Enabled: actor.SkillID > 0 && actor.Rage >= actor.SkillCostRage && actor.CooldownLeft == 0 && len(targets) > 0, TargetType: actor.SkillTarget, EffectType: actor.SkillEffect, EffectKey: actor.SkillEffectKey, Animation: actor.SkillAnimation, ValidTargets: targets, Description: fmt.Sprintf("消耗 %d 怒气：%s", actor.SkillCostRage, actor.SkillDesc)})
	return actions
}

func applyDamage(actor *BattleUnit, target *BattleUnit, multiplier uint32) uint32 {
	raw := effectiveATK(actor) * multiplier / 100
	def := effectiveDEF(target)
	if target.Defending {
		def *= 2
	}
	damage := uint32(1)
	if raw > def {
		damage = raw - def
	}
	if damage >= target.HP {
		damage = target.HP
		target.HP = 0
		target.Alive = false
		target.Defending = false
	} else {
		target.HP -= damage
		target.Rage = min32(100, target.Rage+10)
	}
	return damage
}

func chooseEncounterVariant(variants []model.StageEncounterVariant, recommendPower uint64, now time.Time) *model.StageEncounterVariant {
	fits := make([]model.StageEncounterVariant, 0, len(variants))
	for _, variant := range variants {
		if variant.Status != 1 {
			continue
		}
		if variant.MinPower <= recommendPower && recommendPower <= variant.MaxPower {
			fits = append(fits, variant)
			continue
		}
		low := recommendPower * 70 / 100
		high := recommendPower * 130 / 100
		if low <= variant.EstimatedPower && variant.EstimatedPower <= high {
			fits = append(fits, variant)
		}
	}
	if len(fits) == 0 {
		for i := range variants {
			if variants[i].Status == 1 {
				return &variants[i]
			}
		}
		return nil
	}
	totalWeight := uint32(0)
	for _, variant := range fits {
		if variant.Weight == 0 {
			variant.Weight = 1
		}
		totalWeight += variant.Weight
	}
	pick := uint32(now.UnixNano() % int64(totalWeight))
	for i := range fits {
		weight := fits[i].Weight
		if weight == 0 {
			weight = 1
		}
		if pick < weight {
			return &fits[i]
		}
		pick -= weight
	}
	return &fits[0]
}

func validateSkillConfig(skill model.SkillConfig) error {
	switch skill.EffectType {
	case EffectDamage, EffectHeal, EffectAttackBuff, EffectDefenseBuff, EffectDefend:
	default:
		return ErrBattleConfigMissing
	}
	switch skill.TargetType {
	case "enemy", "ally", "ally_lowest", "self":
	default:
		return ErrBattleConfigMissing
	}
	return nil
}

func selectFriendlyTarget(state *BattleState, targetType string, targetID string) *BattleUnit {
	switch targetType {
	case "self":
		return findUnit(state.PlayerUnits, state.CurrentActorID)
	case "ally_lowest":
		return lowestHPUnit(state.PlayerUnits)
	case "ally":
		target := findUnit(state.PlayerUnits, targetID)
		if target != nil && target.Alive {
			return target
		}
	}
	return nil
}

func targetsForSkill(actor BattleUnit, state BattleState) []string {
	switch actor.SkillTarget {
	case "enemy":
		return aliveUnitIDs(state.EnemyUnits)
	case "ally":
		return aliveUnitIDs(state.PlayerUnits)
	case "ally_lowest":
		if unit := lowestHPUnit(state.PlayerUnits); unit != nil {
			return []string{unit.ID}
		}
	case "self":
		return []string{actor.ID}
	}
	return nil
}

func findUnit(units []BattleUnit, id string) *BattleUnit {
	for i := range units {
		if units[i].ID == id {
			return &units[i]
		}
	}
	return nil
}

func lowestHPUnit(units []BattleUnit) *BattleUnit {
	var selected *BattleUnit
	for i := range units {
		if !units[i].Alive {
			continue
		}
		if selected == nil || units[i].HP < selected.HP {
			selected = &units[i]
		}
	}
	return selected
}

func aliveUnitIDs(units []BattleUnit) []string {
	ids := make([]string, 0, len(units))
	for _, unit := range units {
		if unit.Alive {
			ids = append(ids, unit.ID)
		}
	}
	return ids
}

func allDefeated(units []BattleUnit) bool {
	for _, unit := range units {
		if unit.Alive {
			return false
		}
	}
	return true
}

func skillIDForHero(cfg model.HeroConfig) uint64 {
	role := strings.ToLower(cfg.Role)
	switch {
	case cfg.ID == 1 || strings.Contains(role, "warrior") || strings.Contains(cfg.Role, "战"):
		return 1
	case cfg.ID == 2 || strings.Contains(role, "tank") || strings.Contains(cfg.Role, "坦"):
		return 2
	case cfg.ID == 3 || strings.Contains(role, "guard") || strings.Contains(cfg.Role, "守"):
		return 3
	case cfg.ID == 4 || strings.Contains(role, "assassin") || strings.Contains(cfg.Role, "刺"):
		return 4
	case cfg.ID == 5 || strings.Contains(role, "support") || strings.Contains(cfg.Role, "辅"):
		return 5
	default:
		return 1
	}
}

func totalBattlePower(units []BattleUnit) uint64 {
	var total uint64
	for _, unit := range units {
		total += uint64(effectiveATK(&unit)*3 + unit.MaxHP/5 + effectiveDEF(&unit)*2 + unit.Level*10 + unit.Star*100)
	}
	return total
}

func buildFailureHint(state BattleState) string {
	enemyHP := uint32(0)
	for _, enemy := range state.EnemyUnits {
		enemyHP += enemy.HP
	}
	if enemyHP > 0 {
		return "挑战未通过：敌方仍有余力，建议提升输出或优先集火低血量敌人。"
	}
	return "挑战未通过：队伍承伤不足，建议带上坦克或治疗英雄。"
}

func encodeBattleState(state BattleState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeBattleState(data string) (BattleState, error) {
	var state BattleState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return state, err
	}
	return state, nil
}

func mustEncodeBattleState(state BattleState) string {
	data, err := encodeBattleState(state)
	if err != nil {
		panic(err)
	}
	return data
}

func mustEncodeBattleResult(result *BattleResult) string {
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func buildBattleResponse(session *model.PlayerBattleSession, state BattleState, result *BattleResult) *BattleResponse {
	if session.Status == model.BattleStatusActive {
		state.refreshTurn()
	} else {
		state.Status = session.Status
		state.AvailableActions = nil
		state.SelectedTargets = nil
		state.SelectableActors = nil
	}
	return &BattleResponse{SessionID: session.ID, StageID: session.StageID, Status: session.Status, ExpiresAt: session.ExpiresAt, State: state, Result: result}
}

func endRoundTick(units []BattleUnit) {
	for i := range units {
		units[i].Defending = false
		if units[i].CooldownLeft > 0 {
			units[i].CooldownLeft--
		}
		filtered := units[i].Buffs[:0]
		for _, buff := range units[i].Buffs {
			if buff.RemainingRounds > 1 {
				buff.RemainingRounds--
				filtered = append(filtered, buff)
			}
		}
		units[i].Buffs = filtered
	}
}

func effectiveATK(unit *BattleUnit) uint32 {
	return addBuffs(unit.ATK, unit.Buffs, "atk")
}

func effectiveDEF(unit *BattleUnit) uint32 {
	return addBuffs(unit.DEF, unit.Buffs, "def")
}

func addBuffs(base uint32, buffs []BattleBuff, stat string) uint32 {
	value := int32(base)
	for _, buff := range buffs {
		if buff.Stat == stat {
			value += buff.Amount
		}
	}
	if value < 1 {
		return 1
	}
	return uint32(value)
}

func actorSkillMultiplier(actor *BattleUnit) uint32 {
	if actor.SkillMultiplier > 0 {
		return actor.SkillMultiplier
	}
	if actor.SkillID == 4 {
		return 170
	}
	if actor.SkillEffect == EffectHeal {
		if actor.SkillID == 5 {
			return 140
		}
		return 120
	}
	if actor.SkillEffect == EffectDamage && actor.SkillID > 0 {
		return 150
	}
	return 100
}

func actorSkillStatDelta(actor *BattleUnit) int32 {
	if actor.SkillStatDelta != 0 {
		return actor.SkillStatDelta
	}
	if actor.SkillEffect == EffectDefenseBuff {
		return 80
	}
	if actor.SkillEffect == EffectAttackBuff {
		return 60
	}
	return 0
}

func actorSkillDuration(actor *BattleUnit) uint32 {
	if actor.SkillDuration > 0 {
		return actor.SkillDuration
	}
	if actor.SkillEffect == EffectAttackBuff {
		return 2
	}
	return 1
}

func skinWithDefaults(skin model.CardSkinConfig, ownerType string, ownerID uint64) model.CardSkinConfig {
	if skin.CardArt == "" {
		skin.CardArt = "/static/assets/cards/enemy-minion.svg"
		if ownerType == "hero" {
			skin.CardArt = "/static/assets/cards/hero-warrior.svg"
		}
	}
	if skin.PortraitArt == "" {
		skin.PortraitArt = "/static/assets/portraits/mountain-ape.svg"
		if ownerType == "hero" {
			skin.PortraitArt = "/static/assets/portraits/sun-wukong.svg"
		}
	}
	skin.AttackAnimation = firstNonEmpty(skin.AttackAnimation, "fx-attack")
	skin.SkillAnimation = firstNonEmpty(skin.SkillAnimation, "fx-skill")
	skin.HitAnimation = firstNonEmpty(skin.HitAnimation, "fx-hit-spark")
	skin.DefeatAnimation = firstNonEmpty(skin.DefeatAnimation, "fx-defeat-smoke")
	skin.IdleAnimation = firstNonEmpty(skin.IdleAnimation, "fx-idle-breathe")
	return skin
}

func animationsFromSkin(skin model.CardSkinConfig) BattleAnimationKeys {
	return BattleAnimationKeys{Attack: skin.AttackAnimation, Skill: skin.SkillAnimation, Hit: skin.HitAnimation, Defeat: skin.DefeatAnimation, Idle: skin.IdleAnimation}
}

func animationsFromEnemy(enemy model.EnemyConfig, skin model.CardSkinConfig) BattleAnimationKeys {
	return BattleAnimationKeys{
		Attack: firstNonEmpty(enemy.AttackAnimation, skin.AttackAnimation),
		Skill:  firstNonEmpty(enemy.SkillAnimation, skin.SkillAnimation),
		Hit:    firstNonEmpty(enemy.HitAnimation, skin.HitAnimation),
		Defeat: firstNonEmpty(enemy.DefeatAnimation, skin.DefeatAnimation),
		Idle:   firstNonEmpty(enemy.IdleAnimation, skin.IdleAnimation),
	}
}

func effectKey(skill model.SkillConfig, fallback string) string {
	return firstNonEmpty(skill.EffectKey, fallback)
}

func animationKey(skill model.SkillConfig, fallback string) string {
	return firstNonEmpty(skill.AnimationKey, fallback)
}

func heroConfigIDs(configs map[uint64]model.HeroConfig) []uint64 {
	ids := make([]uint64, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	return ids
}

func enemyConfigIDs(configs map[uint64]model.EnemyConfig) []uint64 {
	ids := make([]uint64, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	return ids
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func battleStatName(stat string) string {
	if stat == "atk" {
		return "攻击"
	}
	return "防御"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func min32(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
