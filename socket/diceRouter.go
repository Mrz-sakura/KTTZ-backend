package socket

import (
	"app-bff/mod"
	"app-bff/pkg/config"
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
)

func (server *WebSocketServer) StartThrowsHandler(c *Client, msg *types.Message) {
	args := msg.Data
	message := msg

	rc, err := mod.GetRedisClient()
	if err != nil {
		message.Error = err.Error()
	}

	// TODO 1是写死的,后续换成userid
	key := fmt.Sprintf("%s%d", config.GetString("redis_key.dice_key"), 1)

	diceValue := common.GetDiceValue(fmt.Sprintf("%s%d", config.GetString("redis_key.dice_key"), 1), nil)

	locked_indexs := utils.MapToSliceString(args, "locked_indexs")

	for i := 0; i < 5; i++ {
		// 如果当前索引在锁定索引列表中，则不生成新的随机数
		if len(locked_indexs) > 0 && utils.ArrayStrContainsInt(locked_indexs, i) {
			continue
		}
		diceValue.Value[i] = rand.Intn(6) + 1
	}

	setv, err := json.Marshal(diceValue.Value)
	if err != nil {
		message.Error = err.Error()
	}

	_, err = rc.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		message.Error = err.Error()
	}

	message.Data = common.GenerateData("dice_values", diceValue.Value, nil)

	server.BroadcastMessage(c, message)

	server.PlayerAction(c.GameID, c)
}

func (server *WebSocketServer) SetScoreHandler(c *Client, msg *types.Message) {
	args := msg.Data
	message := msg

	rc, err := mod.GetRedisClient()
	if err != nil {
		message.Error = err.Error()
	}
	DiceValue := common.GetDiceValue(fmt.Sprintf("%s%d", config.GetString("redis_key.dice_key"), 1), nil)

	// TODO 1是写死的,后续换成userid
	key := fmt.Sprintf("%s%d", config.GetString("redis_key.dice_score_key"), 1)

	DiceScore := &types.DiceScore{}
	// 获取redis的值,如果没有,代表是新的一轮
	if val, err := rc.Get(context.Background(), key).Result(); err == nil {
		err = json.Unmarshal([]byte(val), DiceScore)
		if err != nil {
			message.Error = err.Error()
		}
	}

	// 选择填到哪个类型上
	select_section := utils.MapToString(args, "select_section")
	switch select_section {
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
		common.CheckReward(DiceScore)
	case types.TWO:
		if DiceScore.Two != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Two++
			}
		}
		common.CheckReward(DiceScore)
	case types.THREE:
		if DiceScore.Three != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Three++
			}
		}
		common.CheckReward(DiceScore)
	case types.FOUR:
		if DiceScore.Four != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Four++
			}
		}
		common.CheckReward(DiceScore)
	case types.FIVE:
		if DiceScore.Five != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Five++
			}
		}
		common.CheckReward(DiceScore)
	case types.SIX:
		if DiceScore.Six != 0 {
			break
		}
		for _, v := range DiceValue.Value {
			if v == 2 {
				DiceScore.Six++
			}
		}
		common.CheckReward(DiceScore)
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

	DiceScore.Sum = common.GetSum(DiceScore)

	setv, err := json.Marshal(DiceScore)
	if err != nil {
		message.Error = err.Error()
	}

	_, err = rc.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		message.Error = err.Error()
	}

	message.Data = common.GenerateData("dice_score", DiceScore, nil)

	server.BroadcastMessage(c, message)
}

//func GetDiceInfoHandler(s *WebSocketServer, c *Client, msg *types.Message) {
//	args := msg.Data
//	message := msg
//
//	rc, err := mod.GetRedisClient()
//	if err != nil {
//		message.Error = err.Error()
//	}
//	//diceKey := common.GetDiceKey()
//	DiceValue := common.GetDiceValue(fmt.Sprintf("%s%d", config.GetString("redis_key.dice_key"), 1), nil)
//
//	// TODO 1是写死的,后续换成userid
//	key := fmt.Sprintf("%s%d", config.GetString("redis_key.dice_score_key"), 1)
//
//	DiceScore := &types.DiceScore{}
//	// 获取redis的值,如果没有,代表是新的一轮
//	if val, err := rc.Get(context.Background(), key).Result(); err == nil {
//		err = json.Unmarshal([]byte(val), DiceScore)
//		if err != nil {
//			message.Error = err.Error()
//		}
//	}
//
//	message.Data = common.GenerateData("dice_score", DiceScore, nil)
//	message.Data = common.GenerateData("dice_values", DiceValue, nil)
//
//	s.BroadcastMessage(c, message)
//}
