package random

import "math/rand"

func WeightIndex(weights []uint32) int {
	var total uint32
	for _, weight := range weights {
		total += weight
	}

	if total == 0 {
		return -1
	}

	point := uint32(rand.Intn(int(total))) + 1
	var current uint32
	for i, weight := range weights {
		current += weight
		if point <= current {
			return i
		}
	}

	return -1
}