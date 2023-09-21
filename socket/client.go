package socket

import "github.com/gorilla/websocket"

type Client struct {
	conn     *websocket.Conn
	ID       string `json:"id"`
	Room     *Room  `json:"room"`
	Game     *Game  `json:"game"`
	RoomName string `json:"room_name"`
	Name     string `json:"name"`
	GameID   string `json:"game_id"`
}

func (client *Client) DoCompleted(completedChan chan *Client) {
	// 这里可以添加逻辑以处理玩家的行动
	// ...

	// 假设玩家完成了回合
	completedChan <- client
}
