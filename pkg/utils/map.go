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

func MapToString(args map[string]interface{}, key string) string {
	//args := make(map[string]interface{})
	// 假设 locked_indexs 是一个 []interface{} 类型的切片
	str, ok := args[key].(string)
	if !ok {
		log.Fatal("args key is not a []interface{}", key)
	}

	return str
}
