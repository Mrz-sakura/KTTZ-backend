package socket

import "app-bff/socket/types"

func (server *WebSocketServer) InitRoutes() {

	server.routes[types.START_THROWS] = server.StartThrowsHandler
	server.routes[types.SET_SCORE] = server.SetScoreHandler
	server.routes[types.GAME_CREATED] = server.HandleCreateGame
}
