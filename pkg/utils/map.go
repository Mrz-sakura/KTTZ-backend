package utils

import (
	"fmt"
	"log"
)

func MapToSliceString(args map[string]interface{}, key string) []string {
	interfaces, ok := args[key].([]interface{})
	if !ok {
		fmt.Println("args KEY ERROR :", key)
		return make([]string, 0)
	}
	indexStrings := make([]string, len(interfaces))

	for i, val := range interfaces {
		indexStrings[i], ok = val.(string)
		if !ok {
			fmt.Println("args i ERROR :", i)
			return make([]string, 0)
		}
	}
	return indexStrings
}
func MapToSliceInt(args map[string]interface{}, key string) []int {
	fmt.Println(args[key])
	interfaces, ok := args[key].([]interface{})
	if !ok {
		fmt.Println("Error:", key)
		return make([]int, 0)
	}
	indexInt := make([]int, len(interfaces))

	for i, val := range interfaces {
		if intValue, ok := val.(int); ok {
			indexInt[i] = intValue
		} else if floatValue, ok := val.(float64); ok { // If not int, try float64
			indexInt[i] = int(floatValue)
		} else {
			fmt.Println("转化失败:", i)
			return make([]int, 0)
		}
	}
	return indexInt
}

func MapToString(args map[string]interface{}, key string) string {
	//args := make(map[string]interface{})
	// 假设 locked_indexs 是一个 []interface{} 类型的切片
	str, ok := args[key].(string)
	if !ok {
		log.Fatal("args key is not a []interface{}", key)
	}

	return str
}
