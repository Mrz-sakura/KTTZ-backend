package socket

import (
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
)

// 开始投掷
func (server *WebSocketServer) HandlerStartThrows(c *Client, message *types.Message) {
	isCanPlay := server.CheckPlayerCanPlay(c, message)
	if !isCanPlay {
		server.SendMessageToClient(c, message)
		return
	}

	dice, err := server.CreateDiceValue(c, message)

	_, err = server.UpdateGame(c, c.Game)

	message = &types.Message{
		Type: message.Type,
		From: &types.ClientInfo{ID: c.ID},
		Data: map[string]interface{}{
			"dice_values": dice,
		},
	}

	if err != nil {
		message.Error = err.Error()
	}

	// 广播消息
	server.BroadcastMessage(c, message)
}

// 临时更新选中的值
func (server *WebSocketServer) HandleUpdateTmpDiceLocks(c *Client, message *types.Message) {
	server.BroadGameMessage(c.Game, message)
}
func (server *WebSocketServer) CreateDiceValue(c *Client, message *types.Message) (*types.Dice, error) {
	lockedIndexs := utils.MapToSliceInt(message.Data, "locked_indexs")

	fmt.Println(lockedIndexs, "lockedI=======ndexslockedIndexslockedIndexslockedIndexs")
	key := common.GetDiceKey(c.Game.ID, c.Game.Round, c.ID)
	diceValue, err := server.GetDiceValue(key, c)

	if err != nil {
		diceValue = &types.Dice{
			GameID:       c.Game.ID,
			Round:        c.Game.Round,
			Value:        make([]int, 5),
			LockedIndexs: lockedIndexs,
			Frequency:    3,
		}
	}

	diceValue.Frequency--
	c.Game.RoundsInfo.CurrentPlayerActions++

	// 去掉第0个的选项
	for i := 1; i < 6; i++ {
		// 如果当前索引在锁定索引列表中，则不生成新的随机数
		if len(lockedIndexs) > 0 && utils.SContains(lockedIndexs, i) {
			continue
		}
		diceValue.Value[i-1] = rand.Intn(6) + 1
	}

	c.Game.Dice = diceValue
	_, err = server.UpdateGame(c, c.Game)
	if err != nil {
		return nil, err
	}

	err = server.UpdateDice(c, c.Game)
	if err != nil {
		return nil, err
	}

	err = server.UpdateDiceLocks(c, c.Game)
	if err != nil {
		return nil, err
	}

	return diceValue, nil
}

// 更新redis rounds数组里的Dice的值  1:Dice 2:Dice
func (server *WebSocketServer) UpdateDice(c *Client, game *Game) error {
	key := common.GetDiceRoundsKey(game.ID, c.ID)

	diceData, err := json.Marshal(game.Dice)
	if err != nil {
		return err
	}

	dcKey := common.GetDiceKey(c.Game.ID, c.Game.Round, c.ID)
	err = server.Redis.HSet(context.Background(), dcKey, game.Dice.Frequency, diceData).Err()
	if err != nil {
		return err
	}

	return server.Redis.HSet(context.Background(), key, game.Round, diceData).Err()
}

func (server *WebSocketServer) GetDiceValue(key string, c *Client) (*types.Dice, error) {
	diceValue, err := server.Redis.HGet(context.Background(), key, strconv.Itoa(c.Game.Round)).Result()
	if err != nil {
		return nil, err
	}

	dice := &types.Dice{}

	err = json.Unmarshal([]byte(diceValue), dice)
	if err != nil {
		return nil, err
	}

	return dice, nil
}

func (server *WebSocketServer) HandlerUpdateDicelocks(c *Client, message *types.Message) {
	//isCanPlay := server.CheckPlayerCanPlay(c, message)
	//if !isCanPlay {
	//	server.SendMessageToClient(c, message)
	//	return
	//}
	lockedIndexs := utils.MapToSliceInt(message.Data, "locked_indexs")

	c.Game.Dice.LockedIndexs = lockedIndexs
	err := server.UpdateDiceLocks(c, c.Game)

	if err != nil {
		message.Error = err.Error()
		server.SendMessageToClient(c, message)
		return
	}

	_, err = server.UpdateGame(c, c.Game)

	if err != nil {
		message.Error = err.Error()
		server.SendMessageToClient(c, message)
		return
	}
}

// 更新骰子锁,按回合更新
func (server *WebSocketServer) UpdateDiceLocks(c *Client, game *Game) error {
	key := common.GetDiceRoundsLocksKey(game.ID, game.Round, c.ID)

	diceData, err := json.Marshal(game.Dice.LockedIndexs)
	if err != nil {
		return err
	}

	return server.Redis.HSet(context.Background(), key, game.Round, diceData).Err()
}

//func (server *WebSocketServer) GetDiceValue(key string, c *Client) (*types.Dice, error) {
//	diceValue, err := server.Redis.Get(context.Background(), key).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	dice := &types.Dice{}
//
//	err = json.Unmarshal([]byte(diceValue), dice)
//	if err != nil {
//		return nil, err
//	}
//
//	return dice, nil
//}
