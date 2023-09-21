package types

import "time"

type Message struct {
	Type  string                 `json:"type"`
	From  *ClientInfo            `json:"from"`
	To    *ClientInfo            `json:"to"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Error string                 `json:"error"`
}

type ClientInfo struct {
	ID       string `json:"id"`
	RoomID   string `json:"room_id,omitempty"`
	RoomName string `json:"room_name,omitempty"`
	Name     string `json:"name,omitempty"`
	GameID   string `json:"game_id,omitempty"`
}

type GameInfo struct {
	ID            string          `json:"id"`
	RoomID        string          `json:"room_id"`
	RoomName      string          `json:"room_name"`
	Round         int             `json:"round"`      // 新增字段，表示当前回合
	MaxRounds     int             `json:"max_rounds"` // 新增字段，表示最大回合数 12
	Players       map[string]bool `json:"players"`
	PlayerActions map[string]bool `json:"player_actions"` // 新增字段，记录每个玩家是否已操作
	CreatedTime   time.Time       `json:"created_time"`
	EndTime       time.Time       `json:"end_time"`
	CreatedUser   string          `json:"created_user"`
}

type RoomInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"room_name"`
	Players     map[string]bool `json:"players"`
	CreatedTime time.Time       `json:"created_time"`
	CreatedUser string          `json:"created_user"`
	IsGameStart bool            `json:"is_game_start"` // 是否已经开始游戏
}

var (
	START_THROWS = "start_throws"
	SET_SCORE    = "set_score"
	GAME_CREATED = "game_created"
	JOIN_ROOM    = "join_room"
	ROOMLIST     = "room_list"
	ROOMINFO     = "room_info"
	ROOM_CREATED = "room_created"
	GETGAME      = "get_game" // 获取游戏详情
	LEAVE_ROOM   = "leave_room"
)

var MsgTypeMap = map[string]string{
	START_THROWS: "开始投掷骰子",
	SET_SCORE:    "设置分数",
	GAME_CREATED: "游戏创建",
}

var (
	ONE    = "one"
	TWO    = "two"
	THREE  = "three"
	FOUR   = "four"
	FIVE   = "five"
	SIX    = "six"
	REWARD = "reward" // 奖励分
	ALL    = "all"
	STTH   = "stth" // 四骰同花
	HL     = "hl"   // 葫芦
	DS     = "ds"   // 大顺
	XS     = "xs"   // 小顺
	KT     = "kt"   // 快艇
	Sum    = "sum"  // 总和
)
