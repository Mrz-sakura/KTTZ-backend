package socket

func (server *WebSocketServer) GetPlayerClientByID(Players map[string]bool) map[*Client]bool {
	resp := make(map[*Client]bool)
	for k := range Players {
		player, ok := server.clients.Load(k)
		if !ok {
			continue
		}
		resp[player.(*Client)] = true
	}
	return resp
}

func (server *WebSocketServer) GetPlayersByClient(Players map[*Client]bool) map[string]bool {
	players := make(map[string]bool)
	for player, v := range Players {
		players[player.ID] = v
	}
	return players
}

func (server *WebSocketServer) GetPlayersSliceByClient(Players []*Client) []string {
	if Players == nil {
		return []string{}
	}
	
	players := make([]string, len(Players))
	for _, v := range Players {
		players = append(players, v.ID)
	}
	return players
}

func (server *WebSocketServer) GetPlayerActionsByClient(PlayerActions map[*Client]bool) map[string]bool {
	players := make(map[string]bool)
	for player, v := range PlayerActions {
		players[player.ID] = v
	}
	return players
}
