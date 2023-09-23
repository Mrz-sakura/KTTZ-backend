package common

import (
	"app-bff/pkg/config"
	"fmt"
)

func GenerateData(key string, value interface{}, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}
	data[key] = value

	return data
}

// gameID 游戏的ID,partID 对局回合的ID,userID 用户ID
func GetDiceKey(gameID string, roundID int, userID string) string {
	return fmt.Sprintf("%s_%s_%d_%s", config.GetString("redis_key.dice_key"), gameID, roundID, userID)
}

func GetDiceRoundsKey(gameID string, userID string) string {
	return fmt.Sprintf("%s_%s_%s", config.GetString("redis_key.dice_round_key"), gameID, userID)
}
func GetDiceRoundsLocksKey(gameID string, roundID int, userID string) string {
	return fmt.Sprintf("%s_%s_%d_%s", config.GetString("redis_key.dice_round_locks_key"), gameID, roundID, userID)
}

func GetScoreKey(gameID string, userID string) string {
	return fmt.Sprintf("%s_%s_%s", config.GetString("redis_key.dice_score_key"), gameID, userID)
}
func GetScoreValueKey(gameID string, userID string) string {
	return fmt.Sprintf("%s_%s_%s", config.GetString("redis_key.dice_score_value_key"), gameID, userID)
}
func GetGameCreatedKey(gameID string) string {
	return fmt.Sprintf("%s_%s", config.GetString("redis_key.game_created_key"), gameID)
}

// 获取房间redis key
func GetRoomCreatedKey(roomID string) string {
	return fmt.Sprintf("%s_%s", config.GetString("redis_key.room_created_key"), roomID)
}
func GetRoomListKey() string {
	return fmt.Sprintf("%s", config.GetString("redis_key.room_list_key"))
}
func GetGameListKey() string {
	return fmt.Sprintf("%s", config.GetString("redis_key.game_list_key"))
}
