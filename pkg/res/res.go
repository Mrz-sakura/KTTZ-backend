package res

import (
	"app-bff/pkg/errorss"
	"app-bff/pkg/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`              // 业务状态码
	Msg     string      `json:"msg"`               // 用户可见错误信息
	Data    interface{} `json:"data,omitempty"`    // 数据
	Reason  string      `json:"reason,omitempty"`  // 失败原因,开发用
	Request interface{} `json:"request,omitempty"` // 请求体,开发用
}

type Request struct {
	Path string      `json:"path"`
	Body interface{} `json:"body"`
}

var resOk = Response{
	Code: types.RESPONSE_OK,
	Msg:  types.RESPONSE_MEG_MAP[types.RESPONSE_OK],
}

var resError = Response{
	Code: errorss.SYSTEM_ERROR,
	Msg:  errorss.ERR_MSG_MAP[errorss.SYSTEM_ERROR],
}

func UnAuth(c *gin.Context) {
	r := Response{
		Code: errorss.UNAUTHORIZED,
		Msg:  errorss.ERR_MSG_MAP[errorss.UNAUTHORIZED],
	}

	c.AbortWithStatusJSON(http.StatusOK, r)
}

func Ok(c *gin.Context, data interface{}) {
	r := Response{
		Code: types.RESPONSE_OK,
		Msg:  types.RESPONSE_MEG_MAP[types.RESPONSE_OK],
		Data: data,
	}

	c.JSON(http.StatusOK, r)
}

func Error(c *gin.Context, err error) {
	r := resError
	e := errorss.FromError(err)

	if e != nil {
		r.Code = e.Code()
		r.Msg = e.Msg()
	}

	r.Reason = err.Error()

	c.AbortWithStatusJSON(http.StatusOK, r)
}
