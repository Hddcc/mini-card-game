package model

type Reward struct {
	Type  string `json:"type"`
	ID    uint64 `json:"id"`
	Count uint64 `json:"count"`
}
