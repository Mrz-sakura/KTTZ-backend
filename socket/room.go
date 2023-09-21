package socket

import (
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

func (server *WebSocketServer) HandleGetRoomList(client *Client, message *types.Message) {
	var err error
	fmt.Println("do roomList:")

	roomList, err := server.GetRoomList()

	fmt.Println("roomList:", roomList, err)
	// 创建一个消息来通知客户端游戏已创建
	response := types.Message{
		Type: types.ROOMLIST,
		Data: map[string]interface{}{
			"room_list": roomList,
		},
		From: &types.ClientInfo{
			ID: client.ID,
		},
	}

	if err != nil {
		response.Error = err.Error()
	}

	// 广播消息到所有在房间中的客户端
	server.SendMessageToClient(client, &response)
}

func (server *WebSocketServer) HandleCreateRoom(client *Client, message *types.Message) {
	var err error
	roomID := utils.MapToString(message.Data, "room_id")

	response := &types.Message{}
	if roomID == "" {
		response.Error = "请输入房间的ID..."
		server.SendMessageToClient(client, response)
		return
	}

	room, err := server.GetRoomByIDIFExist(roomID)

	if room != nil {
		server.SendMessageToClient(client, response)
		return
	}

	room, err = server.CreateRoom(roomID, client)
	server.Rooms[roomID] = room

	room.Players[client] = true
	client.Room = room

	clientList, err := server.GetRoomClients(roomID)

	response.Type = types.ROOM_CREATED

	rooms, err := server.GetRoomList()
	response.Data = map[string]interface{}{
		"room_id":     roomID,
		"client_list": clientList,
		"room_list":   rooms,
	}
	response.From = &types.ClientInfo{ID: client.ID}

	if err != nil {
		response.Error = err.Error()
		server.SendMessageToClient(client, response)
		return
	}

	server.BroadcastMessage(client, response)
}

func (server *WebSocketServer) GetRoomClients(roomID string) ([]string, error) {
	if roomID == "" {
		return nil, fmt.Errorf("room name cannot be empty")
	}

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
	room := &Room{
		ID:          roomID,
		Name:        "",
		Players:     make(map[*Client]bool),
		CreatedTime: time.Now(),
		CreatedUser: client.ID,
		IsGameStart: false,
	}
	// 初始設置一個用戶
	room.Players[client] = true

	server.Rooms[roomID] = room

	key := common.GetRoomCreatedKey(roomID)
	roomInfo := server.RoomClientToInfo(room)

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
func (server *WebSocketServer) DeleteRoom(roomID string) error {
	roomsKey := common.GetRoomListKey()
	key := common.GetRoomCreatedKey(roomID)
	var wg sync.WaitGroup

	// 删除内存中的房间
	delete(server.Rooms, roomID)

	// 创建一个错误通道
	errCh := make(chan error, 2)
	wg.Add(2) // 因为有两个goroutine，所以计数设置为2

	// 删除第一个key
	go func() {
		defer wg.Done()
		err := server.Redis.Del(context.Background(), key).Err()
		errCh <- err
	}()

	// 删除第二个key
	go func() {
		defer wg.Done()
		err := server.Redis.HDel(context.Background(), roomsKey, roomID).Err()
		errCh <- err
	}()

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// 收集错误
	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// 判断是否有错误
	if len(errs) > 0 {
		return fmt.Errorf("deletion errors: %v", errs)
	}

	return nil
}

func (server *WebSocketServer) LeaveRoom(client *Client, message *types.Message) {
	//server.roomsMutex.Lock()
	//defer server.roomsMutex.Unlock()
	response := &types.Message{}
	roomID := utils.MapToString(message.Data, "room_id")

	if client.Room == nil || client.Room.ID == "" {
		response.Error = "您当前不在房间"
		server.SendMessageToClient(client, response)
		return
	}

	room, err := server.GetRoomByIDIFExist(roomID)
	if err != nil {
		response.Error = err.Error()
		server.SendMessageToClient(client, response)
		return
	}

	delete(room.Players, client)
	if len(room.Players) == 0 {
		err = server.DeleteRoom(roomID)
		if err != nil {
			response.Error = err.Error()
			server.SendMessageToClient(client, response)
			return
		}
	}

	client.Room = nil // 清除客户端的房间引用

	response.Type = types.LEAVE_ROOM
	response.From = &types.ClientInfo{ID: client.ID}

	server.SendMessageToClient(client, response)
	server.BroadSysMessage(roomID, response)
}

func (server *WebSocketServer) GetRoomList() ([]*types.RoomInfo, error) {

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
				ID:          roomInfo.ID,
				Name:        roomInfo.Name,
				Players:     server.GetPlayerClientByID(roomInfo.Players),
				CreatedTime: roomInfo.CreatedTime,
				CreatedUser: roomInfo.CreatedUser,
			}
		}
		server.Rooms[roomID] = room
	}
	return room, err
}
func (server *WebSocketServer) GetRoomInfoByID(roomID string) (*types.RoomInfo, error) {
	var rooms map[string]*types.RoomInfo
	var resp *types.RoomInfo
	var err error

	// 获取原始的客户端的room
	room, exists := server.Rooms[roomID]
	if !exists {
		rooms, err = server.GetRoomMap()
		resp, _ = rooms[roomID]
	}

	if resp == nil {
		resp = server.RoomClientToInfo(room)
	}

	return resp, err
}

func (server *WebSocketServer) RoomClientToInfo(roomClient *Room) *types.RoomInfo {
	return &types.RoomInfo{
		ID:          roomClient.ID,
		Name:        roomClient.Name,
		Players:     server.GetPlayersByClient(roomClient.Players),
		CreatedTime: roomClient.CreatedTime,
		CreatedUser: roomClient.CreatedUser,
		IsGameStart: roomClient.IsGameStart,
	}
}
