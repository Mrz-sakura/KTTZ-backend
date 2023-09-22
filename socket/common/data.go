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

// 检测四骰同花
func CheckSTTH(arr []int) int {
	// 创建一个 map 用于统计每个元素的出现次数
	elementCount := make(map[int]int)
	var isSTTH bool
	var sum int

	// 遍历数组并统计每个元素的出现次数
	for _, element := range arr {
		elementCount[element]++
	}

	// 检查是否有任何元素出现了4次
	for _, count := range elementCount {
		if count == 4 {
			isSTTH = true
		}
	}
	if isSTTH {
		for i := 0; i < len(arr); i++ {
			sum += arr[i]
		}
	}

	return sum
}
func ChcekHL(arr []int) int {
	// 创建一个 map 用于统计每个数字的出现次数
	elementCount := make(map[int]int)

	// 遍历数组并统计每个数字的出现次数
	for _, element := range arr {
		elementCount[element]++
	}

	// 初始化标志变量
	hasThree := false
	hasTwo := false

	// 遍历 map，检查是否有3个相同的数字和2个相同的数字同时出现
	for _, count := range elementCount {
		if count == 3 {
			hasThree = true
		} else if count == 2 {
			hasTwo = true
		}
	}
	var sum int
	if hasTwo && hasThree {
		for i := 0; i < len(arr); i++ {
			sum += arr[i]
		}
		return sum
	}
	// 返回结果
	return 0
}

func CheckDS(arr []int) int {
	if len(arr) < 5 {
		return 0
	}
	for i := 0; i <= len(arr)-5; i++ {
		if (arr[i] == 1 && arr[i+1] == 2 && arr[i+2] == 3 && arr[i+3] == 4 && arr[i+4] == 5) ||
			(arr[i] == 2 && arr[i+1] == 3 && arr[i+2] == 4 && arr[i+3] == 5 && arr[i+4] == 6) {
			return 30
		}
	}
	return 0
}

func CheckXS(arr []int) int {
	if len(arr) < 4 {
		return 0
	}
	for i := 0; i <= len(arr)-4; i++ {
		if (arr[i] == 1 && arr[i+1] == 2 && arr[i+2] == 3 && arr[i+3] == 4) ||
			(arr[i] == 2 && arr[i+1] == 3 && arr[i+2] == 4 && arr[i+3] == 5) ||
			(arr[i] == 3 && arr[i+1] == 4 && arr[i+2] == 5 && arr[i+3] == 6) {
			return 15
		}
	}
	return 0
}

// 检测快艇
func CheckKT(arr []int) int {
	// 创建一个 map 用于统计每个元素的出现次数
	elementCount := make(map[int]int)
	var isKT bool

	// 遍历数组并统计每个元素的出现次数
	for _, element := range arr {
		elementCount[element]++
	}

	// 检查是否有任何元素出现了4次
	for _, count := range elementCount {
		if count == 5 {
			isKT = true
		}
	}
	if isKT {
		return 50
	}

	return 0
}
