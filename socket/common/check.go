package common

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

	if containsSubset(arr, []int{1, 2, 3, 4, 5}) ||
		containsSubset(arr, []int{2, 3, 4, 5, 6}) {
		return 30
	}

	return 0
}
func CheckXS(arr []int) int {
	if len(arr) < 4 {
		return 0
	}

	if containsSubset(arr, []int{1, 2, 3, 4}) ||
		containsSubset(arr, []int{2, 3, 4, 5}) ||
		containsSubset(arr, []int{3, 4, 5, 6}) {
		return 15
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

func containsSubset(arr []int, subset []int) bool {
	count := 0
	for _, v := range subset {
		for _, a := range arr {
			if a == v {
				count++
				break
			}
		}
	}
	return count == len(subset)
}
