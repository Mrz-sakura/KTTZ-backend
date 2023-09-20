package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type List struct {
	HTTPGETList string `path:"/user/list" method:"GET"`
}

// 控制器的方法 分页查询
func (api *List) GETList(c *gin.Context) {
	users := []int{1, 2, 3}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": users,
	})
}
