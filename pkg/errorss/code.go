package errorss

const ITOA = 10000
const (
	SYSTEM_ERROR  = ITOA + 0
	BAD_PARAMETER = ITOA + 1

	UNAUTHORIZED = ITOA + 10000 // 未登录

	DATA_PARSE_ERROR = ITOA + 10 // 一般是json,marshal出现的问题
)

var ERR_MSG_MAP = map[int]string{
	SYSTEM_ERROR: "系统错误",
	UNAUTHORIZED: "登录失效,请重新登录",

	DATA_PARSE_ERROR:   "数组解析失败,请稍后重试...",
	SOCKET_REDIS_ERROR: "数据设置失败,请稍后重试...",
}

const SCOKET_IOTA = 80000
const (
	SOCKET_ERROR        = SCOKET_IOTA + 0
	SOCKET_NO_PERMISSON = SCOKET_IOTA + 1 // 无权限创建房间或者游戏

	SOCKET_REDIS_ERROR = SCOKET_IOTA + 1000 // redis 错误
)
