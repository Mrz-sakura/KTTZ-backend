package dice

type NumberRequest struct {
	LockedIndexes []string `json:"locked_indexs"`
}

type DiceState struct {
	DiceValues []int `json:"dice_values"`
}
