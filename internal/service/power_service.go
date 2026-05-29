package service

func CalcHeroPower(baseATK uint32, baseHP uint32, baseDEF uint32, level uint32, star uint32, factor uint32) uint64 {
	base := uint64(baseATK*3 + baseHP/5 + baseDEF*2)
	growth := uint64(level)*10 + uint64(star)*100
	return (base + growth) * uint64(factor) / 100
}
