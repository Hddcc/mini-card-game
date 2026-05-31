package service

import (
	"errors"
	"testing"
	"time"

	"mini-card-game/internal/model"
)

func testBattleState(t *testing.T) BattleState {
	t.Helper()
	state, err := buildInitialBattleState(
		[]model.PlayerTeam{{Slot: 1, PlayerHeroID: 1}},
		map[uint64]model.PlayerHero{1: {ID: 1, HeroConfigID: 1, Level: 10, Star: 2}},
		map[uint64]model.HeroConfig{1: {ID: 1, Name: "孙悟空", Role: "战士", BaseHP: 1200, BaseATK: 260, BaseDEF: 110, PowerFactor: 120}},
		map[uint64]model.SkillConfig{1: {ID: 1, Name: "破魔斩", TargetType: "enemy", EffectType: "damage", Multiplier: 150, CostRage: 50, Cooldown: 1, EffectKey: "effect-slash", AnimationKey: "fx-slash"}},
		map[uint64]model.CardSkinConfig{1: {OwnerType: "hero", OwnerID: 1, CardArt: "/hero.svg", PortraitArt: "/hero-portrait.svg", AttackAnimation: "fx-attack", SkillAnimation: "fx-skill", HitAnimation: "fx-hit", DefeatAnimation: "fx-defeat", IdleAnimation: "fx-idle"}},
		[]model.StageEncounterEnemy{{VariantID: 1, EnemyConfigID: 1, Slot: 1, Level: 1, Count: 1, SkillID: 101}},
		&model.StageEncounterVariant{ID: 1, StageID: 1, Name: "测试遭遇", MinPower: 1, MaxPower: 1000, EstimatedPower: 500, Weight: 1, Status: 1},
		map[uint64]model.EnemyConfig{1: {ID: 1, Name: "山猿小妖", Role: "minion", BaseHP: 520, BaseATK: 95, BaseDEF: 35, SkillID: 101}},
		map[uint64]model.SkillConfig{101: {ID: 101, Name: "妖袭", TargetType: "ally_lowest", EffectType: "damage", Multiplier: 100}},
		map[uint64]model.CardSkinConfig{1: {OwnerType: "enemy", OwnerID: 1, CardArt: "/enemy.svg", PortraitArt: "/enemy-portrait.svg", AttackAnimation: "fx-claw", SkillAnimation: "fx-bite", HitAnimation: "fx-hit", DefeatAnimation: "fx-defeat", IdleAnimation: "fx-idle"}},
	)
	if err != nil {
		t.Fatalf("build state: %v", err)
	}
	state.refreshTurn()
	return state
}

func TestBuildInitialBattleStateSpendsNoRuntimeState(t *testing.T) {
	state := testBattleState(t)

	if len(state.PlayerUnits) != 1 || len(state.EnemyUnits) != 1 {
		t.Fatalf("unexpected unit count: players=%d enemies=%d", len(state.PlayerUnits), len(state.EnemyUnits))
	}
	if state.PlayerUnits[0].HP != state.PlayerUnits[0].MaxHP {
		t.Fatalf("player should start at full hp")
	}
	if state.CurrentActorID != "p1" {
		t.Fatalf("current actor = %s", state.CurrentActorID)
	}
	if len(state.SelectableActors) != 1 || state.SelectableActors[0] != "p1" {
		t.Fatalf("selectable actors = %#v", state.SelectableActors)
	}
	if state.PlayerUnits[0].BoardSide != "bottom" || state.EnemyUnits[0].BoardSide != "top" {
		t.Fatalf("unexpected board sides")
	}
	if state.PlayerUnits[0].CardArt == "" || state.PlayerUnits[0].Animations.Attack == "" {
		t.Fatalf("missing visual hooks")
	}
	if len(state.AvailableActions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(state.AvailableActions))
	}
}

func TestApplyPlayerActionRejectsDeadTarget(t *testing.T) {
	state := testBattleState(t)
	state.EnemyUnits[0].Alive = false
	state.EnemyUnits[0].HP = 0

	err := applyPlayerAction(&state, BattleActionInput{Action: ActionNormal, ActorID: "p1", TargetID: "e1_1"})
	if !errors.Is(err, ErrInvalidBattleAction) {
		t.Fatalf("expected invalid action, got %v", err)
	}
}

func TestApplyPlayerActionRejectsInvalidActor(t *testing.T) {
	state := testBattleState(t)

	err := applyPlayerAction(&state, BattleActionInput{Action: ActionNormal, ActorID: "e1_1", TargetID: "e1_1"})
	if !errors.Is(err, ErrInvalidBattleAction) {
		t.Fatalf("expected invalid actor, got %v", err)
	}
}

func TestApplyPlayerSkillCanWinBattle(t *testing.T) {
	state := testBattleState(t)
	state.EnemyUnits[0].HP = 50

	err := applyPlayerAction(&state, BattleActionInput{Action: ActionSkill, ActorID: "p1", TargetID: "e1_1"})
	if err != nil {
		t.Fatalf("apply skill: %v", err)
	}
	if !allDefeated(state.EnemyUnits) {
		t.Fatalf("enemy should be defeated")
	}
	if state.PlayerUnits[0].Rage != 0 {
		t.Fatalf("rage should be spent, got %d", state.PlayerUnits[0].Rage)
	}
	if state.Logs[len(state.Logs)-1].EffectKey == "" || state.Logs[len(state.Logs)-1].AnimationKey == "" {
		t.Fatalf("missing effect metadata")
	}
}

func TestHealingAndBuffSkills(t *testing.T) {
	state := testBattleState(t)
	state.PlayerUnits = append(state.PlayerUnits, BattleUnit{
		ID: "p2", Side: "player", BoardSide: "bottom", BoardSlot: 2, Name: "唐三藏",
		HP: 300, MaxHP: 1000, ATK: 200, DEF: 100, Rage: 100, Alive: true,
		SkillID: 5, SkillName: "佛光回春", SkillTarget: "ally_lowest", SkillEffect: EffectHeal, SkillMultiplier: 140, SkillCostRage: 45,
		SkillEffectKey: "effect-heal", SkillAnimation: "fx-heal", Animations: BattleAnimationKeys{Skill: "fx-heal"},
	})

	err := applyPlayerAction(&state, BattleActionInput{Action: ActionSkill, ActorID: "p2"})
	if err != nil {
		t.Fatalf("heal skill: %v", err)
	}
	if state.PlayerUnits[1].HP <= 300 {
		t.Fatalf("expected healing, hp=%d", state.PlayerUnits[1].HP)
	}

	state.refreshTurn()
	state.PlayerUnits[0].SkillID = 6
	state.PlayerUnits[0].SkillName = "斗战号令"
	state.PlayerUnits[0].SkillTarget = "ally"
	state.PlayerUnits[0].SkillEffect = EffectAttackBuff
	state.PlayerUnits[0].SkillStatDelta = 60
	state.PlayerUnits[0].SkillDuration = 2
	state.PlayerUnits[0].Rage = 100
	err = applyPlayerAction(&state, BattleActionInput{Action: ActionSkill, ActorID: "p1", TargetID: "p2"})
	if err != nil {
		t.Fatalf("attack buff skill: %v", err)
	}
	if effectiveATK(&state.PlayerUnits[1]) <= state.PlayerUnits[1].ATK {
		t.Fatalf("expected attack buff")
	}

	state.refreshTurn()
	state.PlayerUnits[0].SkillEffect = EffectDefenseBuff
	state.PlayerUnits[0].SkillTarget = "self"
	state.PlayerUnits[0].SkillStatDelta = 80
	state.PlayerUnits[0].Rage = 100
	state.PlayerUnits[0].CooldownLeft = 0
	err = applyPlayerAction(&state, BattleActionInput{Action: ActionSkill, ActorID: "p1", TargetID: "p1"})
	if err != nil {
		t.Fatalf("defense buff skill: %v", err)
	}
	if effectiveDEF(&state.PlayerUnits[0]) <= state.PlayerUnits[0].DEF {
		t.Fatalf("expected defense buff")
	}
}

func TestResolveEnemyTurnCanDefeatPlayer(t *testing.T) {
	state := testBattleState(t)
	state.PlayerUnits[0].HP = 20
	state.EnemyUnits[0].ATK = 500

	resolveEnemyTurn(&state)

	if !allDefeated(state.PlayerUnits) {
		t.Fatalf("player should be defeated")
	}
}

func TestChooseEncounterVariantAndFallback(t *testing.T) {
	variants := []model.StageEncounterVariant{
		{ID: 1, StageID: 1, Name: "too high", MinPower: 2000, MaxPower: 3000, EstimatedPower: 2500, Weight: 1, Status: 1},
		{ID: 2, StageID: 1, Name: "fit", MinPower: 400, MaxPower: 700, EstimatedPower: 500, Weight: 1, Status: 1},
	}
	selected := chooseEncounterVariant(variants, 500, time.Unix(0, 0))
	if selected == nil || selected.ID != 2 {
		t.Fatalf("expected fitting variant, got %#v", selected)
	}

	selected = chooseEncounterVariant([]model.StageEncounterVariant{{ID: 3, Status: 1, MinPower: 2000, MaxPower: 3000, EstimatedPower: 2500}}, 500, time.Unix(0, 0))
	if selected == nil || selected.ID != 3 {
		t.Fatalf("expected fallback variant, got %#v", selected)
	}
}

func TestCompactEncounterEnemiesDedupesAndCapsRows(t *testing.T) {
	rows := []model.StageEncounterEnemy{
		{EnemyConfigID: 1, Slot: 1, Level: 2, Count: 3, SkillID: 101},
		{EnemyConfigID: 1, Slot: 1, Level: 2, Count: 3, SkillID: 101},
		{EnemyConfigID: 2, Slot: 2, Level: 2, Count: 3, SkillID: 102},
	}

	compact := compactEncounterEnemies(rows, 4)

	if len(compact) != 2 {
		t.Fatalf("expected 2 compact rows, got %d", len(compact))
	}
	var total uint32
	for _, row := range compact {
		total += row.Count
	}
	if total != 4 {
		t.Fatalf("expected capped total count 4, got %d", total)
	}
	if compact[0].Count != 3 || compact[1].Count != 1 {
		t.Fatalf("unexpected compact counts: %#v", compact)
	}
}

func TestCompactBattleStateEnemiesKeepsBoardReadable(t *testing.T) {
	state := BattleState{Status: model.BattleStatusActive}
	for i := 1; i <= 7; i++ {
		state.EnemyUnits = append(state.EnemyUnits, BattleUnit{
			ID:        "e",
			Side:      "enemy",
			BoardSide: "top",
			BoardSlot: uint8(i),
			Slot:      uint8(i),
			HP:        100,
			Alive:     i != 1,
		})
	}

	changed := compactBattleStateEnemies(&state, 4)

	if !changed {
		t.Fatalf("expected state to be compacted")
	}
	if len(state.EnemyUnits) != 4 {
		t.Fatalf("expected 4 enemies, got %d", len(state.EnemyUnits))
	}
	for _, enemy := range state.EnemyUnits {
		if !enemy.Alive {
			t.Fatalf("expected living enemies to be kept first")
		}
	}
}

func TestTerminalSessionRejectsFurtherActions(t *testing.T) {
	err := ensureActionableSession(&model.PlayerBattleSession{Status: model.BattleStatusWon, ExpiresAt: time.Now().Add(time.Minute)}, time.Now())
	if !errors.Is(err, ErrBattleFinished) {
		t.Fatalf("expected finished error, got %v", err)
	}
}

func TestSpendBattleStamina(t *testing.T) {
	now := time.Now()
	asset := &model.PlayerAsset{PlayerID: 1, Stamina: 10}

	if err := spendBattleStamina(asset, 6, now); err != nil {
		t.Fatalf("spend stamina: %v", err)
	}
	if asset.Stamina != 4 {
		t.Fatalf("stamina = %d", asset.Stamina)
	}

	err := spendBattleStamina(asset, 6, now)
	if !errors.Is(err, ErrNotEnoughStamina) {
		t.Fatalf("expected not enough stamina, got %v", err)
	}
}

func TestExpiredSessionRejectsAction(t *testing.T) {
	err := ensureActionableSession(&model.PlayerBattleSession{Status: model.BattleStatusActive, ExpiresAt: time.Now().Add(-time.Minute)}, time.Now())
	if !errors.Is(err, ErrBattleExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
}
