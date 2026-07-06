package service

import "mini-card-game/internal/model"

const MaxHeroStar uint32 = 10

type StarUpCost struct {
	Shard uint32 `json:"shard"`
	Gold  uint64 `json:"gold"`
}

type HeroBattleStats struct {
	MaxHP uint32 `json:"max_hp"`
	ATK   uint32 `json:"atk"`
	DEF   uint32 `json:"def"`
}

var starUpCosts = map[uint32]StarUpCost{
	1: {Shard: 10, Gold: 100},
	2: {Shard: 15, Gold: 300},
	3: {Shard: 20, Gold: 600},
	4: {Shard: 30, Gold: 1000},
	5: {Shard: 40, Gold: 1500},
	6: {Shard: 55, Gold: 2200},
	7: {Shard: 70, Gold: 3000},
	8: {Shard: 90, Gold: 4000},
	9: {Shard: 120, Gold: 5500},
}

func DuplicateHeroShardAmount(quality uint8) uint32 {
	switch {
	case quality >= 5:
		return 3
	case quality == 4:
		return 5
	default:
		return 8
	}
}

func NextStarUpCost(star uint32) (StarUpCost, bool) {
	if star >= MaxHeroStar {
		return StarUpCost{}, false
	}
	cost, ok := starUpCosts[star]
	return cost, ok
}

func CalcHeroBattleStats(cfg model.HeroConfig, level uint32, star uint32) HeroBattleStats {
	if level == 0 {
		level = 1
	}
	if star == 0 {
		star = 1
	}
	hpStar, atkStar, defStar := roleStarGrowth(cfg.Role)
	starSteps := star - 1
	return HeroBattleStats{
		MaxHP: cfg.BaseHP + level*80 + starSteps*hpStar,
		ATK:   cfg.BaseATK + level*8 + starSteps*atkStar,
		DEF:   cfg.BaseDEF + level*4 + starSteps*defStar,
	}
}

func CalcHeroPower(cfg model.HeroConfig, level uint32, star uint32) uint64 {
	stats := CalcHeroBattleStats(cfg, level, star)
	base := uint64(stats.ATK*3 + stats.MaxHP/5 + stats.DEF*2)
	growth := uint64(level) * 10
	return (base + growth) * uint64(cfg.PowerFactor) / 100
}

func roleStarGrowth(role string) (hp uint32, atk uint32, def uint32) {
	switch role {
	case "tank", "鍧﹀厠", "坦克":
		return 180, 14, 18
	case "warrior", "鎴樺＋", "战士":
		return 120, 28, 12
	case "assassin", "鍒哄", "刺客":
		return 90, 34, 8
	case "support", "杈呭姪", "辅助":
		return 130, 14, 14
	case "guardian", "瀹堟姢", "守护":
		return 170, 12, 22
	default:
		return 120, 20, 10
	}
}
