package errorss

const ITOA = 10000
const (
	SYSTEM_ERROR  = ITOA + 0
	BAD_PARAMETER = ITOA + 1

	UNAUTHORIZED = ITOA + 10000 // 未登录
)

var ERR_MSG_MAP = map[int]string{
	SYSTEM_ERROR: "系统错误",
	UNAUTHORIZED: "登录失效,请重新登录",
}
