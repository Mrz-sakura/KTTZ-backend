package user

import (
	"app-bff/app/service/user"
	"app-bff/route"
)

// 这里每个controller执行init方法都要注册自动路由
func init() {
	route.Register(&user.List{})
}
