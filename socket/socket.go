package socket

import (
	"app-bff/mod"
	"app-bff/pkg/utils"
	"app-bff/socket/common"
	"app-bff/socket/types"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

// WebSocketServer 结构体将包含有关WebSocket服务器的信息和方法
type WebSocketServer struct {
	upgrader   websocket.Upgrader
	clients    sync.Map
	routes     map[string]func(*Client, *types.Message)
	Rooms      map[string]*Room `json:"rooms"`
	roomsMutex sync.Mutex
	games      map[string]*Game
	gamesMutex sync.Mutex

	Redis *redis.Client
}

type Game struct {
	ID            string           `json:"id"`
	RoomName      string           `json:"room_name"`
	Round         int              `json:"round"`      // 新增字段，表示当前回合
	MaxRounds     int              `json:"max_rounds"` // 新增字段，表示最大回合数 12
	Players       map[*Client]bool `json:"players"`
	PlayerActions map[*Client]bool `json:"player_actions"` // 新增字段，记录每个玩家是否已操作
}
type Room struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Players map[*Client]bool `json:"players"`
}

type Client struct {
	conn     *websocket.Conn
	ID       string `json:"id"`
	Room     *Room  `json:"room"`
	RoomName string `json:"room_name"`
	Name     string `json:"name"`
	GameID   string `json:"game_id"`
}

// NewWebSocketServer 创建一个新的WebSocketServer实例
func NewWebSocketServer() *WebSocketServer {
	redis, _ := mod.GetRedisClient()
	server := &WebSocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		routes: make(map[string]func(*Client, *types.Message)),
		Rooms:  make(map[string]*Room),
		games:  make(map[string]*Game),
		Redis:  redis,
	}

	server.InitRoutes()

	return server
}

// handleClients 处理WebSocket客户端连接和消息
func (server *WebSocketServer) HandleClients(c *gin.Context) {
	w := c.Writer
	r := c.Request

	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade Error:", err)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	client := &Client{
		conn: conn,
		ID:   clientID,
	}
	server.clients.Store(clientID, client)

	defer func() {
		server.clients.Delete(clientID)
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read Error:", err)
			return
		}

		var message types.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("Message Unmarshal Error:", err)
			continue
		}

		message.From = &types.ClientInfo{
			ID:       client.ID,
			RoomName: client.RoomName,
			Name:     client.Name,
		}

		if message.Type == "join_room" {
			roomID, ok := message.Data["room_id"].(string)
			if ok {
				server.JoinRoom(client, roomID)
			}
		} else if message.Type == "leave_room" {
			roomID, ok := message.Data["room_id"].(string)
			if ok {
				server.LeaveRoom(client, roomID)
			}
		} else if message.Type == "player_action" {
			gameID, ok := message.Data["game_id"].(string)
			if ok {
				server.PlayerAction(gameID, client)
			}
		} else if handler, found := server.routes[message.Type]; found {
			handler(client, &message)
		} else {
			fmt.Printf("No handler found for message type: %s\n", message.Type)
		}
	}
}

func (server *WebSocketServer) BroadcastMessage(client *Client, message *types.Message) {

	roomName := client.RoomName
	if roomName == "" {
		fmt.Println("Client is not in any room")
		return
	}

	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	if room, ok := server.Rooms[roomName]; ok {
		for client := range room.Players {
			if err := client.conn.WriteJSON(message); err != nil {
				fmt.Println("Broadcast Error:", err)
			}
		}
	} else {
		fmt.Printf("No room found with name: %s\n", roomName)
	}
}

func (server *WebSocketServer) sendJSON(conn *websocket.Conn, message interface{}) error {
	return conn.WriteJSON(message)
}

func (server *WebSocketServer) JoinRoom(client *Client, roomID string) {
	var err error

	//rooms,err := server.GetRoomList()
	//if err!= nil {
	//    fmt.Println("GetRoomList Error:", err)
	//    return
	//}
	//server.Rooms =

	server.roomsMutex.Lock()
	room, exists := server.Rooms[roomID]
	if !exists {
		room, err = server.CreateRoom(roomID)
	}
	server.roomsMutex.Unlock()

	room.Players[client] = true
	client.Room = room

	clientList, err := server.GetRoomClients(roomID)

	response := &types.Message{
		Data: map[string]interface{}{
			"room_id":     roomID,
			"client_list": clientList,
		},
	}
	if err != nil {
		response.Error = err.Error()
	}

	server.BroadcastMessage(client, response)
}

func (server *WebSocketServer) LeaveRoom(client *Client, roomName string) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	room, exists := server.Rooms[roomName]
	if exists {
		delete(room.Players, client)
		if len(room.Players) == 0 {
			delete(server.Rooms, roomName) // 删除空房间
		}
		client.Room = nil // 清除客户端的房间引用
	} else {
		fmt.Printf("No room found with name: %s\n", roomName)
	}
}

// 新方法来创建一个游戏
func (server *WebSocketServer) CreateGame(client *Client, message *types.Message) {
	server.gamesMutex.Lock()
	defer server.gamesMutex.Unlock()

	// 生成一个游戏ID（这里使用简单的UUID生成方式，你可以更改为任何其他方式）
	gameID := utils.GenUniqueID(client.ID)

	game := &Game{
		ID:            gameID,
		RoomName:      client.Room.Name,
		Round:         1,  // 初始化回合数为 1
		MaxRounds:     12, // 设置最大回合数为 12
		Players:       make(map[*Client]bool),
		PlayerActions: make(map[*Client]bool),
	}

	// 从房间中获取所有客户端并将他们添加到游戏中
	server.roomsMutex.Lock()
	roomClients, exists := server.Rooms[client.Room.ID]
	if exists {
		for player := range roomClients.Players {
			game.Players[player] = true
			game.PlayerActions[player] = false // 初始化玩家行动状态为false
		}
	}
	server.roomsMutex.Unlock()

	// 将新创建的游戏保存到服务器的游戏映射中
	server.games[gameID] = game

	// 创建一个消息来通知客户端游戏已创建
	response := types.Message{
		Type: types.GAME_CREATED,
		Data: map[string]interface{}{
			"game_id": gameID,
			"round":   game.Round, // 添加回合信息到消息中
		},
		From: &types.ClientInfo{
			ID:       client.ID,
			RoomName: client.RoomName,
			Name:     client.Name,
			GameID:   client.GameID, // 同时也可以在消息中包含游戏ID
		},
	}

	err := common.SetGameCreatedData(game.ID, &types.GameInfo{
		ID:        game.ID,
		RoomName:  game.RoomName,
		Round:     game.Round,
		MaxRounds: game.MaxRounds,
		Players: func() map[string]bool {
			players := make(map[string]bool)
			for player := range game.Players {
				players[player.ID] = true
			}
			return players
		}(),
		PlayerActions: func() map[string]bool {
			players := make(map[string]bool)
			for player := range game.Players {
				players[player.ID] = false
			}
			return players
		}(),
	})

	if err != nil {
		response.Error = err.Error()
	}

	// 广播消息到所有在房间中的客户端
	server.BroadcastMessage(client, &response)
}

func (server *WebSocketServer) StartGameRound(game *Game) {
	// 检查当前回合数是否超过最大回合数
	if game.Round > game.MaxRounds {
		fmt.Println("Maximum rounds reached")
		return
	}

	// 创建一个消息来通知客户端新的回合开始
	message := types.Message{
		Type: "game_round_started",
		Data: map[string]interface{}{
			"game_id": game.ID,
			"round":   game.Round, // 包含当前回合信息
		},
	}

	// 广播消息给所有游戏中的玩家
	for client := range game.Players {
		if err := client.conn.WriteJSON(message); err != nil {
			fmt.Println("Broadcast Error:", err)
		}
	}

	// 递增回合数以准备下一轮
	game.Round++
}

func (server *WebSocketServer) PlayerAction(gameID string, client *Client) {
	server.gamesMutex.Lock()
	defer server.gamesMutex.Unlock()

	game, exists := server.games[gameID]
	if !exists {
		fmt.Printf("No game found with ID: %s\n", gameID)
		return
	}

	game.PlayerActions[client] = true

	_ = common.SetGameCreatedData(game.ID, &types.GameInfo{
		ID:        game.ID,
		RoomName:  game.RoomName,
		Round:     game.Round,
		MaxRounds: game.MaxRounds,
		Players: func() map[string]bool {
			players := make(map[string]bool)
			for player := range game.Players {
				players[player.ID] = true
			}
			return players
		}(),
		PlayerActions: func() map[string]bool {
			players := make(map[string]bool)
			for player, v := range game.PlayerActions {
				players[player.ID] = v
			}
			return players
		}(),
	})

	allPlayersActed := true
	for _, acted := range game.PlayerActions {
		if !acted {
			allPlayersActed = false
			break
		}
	}

	if allPlayersActed {
		// 重置所有玩家的行动状态为false以准备下一轮
		for player := range game.PlayerActions {
			game.PlayerActions[player] = false
		}
		server.StartGameRound(game)
	}
}
