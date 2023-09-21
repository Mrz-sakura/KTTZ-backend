package socket

import "app-bff/socket/types"

func (server *WebSocketServer) InitRoutes() {

	server.routes[types.START_THROWS] = server.HandlerStartThrows
	server.routes[types.SET_SCORE] = server.HandlerSetScore
	server.routes[types.GAME_CREATED] = server.HandleCreateGame
	server.routes[types.ROOM_CREATED] = server.HandleCreateRoom
	server.routes[types.ROOMLIST] = server.HandleGetRoomList
	server.routes[types.ROOMINFO] = server.HandleGetRoomList
	server.routes[types.LEAVE_ROOM] = server.LeaveRoom
	//server.routes[types.GETGAME] = server.HandleGetRoomList
}
