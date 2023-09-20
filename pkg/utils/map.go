package utils

import "log"

func MapToSliceString(args map[string]interface{}, key string) []string {
	//args := make(map[string]interface{})
	// 假设 locked_indexs 是一个 []interface{} 类型的切片
	interfaces, ok := args[key].([]interface{})
	if !ok {
		log.Fatal("args key is not a []interface{}", key)
	}

	// 然后将 []interface{} 转换为 []string
	indexStrings := make([]string, len(interfaces))
	for i, val := range interfaces {
		indexStrings[i], ok = val.(string)
		if !ok {
			log.Fatalf("Element at index %d is not a string", i)
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
