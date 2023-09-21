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
	"time"
)

// WebSocketServer 结构体将包含有关WebSocket服务器的信息和方法
type WebSocketServer struct {
	upgrader   websocket.Upgrader
	clients    sync.Map
	routes     map[string]func(*Client, *types.Message)
	Rooms      map[string]*Room `json:"rooms"`
	roomsMutex sync.Mutex
	Games      map[string]*Game
	gamesMutex sync.Mutex

	Redis *redis.Client
}

type Game struct {
	ID            string           `json:"id"`
	RoomID        string           `json:"room_id"`
	RoomName      string           `json:"room_name"`
	Round         int              `json:"round"`      // 新增字段，表示当前回合
	MaxRounds     int              `json:"max_rounds"` // 新增字段，表示最大回合数 12
	Players       map[*Client]bool `json:"players"`
	PlayerActions map[*Client]bool `json:"player_actions"` // 新增字段，记录每个玩家是否已操作
	CreatedTime   time.Time        `json:"created_time"`
	EndTime       time.Time        `json:"end_time"`
	CreatedUser   string           `json:"created_user"`
}
type Room struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Players     map[*Client]bool `json:"players"`
	CreatedTime time.Time        `json:"created_time"`
	CreatedUser string           `json:"created_user"`
	IsGameStart bool             `json:"is_game_start"` // 是否已经开始游戏
}

type Client struct {
	conn     *websocket.Conn
	ID       string `json:"id"`
	Room     *Room  `json:"room"`
	Game     *Game  `json:"game"`
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
		Games:  make(map[string]*Game),
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

	fmt.Printf("client connected to server client_id=>%s", clientID)

	// 存储客户端
	server.clients.Store(clientID, client)

	defer func() {
		server.clients.Delete(clientID)
		fmt.Println("======用户退出 ID======>", clientID)
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read Error:", err)
			return
		}
		fmt.Println("gxxxxxxxxpe=>")

		var message types.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("Message Unmarshal Error:", err)
			continue
		}

		fmt.Println("get message message type=>", message.Type, client.ID)
		fmt.Println("get message message data=>", message.Data, client.ID)

		message.From = &types.ClientInfo{
			ID:       client.ID,
			RoomName: client.RoomName,
			Name:     client.Name,
		}
		if message.Type == "join_room" {
			roomID, ok := message.Data["room_id"].(string)
			if !ok {
				roomID = ""
			}
			server.JoinRoom(client, roomID)

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

	roomID := client.Room.ID
	if roomID == "" {
		fmt.Println("Client is not in any room")
		return
	}

	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	if room, ok := server.Rooms[roomID]; ok {
		for client := range room.Players {
			if err := client.conn.WriteJSON(message); err != nil {
				fmt.Println("Broadcast Error:", err)
			}
		}
	} else {
		fmt.Printf("No room found with name: %s\n", roomID)
	}
}

// 系统广播
func (server *WebSocketServer) BroadSysMessage(roomID string, message *types.Message) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()

	room, err := server.GetRoomByIDIFExist(roomID)
	if err != nil {
		fmt.Printf("没有找到对应的房间信息 %s\n", roomID)
		return
	}
	for client := range room.Players {
		if err := client.conn.WriteJSON(message); err != nil {
			fmt.Println("Broadcast Error:", err)
		}
	}
}
func (server *WebSocketServer) SendMessageToClient(client *Client, message *types.Message) {
	if err := client.conn.WriteJSON(message); err != nil {
		fmt.Println("Send Message Error:", err)
	}
}

func (server *WebSocketServer) sendJSON(conn *websocket.Conn, message interface{}) error {
	return conn.WriteJSON(message)
}

func (server *WebSocketServer) JoinRoom(client *Client, roomID string) {
	var err error
	response := &types.Message{}
	if roomID == "" {
		response.Error = "请输入房间的ID..."
		server.SendMessageToClient(client, response)
		return
	}

	room, err := server.GetRoomByIDIFExist(roomID)

	if room == nil {
		room, err = server.CreateRoom(roomID, client)
		server.Rooms[roomID] = room
	}

	room.Players[client] = true
	client.Room = room

	clientList, err := server.GetRoomClients(roomID)

	response.Type = types.JOIN_ROOM
	response.Data = map[string]interface{}{
		"room_id":     roomID,
		"client_list": clientList,
		"room_info":   server.RoomClientToInfo(room),
	}
	response.From = &types.ClientInfo{ID: client.ID}

	if err != nil {
		response.Error = err.Error()
		server.SendMessageToClient(client, response)
		return
	}

	server.BroadcastMessage(client, response)
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
			"game_id":      gameID,
			"round":        game.Round, // 添加回合信息到消息中
			"created_time": game.CreatedTime,
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

	game, exists := server.Games[gameID]
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
