package types

type DiceValueRequest struct {
	LockedIndexes []string `json:"locked_indexs"`
}
type DiceValue struct {
	Value []int `json:"dice_values"`
}

type DiceScore struct {
	One    int `json:"one"`
	Two    int `json:"two"`
	Three  int `json:"three"`
	Four   int `json:"four"`
	Five   int `json:"five"`
	Six    int `json:"six"`
	Reward int `json:"reward"` // 奖励分
	All    int `json:"all"`
	STTH   int `json:"stth"` // 四骰同花
	HL     int `json:"hl"`   // 葫芦
	DS     int `json:"ds"`   // 大顺
	XS     int `json:"xs"`   // 小顺
	KT     int `json:"kt"`   // 快艇
	Sum    int `json:"sum"`  // 总和
}
