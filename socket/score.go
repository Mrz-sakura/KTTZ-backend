package socket

import (
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"strconv"
)

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

func (server *WebSocketServer) HandlerSetScore(c *Client, message *types.Message) {
	// 获取Dice值和分数键
	scoreKey := common.GetScoreKey(c.Game.ID, c.ID)

	// 获取或创建DiceScore
	diceScore, err := server.GetOrCreateDiceScore(scoreKey)
	if err != nil {
		message.Error = err.Error()
		server.SendMessageToClient(c, message)
		return
	}

	// 保存新的DiceScore到Redis
	message = server.SaveDiceScore(c, scoreKey, diceScore, message)

	// 广播消息
	message.Data = common.GenerateData("dice_score", diceScore, nil)

	server.BroadcastMessage(c, message)
}

// GetOrCreateDiceScore 获取或创建一个新的DiceScore
func (server *WebSocketServer) GetOrCreateDiceScore(key string) (*DiceScore, error) {
	val, err := server.Redis.Get(context.Background(), key).Result()
	diceScore := &DiceScore{}
	if err == nil {
		err = json.Unmarshal([]byte(val), diceScore)
		if err != nil {
			return nil, err
		}
	}

	// 如果err不为nil，意味着Redis中没有这个key，所以我们返回一个新的DiceScore
	return diceScore, nil
}

// SaveDiceScore 保存DiceScore到Redis
func (server *WebSocketServer) SaveDiceScore(c *Client, key string, DiceScore *DiceScore, message *types.Message) *types.Message {
	// 选择填到哪个类型上
	diceKey := common.GetDiceKey(c.Game.ID, strconv.Itoa(c.Game.Round), c.ID)
	DiceValue, err := server.GetDiceValue(diceKey, c)

	selectSection := utils.MapToString(message.Data, "select_section")
	switch selectSection {
	case types.ONE:
		if DiceScore.One != 0 {
			message.Error = "已存在一的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 1 {
				DiceScore.One++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.TWO:
		if DiceScore.Two != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Two++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.THREE:
		if DiceScore.Three != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Three++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.FOUR:
		if DiceScore.Four != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Four++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.FIVE:
		if DiceScore.Five != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Five++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.SIX:
		if DiceScore.Six != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Six++
			}
		}
		server.CheckScoreReward(DiceScore)
	case types.ALL:
		if DiceScore.All != 0 {
			break
		}
		sum := 0
		for _, v := range DiceValue.Value {
			sum += v
		}
		DiceScore.All = sum
	case types.STTH:
		if DiceScore.STTH != 0 {
			break
		}
		score := common.CheckSTTH(DiceValue.Value)
		if score == 0 {
			message.Error = "不是四骰同花"
			break
		}
		DiceScore.STTH = score
	case types.HL:
		if DiceScore.HL != 0 {
			break
		}
		score := common.ChcekHL(DiceValue.Value)
		if score == 0 {
			message.Error = "不是葫芦"
			break
		}
		DiceScore.HL = score
	case types.DS:
		if DiceScore.DS != 0 {
			break
		}
		score := common.CheckDS(DiceValue.Value)
		if score == 0 {
			message.Error = "不是大顺"
			break
		}
		DiceScore.DS = score
	case types.XS:
		if DiceScore.XS != 0 {
			break
		}
		score := common.CheckXS(DiceValue.Value)
		if score == 0 {
			message.Error = "不是小顺"
			break
		}
		DiceScore.XS = score
	case types.KT:
		if DiceScore.KT != 0 {
			break
		}
		score := common.CheckKT(DiceValue.Value)
		if score == 0 {
			message.Error = "不是快艇"
			break
		}
		DiceScore.KT = score
	}

	DiceScore.Sum = server.GetScoreSum(DiceScore)

	c.Game.Scores[c] = DiceScore
	_, err = server.UpdateGame(c.Game)

	if err != nil {
		message.Error = err.Error()
		return message
	}

	setv, err := json.Marshal(DiceScore)
	if err != nil {
		message.Error = err.Error()
		return message
	}

	_, err = server.Redis.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		message.Error = err.Error()
		return message
	}

	return message
}

func (server *WebSocketServer) CheckScoreReward(score *DiceScore) {
	var sum int
	if score.Reward != 0 {
		return
	}
	sum = (score.One + score.Two + score.Three + score.Four + score.Five + score.Six)
	if sum >= 63 {
		score.Reward = 35
	}
}

// 获取总和
func (server *WebSocketServer) GetScoreSum(score *DiceScore) int {
	var sum int

	sum = (score.One + score.Two + score.Three + score.Four + score.Five + score.Six + score.STTH + score.Reward + score.HL + score.DS + score.XS + score.KT)

	score.Sum = sum

	return sum
}

func (server *WebSocketServer) GetScoresByClient(ds map[*Client]*DiceScore) map[string]*types.DiceScore {
	scores := make(map[string]*types.DiceScore)
	for k, v := range ds {
		scores[k.ID] = (*types.DiceScore)(v)
	}
	return scores
}
