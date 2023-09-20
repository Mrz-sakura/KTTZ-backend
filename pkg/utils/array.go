package utils

import "strconv"

// 辅助函数用于检查一个整数是否存在于一个字符串数组中
func ArrayStrContainsInt(arr []string, x int) bool {
	for _, n := range arr {
		if n == strconv.Itoa(x) {
			return true
		}
	}
	return false
}
func ArrayIntContainsInt(arr []int, x int) bool {
	for _, n := range arr {
		if n == x {
			return true
		}
	}
	return false
}

func ArrayStrToArrayInt(arr []string) ([]int, error) {
	var err error
	ret := make([]int, len(arr))
	for i, n := range arr {
		ret[i], err = strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
