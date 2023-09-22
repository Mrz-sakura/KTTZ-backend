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
	Ints   int `json:"ints"`
	Reward int `json:"reward"` // 奖励分
	All    int `json:"all"`
	STTH   int `json:"stth"` // 四骰同花
	HL     int `json:"hl"`   // 葫芦
	DS     int `json:"ds"`   // 大顺
	XS     int `json:"xs"`   // 小顺
	KT     int `json:"kt"`   // 快艇
	Sum    int `json:"sum"`  // 总和
}
type DiceScoreValue struct {
	One    bool `json:"one"`
	Two    bool `json:"two"`
	Three  bool `json:"three"`
	Four   bool `json:"four"`
	Five   bool `json:"five"`
	Six    bool `json:"six"`
	Ints   bool `json:"ints"`
	Reward bool `json:"reward"` // 奖励分
	All    bool `json:"all"`
	STTH   bool `json:"stth"` // 四骰同花
	HL     bool `json:"hl"`   // 葫芦
	DS     bool `json:"ds"`   // 大顺
	XS     bool `json:"xs"`   // 小顺
	KT     bool `json:"kt"`   // 快艇
	Sum    bool `json:"sum"`  // 总和
}
