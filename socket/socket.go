package socket

import (
	"app-bff/mod"
	"app-bff/socket/types"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
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

//type Dice struct {
//	GameID       string `json:"game_id"`
//	Round        int    `json:"round"`         // 轮数
//	Value        []int  `json:"value"`         // 骰子的值
//	LockedIndexs []int  `json:"locked_indexs"` // 本轮锁定的索引
//	Frequency    int    `json:"frequency"`     // 剩余次数
//}

type Room struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Players     map[*Client]bool `json:"players"`
	CreatedTime time.Time        `json:"created_time"`
	CreatedUser string           `json:"created_user"`
	IsGameStart bool             `json:"is_game_start"` // 是否已经开始游戏
}

// NewWebSocketServer 创建一个新的WebSocketServer实例
func NewWebSocketServer() *WebSocketServer {
	rds, err := mod.GetRedisClient()
	if err != nil {
		log.Fatal(err)
	}
	server := &WebSocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		routes: make(map[string]func(*Client, *types.Message)),
		Rooms:  make(map[string]*Room),
		Games:  make(map[string]*Game),
		Redis:  rds,
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
func (server *WebSocketServer) BroadGameMessage(game *Game, message *types.Message) {
	for client := range game.Players {
		if err := client.conn.WriteJSON(message); err != nil {
			fmt.Println("Broadcast Error:", err)
		}
	}
}
func (server *WebSocketServer) BroadRoomMessage(room *Room, message *types.Message) {
	for client := range room.Players {
		if err := client.conn.WriteJSON(message); err != nil {
			fmt.Println("Broadcast Error:", err)
		}
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
func (server *WebSocketServer) BroadAllClientsMessage(message *types.Message) {

	server.clients.Range(func(key, client interface{}) bool {
		if c, ok := client.(*Client); ok {
			c.conn.WriteJSON(message)
			return true
		}
		return false
	})
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
		//room, err = server.CreateRoom(roomID, client)
		//server.Rooms[roomID] = room
		response.Error = "房间不存在..."
		server.SendMessageToClient(client, response)
		return
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
