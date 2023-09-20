package socket

import "app-bff/socket/types"

func (ws *WebSocketServer) InitRoutes() {

	ws.routes[types.START_THROWS] = ws.StartThrowsHandler
	ws.routes[types.SET_SCORE] = ws.SetScoreHandler
	ws.routes[types.GAME_CREATED] = ws.CreateGame
}
