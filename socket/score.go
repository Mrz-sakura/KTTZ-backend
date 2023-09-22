package socket

import (
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
)

func (server *WebSocketServer) HandlerSetScore(c *Client, message *types.Message) {
	// 获取Dice值和分数键
	scoreKey := common.GetScoreKey(c.Game.ID, c.ID)
	scoreValueKey := common.GetScoreValueKey(c.Game.ID, c.ID)

	// 获取或创建DiceScore
	diceScore, err := server.GetOrCreateDiceScore(scoreKey)
	diceScoreValue, err := server.GetOrCreateDiceScoreValue(scoreValueKey)
	if err != nil {
		message.Error = err.Error()
		server.SendMessageToClient(c, message)
		return
	}

	// 保存新的DiceScore到Redis
	message = server.SaveDiceScore(c, scoreKey, scoreValueKey, diceScore, diceScoreValue, message)
	if message.Error != "" {
		server.SendMessageToClient(c, message)
		return
	}

	//  拿到分数后设置用户的分数
	c.Game.RoundsInfo.CurrentPlayerScore = diceScore
	c.Game.RoundsInfo.CurrentPlayerScoreValue = diceScoreValue
	server.CheckPlayerRoundsCompletedAndHasScore(c, c.Game)

	// 广播消息
	message.Data = common.GenerateData("dice_score", diceScore, nil)

	server.BroadGameMessage(c.Game, message)
}

// GetOrCreateDiceScore 获取或创建一个新的DiceScore
func (server *WebSocketServer) GetOrCreateDiceScore(key string) (*types.DiceScore, error) {
	val, err := server.Redis.Get(context.Background(), key).Result()
	diceScore := &types.DiceScore{}
	if err == nil {
		err = json.Unmarshal([]byte(val), diceScore)
		if err != nil {
			return nil, err
		}
	}

	// 如果err不为nil，意味着Redis中没有这个key，所以我们返回一个新的DiceScore
	return diceScore, nil
}

// 保存用户的操作记录,比如记录了;一个全选等等
func (server *WebSocketServer) GetOrCreateDiceScoreValue(key string) (*types.DiceScoreValue, error) {
	val, err := server.Redis.Get(context.Background(), key).Result()
	diceScoreValue := &types.DiceScoreValue{}
	if err == nil {
		err = json.Unmarshal([]byte(val), diceScoreValue)
		if err != nil {
			return nil, err
		}
	}

	return diceScoreValue, nil
}

// SaveDiceScore 保存DiceScore到Redis
func (server *WebSocketServer) SaveDiceScore(c *Client, key string, svKey string, DiceScore *types.DiceScore, DiceScoreValue *types.DiceScoreValue, message *types.Message) *types.Message {
	var err error
	// 选择填到哪个类型上
	drkey := common.GetDiceRoundsKey(c.Game.ID, c.ID)
	DiceValue, err := server.GetDiceValue(drkey, c)
	if err != nil {
		message.Error = err.Error()
		return message
	}

	selectSection := utils.MapToString(message.Data, "select_section")
	switch selectSection {
	case types.ONE:
		if DiceScoreValue.One != false {
			message.Error = "已存在一的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 1 {
				DiceScore.One += v
			}
		}
		DiceScoreValue.One = true
		server.CheckScoreReward(DiceScore)
	case types.TWO:
		if DiceScoreValue.Two != false {
			message.Error = "已存在二的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Two += v
			}
		}
		DiceScoreValue.Two = true
		server.CheckScoreReward(DiceScore)
	case types.THREE:
		if DiceScoreValue.Three != false {
			message.Error = "已存在三的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 3 {
				DiceScore.Three += v
			}
		}
		DiceScoreValue.Three = true
		server.CheckScoreReward(DiceScore)
	case types.FOUR:
		if DiceScoreValue.Four != false {
			message.Error = "已存在四的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 4 {
				DiceScore.Four += v
			}
		}
		DiceScoreValue.Four = true
		server.CheckScoreReward(DiceScore)
	case types.FIVE:
		if DiceScoreValue.Five != false {
			message.Error = "已存在五的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 5 {
				DiceScore.Five += v
			}
		}
		DiceScoreValue.Five = true
		server.CheckScoreReward(DiceScore)
	case types.SIX:
		if DiceScoreValue.Six != false {
			message.Error = "已存在六的点数"
			break
		}
		for _, v := range DiceValue.Value {
			if v == 6 {
				DiceScore.Six += v
			}
		}
		DiceScoreValue.Six = true
		server.CheckScoreReward(DiceScore)
	case types.ALL:
		if DiceScoreValue.All != false {
			message.Error = "已存在全选的点数"
			break
		}
		sum := 0
		for _, v := range DiceValue.Value {
			sum += v
		}
		DiceScoreValue.All = true
		DiceScore.All = sum
	case types.STTH:
		if DiceScoreValue.STTH != false {
			message.Error = "已存在四骰同花的点数"
			break
		}
		score := common.CheckSTTH(DiceValue.Value)
		//if score == 0 {
		//	message.Error = "不是四骰同花"
		//	break
		//}
		DiceScoreValue.STTH = true
		DiceScore.STTH = score
	case types.HL:
		if DiceScoreValue.HL != false {
			message.Error = "已存在葫芦的点数"
			break
		}
		score := common.ChcekHL(DiceValue.Value)
		//if score == 0 {
		//	message.Error = "不是葫芦"
		//	break
		//}
		DiceScoreValue.HL = true
		DiceScore.HL = score
	case types.DS:
		if DiceScoreValue.DS != false {
			message.Error = "已存在大顺的点数"
			break
		}
		score := common.CheckDS(DiceValue.Value)
		//if score == 0 {
		//	message.Error = "不是大顺"
		//	break
		//}
		DiceScoreValue.DS = true
		DiceScore.DS = score
	case types.XS:
		if DiceScoreValue.XS != false {
			message.Error = "已存在小顺的点数"
			break
		}
		score := common.CheckXS(DiceValue.Value)
		//if score == 0 {
		//	message.Error = "不是小顺"
		//	break
		//}
		DiceScoreValue.XS = true
		DiceScore.XS = score
	case types.KT:
		if DiceScoreValue.KT != false {
			message.Error = "已存在快艇的点数"
			break
		}
		score := common.CheckKT(DiceValue.Value)
		//if score == 0 {
		//	message.Error = "不是快艇"
		//	break
		//}
		DiceScoreValue.KT = true
		DiceScore.KT = score
	}

	if message.Error != "" {
		return message
	}

	DiceScore.Sum = server.GetScoreSum(DiceScore)

	c.Game.Scores[c] = DiceScore
	c.Game.ScoresValue[c] = DiceScoreValue

	_, err = server.UpdateGame(c, c.Game)

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

	err = server.saveDiceScoreValue(DiceScoreValue, svKey)
	if err != nil {
		message.Error = err.Error()
		return message
	}

	// 广播score更新到游戏房间
	bm := &types.Message{
		Type: types.SCORE_UPDATE,
		Data: map[string]interface{}{
			"dice_score": DiceScore,
			"game_info":  server.GameClientToInfo(c.Game),
		},
		From: &types.ClientInfo{
			ID:     c.ID,
			GameID: c.Game.ID,
		},
	}

	server.BroadGameMessage(c.Game, bm)

	return message
}
func (server *WebSocketServer) saveDiceScoreValue(DiceScoreValue *types.DiceScoreValue, key string) error {
	setv, err := json.Marshal(DiceScoreValue)
	if err != nil {
		return err
	}

	_, err = server.Redis.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func (server *WebSocketServer) CheckScoreReward(score *types.DiceScore) {
	var sum int
	if score.Reward != 0 {
		return
	}
	sum = (score.One + score.Two + score.Three + score.Four + score.Five + score.Six)
	score.Ints = sum

	if sum >= 63 {
		score.Reward = 35
	}
}

// 获取总和
func (server *WebSocketServer) GetScoreSum(score *types.DiceScore) int {
	var sum int

	sum = (score.One + score.Two + score.Three + score.Four + score.Five + score.Six + score.STTH + score.Reward + score.HL + score.DS + score.XS + score.KT)

	score.Sum = sum

	return sum
}

func (server *WebSocketServer) GetScoresByClient(ds map[*Client]*types.DiceScore) map[string]*types.DiceScore {
	scores := make(map[string]*types.DiceScore)
	for k, v := range ds {
		scores[k.ID] = (v)
	}
	return scores
}
func (server *WebSocketServer) GetScoresValueByClient(ds map[*Client]*types.DiceScoreValue) map[string]*types.DiceScoreValue {
	scores := make(map[string]*types.DiceScoreValue)
	for k, v := range ds {
		scores[k.ID] = v
	}
	return scores
}
