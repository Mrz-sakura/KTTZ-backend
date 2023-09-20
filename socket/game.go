package socket

import (
	"app-bff/pkg/errorss"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

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

	server.gamesMutex.Lock()
	defer server.gamesMutex.Unlock()

	if client.Room.CreatedUser != client.ID {
		return nil, errorss.NewWithCodeMsg(errorss.SOCKET_NO_PERMISSON, "你不是这个房间的创建者,无法开始游戏")
	}

	game := &Game{
		ID:            gameID,
		RoomID:        client.Room.ID,
		Round:         1,  // 初始化回合数为 1
		MaxRounds:     12, // 设置最大回合数为 12
		Players:       make(map[*Client]bool),
		PlayerActions: make(map[*Client]bool),
		CreatedTime:   time.Now(),
		CreatedUser:   client.ID,
	}

	room, err := server.GetRoomByIDIFExist(client.Room.ID)
	if room != nil {
		for player := range room.Players {
			game.Players[player] = true
			game.PlayerActions[player] = false // 初始化玩家行动状态为false
		}
	}

	key := common.GetGameCreatedKey(gameID)
	gameInfo := &types.GameInfo{
		ID:            game.ID,
		RoomID:        game.RoomID,
		Players:       server.GetPlayersByClient(game.Players),
		PlayerActions: server.GetPlayerActionsByClient(game.PlayerActions),
		CreatedTime:   game.CreatedTime,
		CreatedUser:   game.CreatedUser,
	}

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
