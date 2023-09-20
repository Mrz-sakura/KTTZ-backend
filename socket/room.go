package socket

import (
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (server *WebSocketServer) GetRoomClients(roomID string) ([]string, error) {
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

func (server *WebSocketServer) CreateRoom(roomID string, client *Client) (*Room, error) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	room := &Room{
		ID:          roomID,
		Name:        "",
		Players:     make(map[*Client]bool),
		CreatedTime: time.Now(),
		CreatedUser: client.ID,
	}
	// 初始設置一個用戶
	room.Players[client] = true

	server.Rooms[roomID] = room

	key := common.GetRoomCreatedKey(roomID)
	roomInfo := &types.RoomInfo{
		ID:          room.ID,
		Name:        room.Name,
		CreatedUser: room.CreatedUser,
		CreatedTime: room.CreatedTime,
		Players: func() map[string]bool {
			players := make(map[string]bool)
			for player := range room.Players {
				players[player.ID] = true
			}
			return players
		}(),
	}

	setv, err := json.Marshal(roomInfo)
	if err != nil {
		return nil, err
	}

	_, err = server.Redis.Set(context.Background(), key, setv, 0).Result()
	if err != nil {
		return nil, err
	}

	err = server.InsertRoom(roomID, roomInfo)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (server *WebSocketServer) InsertRoom(roomID string, roomInfo *types.RoomInfo) error {
	roomsKey := common.GetRoomListKey()

	roomData, err := json.Marshal(roomInfo)
	if err != nil {
		return err
	}

	return server.Redis.HSet(context.Background(), roomsKey, roomID, roomData).Err()
}

func (server *WebSocketServer) GetRoomList() ([]*types.RoomInfo, error) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	var rooms []*types.RoomInfo

	key := common.GetRoomListKey()

	roomMap, err := server.Redis.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	for _, roomData := range roomMap {
		var room *types.RoomInfo
		err := json.Unmarshal([]byte(roomData), &room)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}
func (server *WebSocketServer) GetRoomMap() (map[string]*types.RoomInfo, error) {

	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	rooms := make(map[string]*types.RoomInfo)

	key := common.GetRoomListKey()

	roomMap, err := server.Redis.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	for _, roomData := range roomMap {
		var room *types.RoomInfo
		err := json.Unmarshal([]byte(roomData), &room)
		if err != nil {
			return nil, err
		}
		rooms[room.ID] = room
	}

	return rooms, nil
}

func (server *WebSocketServer) GetRoomByIDIFExist(roomID string) (*Room, error) {
	var rooms map[string]*types.RoomInfo
	var err error

	room, exists := server.Rooms[roomID]
	if !exists {
		rooms, err = server.GetRoomMap()
		if roomInfo, ok := rooms[roomID]; ok {
			room = &Room{
				ID:      roomInfo.ID,
				Name:    roomInfo.Name,
				Players: server.GetPlayerClientByID(roomInfo.Players),
			}
		}
		server.Rooms[roomID] = room
	}
	return room, err
}
