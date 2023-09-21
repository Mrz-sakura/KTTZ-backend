package socket

import (
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
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

	if c.Game.RoundsInfo.CurrentPlayerActions >= 3 {
		// 发送一个信号到CompletedChan，表示该玩家已完成回合
		c.Game.RoundsInfo.CompletedChan <- c

		// 重置当前玩家和行动次数，以便在NoticePlayersRound中选择下一个玩家
		c.Game.RoundsInfo.CurrentPlayer = nil
		c.Game.RoundsInfo.CurrentPlayerActions = 0
	}

	_, err = server.UpdateGame(c.Game)

	message = &types.Message{
		Type: message.Type,
		From: &types.ClientInfo{ID: c.ID},
	}

	message.Data = common.GenerateData("dice_values", dice, nil)

	if err != nil {
		message.Error = err.Error()
	}

	// 广播消息
	server.BroadcastMessage(c, message)

}

func (server *WebSocketServer) CreateDiceValue(c *Client, message *types.Message) (*types.Dice, error) {
	lockedIndexs := utils.MapToSliceInt(message.Data, "locked_indexs")

	key := common.GetDiceKey(c.Game.ID, strconv.Itoa(c.Game.Round), c.ID)
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

	for i := 0; i < 5; i++ {
		// 如果当前索引在锁定索引列表中，则不生成新的随机数
		if len(lockedIndexs) > 0 && utils.SContains(lockedIndexs, i) {
			continue
		}
		diceValue.Value[i] = rand.Intn(6) + 1
	}

	c.Game.Dice = (*Dice)(diceValue)
	_, err = server.UpdateGame(c.Game)
	if err != nil {
		return nil, err
	}

	err = server.UpdateDice(c, c.Game)
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

	return server.Redis.HSet(context.Background(), key, game.Round, diceData).Err()
}

func (server *WebSocketServer) GetDiceValue(key string, c *Client) (*types.Dice, error) {
	diceValue, err := server.Redis.HGet(context.Background(), key, strconv.Itoa(c.Game.Round)).Result()
	if err != nil {
		return nil, err
	}

	var dice *types.Dice

	err = json.Unmarshal([]byte(diceValue), dice)
	if err != nil {
		return nil, err
	}

	return dice, nil
}
