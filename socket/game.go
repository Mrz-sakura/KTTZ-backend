package socket

import (
	"app-bff/pkg/errorss"
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Game struct {
	ID            string                 `json:"id"`
	RoomID        string                 `json:"room_id"`
	RoomName      string                 `json:"room_name"`
	Round         int                    `json:"round"`      // 新增字段，表示当前回合
	MaxRounds     int                    `json:"max_rounds"` // 新增字段，表示最大回合数 12
	Players       map[*Client]bool       `json:"players"`
	PlayerActions map[*Client]bool       `json:"player_actions"` // 新增字段，记录每个玩家是否已操作
	CreatedTime   time.Time              `json:"created_time"`
	IsEnd         bool                   `json:"is_end"`
	EndTime       time.Time              `json:"end_time"`
	CreatedUser   string                 `json:"created_user"`
	Dice          *Dice                  `json:"dices"`
	RoundsInfo    *Rounds                `json:"rounds_info"`
	Scores        map[*Client]*DiceScore `json:"scores"`
}

type Rounds struct {
	CurrentPlayer        *Client   `json:"current_player"` // 当前正在游戏回合的玩家
	CurrentPlayerActions int       `json:"current_player_actions"`
	CompletedPlayer      []*Client `json:"completed_player"`
	IsCompleted          bool      `json:"is_completed"`
	CompletedChan        chan *Client
}

// 新方法来创建一个游戏
func (server *WebSocketServer) HandleCreateGame(client *Client, message *types.Message) {
	var err error

	// 生成一个游戏ID（这里使用简单的UUID生成方式，你可以更改为任何其他方式）
	gameID := utils.GenUniqueID(client.ID)
	// 将新创建的游戏保存到服务器的游戏映射中
	game, err := server.CreateGame(gameID, client)

	// 从房间中获取所有客户端并将他们添加到游戏中
	server.Games[gameID] = game
	client.Game = game

	// 创建一个消息来通知客户端游戏已创建
	response := types.Message{
		Type: types.GAME_CREATED,
		Data: map[string]interface{}{
			"game_id":   gameID,
			"game_info": server.GameClientToInfo(game),
		},
		From: &types.ClientInfo{
			ID:     client.ID,
			RoomID: client.Room.ID,
			Name:   client.Name,
			GameID: gameID, // 同时也可以在消息中包含游戏ID
		},
	}

	if err != nil {
		response.Error = err.Error()
	}

	// 广播消息到所有在房间中的客户端
	server.BroadcastMessage(client, &response)
	// 两秒后提示开始第一回合的游戏
	time.Sleep(time.Second * 2)

	server.StartGameRound(client, game)
}

//func (server *WebSocketServer) HandleGetGameOne(client *Client, message *types.Message) {
//	var err error
//
//	id := utils.MapToString(message.Data, "game_id")
//	roomList, err := server.GetGameByID(id)
//
//	// 创建一个消息来通知客户端游戏已创建
//	response := types.Message{
//		Type: types.ROOMLIST,
//		Data: map[string]interface{}{
//			"room_list": roomList,
//		},
//		From: &types.ClientInfo{
//			ID: client.ID,
//		},
//	}
//
//	if err != nil {
//		response.Error = err.Error()
//	}
//
//	// 广播消息到所有在房间中的客户端
//	server.SendMessageToClient(client, &response)
//}

func (server *WebSocketServer) GetGameClients(roomID string) ([]string, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room name cannot be empty")
	}

	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	room, exists := server.Rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("no room found with name: %s", roomID)
	}

	clientList := make([]string, 0, len(room.Players))
	for roomClient := range room.Players {
		clientList = append(clientList, roomClient.ID)
	}

	return clientList, nil
}

func (server *WebSocketServer) CreateGame(gameID string, client *Client) (*Game, error) {

	if client.Room.CreatedUser != client.ID {
		return nil, errorss.NewWithCodeMsg(errorss.SOCKET_NO_PERMISSON, "你不是这个房间的创建者,无法开始游戏")
	}

	game := &Game{
		ID:            gameID,
		RoomID:        client.Room.ID,
		Round:         0,  // 初始化回合数为 0
		MaxRounds:     12, // 设置最大回合数为 12
		Players:       make(map[*Client]bool),
		PlayerActions: make(map[*Client]bool),
		CreatedTime:   time.Now(),
		CreatedUser:   client.ID,
		RoundsInfo: &Rounds{
			CurrentPlayer:        nil,
			CurrentPlayerActions: 0,
			CompletedPlayer:      make([]*Client, 0),
			IsCompleted:          false,
			CompletedChan:        make(chan *Client),
		},
	}

	room, err := server.GetRoomByIDIFExist(client.Room.ID)
	if room != nil {
		for player := range room.Players {
			game.Players[player] = true
			game.PlayerActions[player] = false // 初始化玩家行动状态为false
		}
	}

	key := common.GetGameCreatedKey(gameID)
	gameInfo := server.GameClientToInfo(game)

	setv, err := json.Marshal(gameInfo)
	if err != nil {
		return nil, errorss.NewWithCode(errorss.DATA_PARSE_ERROR)
	}

	_, err = server.Redis.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		return nil, errorss.NewWithCode(errorss.SOCKET_REDIS_ERROR)
	}

	err = server.InsertGame(gameID, gameInfo)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (server *WebSocketServer) InsertGame(gameID string, gameInfo *types.GameInfo) error {
	gamesKey := common.GetRoomListKey()

	gameData, err := json.Marshal(gameInfo)
	if err != nil {
		return errorss.NewWithCode(errorss.DATA_PARSE_ERROR)
	}

	err = server.Redis.HSet(context.Background(), gamesKey, gameID, gameData).Err()
	if err != nil {
		return errorss.NewWithCode(errorss.SOCKET_REDIS_ERROR)
	}

	return nil
}
func (server *WebSocketServer) UpdateGame(game *Game) (*types.GameInfo, error) {
	server.Games[game.ID] = game

	gamesKey := common.GetGameListKey()
	key := common.GetGameCreatedKey(game.ID)

	gameInfo := server.GameClientToInfo(game)
	gameData, err := json.Marshal(gameInfo)
	if err != nil {
		return nil, err
	}

	_, err = server.Redis.Set(context.Background(), key, gameData, 0).Result()
	if err != nil {
		return nil, err
	}

	err = server.Redis.HSet(context.Background(), gamesKey, game.ID, gameData).Err()
	return gameInfo, err
}

func (server *WebSocketServer) GetGameList() ([]*types.GameInfo, error) {
	server.gamesMutex.Lock()
	defer server.gamesMutex.Unlock()

	var games []*types.GameInfo

	key := common.GetGameListKey()

	gameMap, err := server.Redis.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	for _, gameData := range gameMap {
		var game *types.GameInfo
		err := json.Unmarshal([]byte(gameData), &game)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}
func (server *WebSocketServer) GetGameMap() (map[string]*types.GameInfo, error) {

	server.gamesMutex.TryLock()
	defer server.gamesMutex.Unlock()

	games := make(map[string]*types.GameInfo)

	key := common.GetGameListKey()

	gameMap, err := server.Redis.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	for _, gameData := range gameMap {
		var game *types.GameInfo
		err := json.Unmarshal([]byte(gameData), &game)
		if err != nil {
			return nil, err
		}
		games[game.ID] = game
	}

	return games, nil
}
func (server *WebSocketServer) GetGameByID() (*types.GameInfo, error) {
	server.gamesMutex.Lock()
	defer server.gamesMutex.Unlock()

	key := common.GetGameListKey()

	gameData, err := server.Redis.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var game *types.GameInfo
	err = json.Unmarshal([]byte(gameData), &game)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (server *WebSocketServer) StartGameRound(c *Client, game *Game) {
	// 递增回合数以准备下一轮
	game.Round++

	// 检查当前回合数是否超过最大回合数
	if game.Round > game.MaxRounds {
		game.EndTime = time.Now() // 设置游戏结束时间
		game.IsEnd = true

		_, err := server.UpdateGame(game)

		message := &types.Message{
			Type: types.GAME_END,
			Data: map[string]interface{}{
				"game_id":   game.ID,
				"round":     game.Round,     // 包含当前回合信息
				"max_round": game.MaxRounds, // 包含当前回合信息
				"game_info": server.GameClientToInfo(game),
				"message":   fmt.Sprintf("游戏结束啦,我们下局再战~"),
			},
		}

		if err != nil {
			message.Error = err.Error()
		}

		server.BroadGameMessage(c, game, message)
		return
	}

	game.Dice = &Dice{
		GameID:       game.ID,
		Round:        game.Round,
		Value:        nil,
		LockedIndexs: nil,
		Frequency:    3,
	}

	// 创建一个消息来通知客户端新的回合开始
	message := &types.Message{
		Type: types.GAME_ROUND_START,
		Data: map[string]interface{}{
			"game_id":   game.ID,
			"round":     game.Round,     // 包含当前回合信息
			"max_round": game.MaxRounds, // 包含当前回合信息
			"game_info": server.GameClientToInfo(game),
			"message":   fmt.Sprintf("第%d回合的游戏开始~", game.Round),
		},
	}

	err := server.UpdateDice(c, game)
	if err != nil {
		message.Error = err.Error()
	}

	_, err = server.UpdateGame(game)
	if err != nil {
		message.Error = err.Error()
	}

	server.BroadGameMessage(c, game, message)

	time.Sleep(time.Second * 2)

	// 开始通知玩家的回合开始
	go server.NoticeNextPlayersRound(game)

	// 另一个goroutine用于检测所有玩家是否完成了回合
	go server.CheckPlayerRoundsCompleted(c, game)
}

func (server *WebSocketServer) GameClientToInfo(game *Game) *types.GameInfo {
	fmt.Println(game.RoundsInfo)
	gameInfo := &types.GameInfo{
		ID:            game.ID,
		RoomID:        game.RoomID,
		Players:       server.GetPlayersByClient(game.Players),
		PlayerActions: server.GetPlayerActionsByClient(game.PlayerActions),
		CreatedTime:   game.CreatedTime,
		CreatedUser:   game.CreatedUser,
		Dice:          (*types.Dice)(game.Dice),
		Round:         game.Round,
		MaxRounds:     game.MaxRounds,
		EndTime:       game.EndTime,
		IsEnd:         game.IsEnd,
		RoundsInfo: &types.RoundsInfo{
			CompletedPlayer: server.GetPlayersSliceByClient(game.RoundsInfo.CompletedPlayer),
			IsCompleted:     game.RoundsInfo.IsCompleted,
		},
		Scores: server.GetScoresByClient(game.Scores),
	}
	if game.RoundsInfo.CurrentPlayer != nil {
		gameInfo.RoundsInfo.CurrentPlayer = game.RoundsInfo.CurrentPlayer.ID
		gameInfo.RoundsInfo.CurrentPlayerActions = game.RoundsInfo.CurrentPlayerActions
	}

	return gameInfo
}

func (server *WebSocketServer) NoticeNextPlayersRound(game *Game) {
	for player := range game.Players {
		// 跳过已经完成回合的玩家
		if contains(game.RoundsInfo.CompletedPlayer, player) {
			continue
		}

		// 如果当前玩家为空或行动次数已用完，切换到下一个玩家
		if game.RoundsInfo.CurrentPlayer == nil || game.RoundsInfo.CurrentPlayerActions >= 3 {
			game.RoundsInfo.CurrentPlayer = player
			game.RoundsInfo.CurrentPlayerActions = 0
		}

		// 通知当前玩家他的回合开始
		message := &types.Message{
			Type: types.PLAYER_TURN_START,
			Data: map[string]interface{}{
				"game_id":   game.ID,
				"game_info": game,
				"player_id": game.RoundsInfo.CurrentPlayer.ID,
				"message":   fmt.Sprintf("亲爱的玩家 %s,你的回合开始！", game.RoundsInfo.CurrentPlayer.ID),
			},
		}

		server.SendMessageToClient(game.RoundsInfo.CurrentPlayer, message)
		// 由于我们已经找到了下一个玩家，可以跳出循环
		break
	}
}

// 辅助函数：检查某个玩家是否在已完成回合的玩家列表中
func contains(players []*Client, player *Client) bool {
	for _, p := range players {
		if p == player {
			return true
		}
	}
	return false
}

func (server *WebSocketServer) CheckPlayerRoundsCompleted(c *Client, game *Game) {
	for {
		completedPlayer := <-game.RoundsInfo.CompletedChan
		game.RoundsInfo.CompletedPlayer = append(game.RoundsInfo.CompletedPlayer, completedPlayer)

		_, err := server.UpdateGame(game)
		if err != nil {
			return
		}
		// 所有玩家都完成了这一回合操作
		if len(game.RoundsInfo.CompletedPlayer) >= len(game.Players) {
			game.RoundsInfo.IsCompleted = true

			_, err := server.UpdateGame(game)
			if err != nil {
				return
			}
			fmt.Println("所有玩家都完成了回合")
			// 开始下一个回合
			server.StartGameRound(c, game)
			return
		} else {
			server.NoticeNextPlayersRound(game)
		}
	}
}

func (server *WebSocketServer) CheckPlayerCanPlay(c *Client, message *types.Message) bool {
	// 首先，检查是否是当前玩家的回合
	if c.Game.RoundsInfo.CurrentPlayer != c {
		message.Error = "不是你的回合"
		return true
	}

	// 然后，检查该玩家是否还有剩余的行动次数
	if c.Game.RoundsInfo.CurrentPlayerActions >= 3 {
		message.Error = "你已经用完了本回合的所有行动次数"
		return true
	}
	return false
}
