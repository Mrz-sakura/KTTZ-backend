package route

import (
	"app-bff/pkg/config"
	"app-bff/socket"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// 路由结构体
type Route struct {
	path       string        //url路径
	httpMethod string        //http方法 get post
	Method     reflect.Value //方法路由
}

// 路由集合
var Routes = []Route{}

func InitRouter() *gin.Engine {
	//初始化路由
	r := gin.Default()
	// 绑定socket
	InitSocket(r)
	//绑定基本路由，访问路径：/User/List
	Bind(r)
	return r
}

func Run(r *gin.Engine) {
	host := fmt.Sprintf("%s", config.GetString("server.host"))
	port := fmt.Sprintf(":%d", config.GetInt("server.port"))

	if err := r.Run(fmt.Sprintf("%s%s", host, port)); err != nil {
		log.Fatal("Error running")
	}
}

func InitSocket(r *gin.Engine) {
	server := socket.NewWebSocketServer()

	r.GET("/ws", server.HandleClients)
}

// 注册控制器
func Register(controller interface{}) bool {
	var valueEle reflect.Value

	value := reflect.ValueOf(controller)
	ctrlName := reflect.TypeOf(controller)
	module := ctrlName.String()

	if ctrlName.Kind() == reflect.Ptr {
		valueEle = value.Elem()
		ctrlName = ctrlName.Elem()
	}

	//遍历方法
	for i := 0; i < valueEle.NumField(); i++ {
		methodName := ctrlName.Field(i).Name

		path := ctrlName.Field(i).Tag.Get("path")
		httpMethod := ctrlName.Field(i).Tag.Get("method")

		if strings.HasPrefix(methodName, "HTTP") {
			methodName = strings.TrimPrefix(methodName, "HTTP")
		}

		method := value.MethodByName(methodName)

		if !method.IsValid() {
			fmt.Println("module:==>" + module + "==> method not found:" + methodName)
			continue
		}

		route := Route{path: path, Method: method, httpMethod: httpMethod}
		Routes = append(Routes, route)
	}
	fmt.Println("Routes=", Routes)
	return true
}

// 绑定路由 m是方法GET POST等
// 绑定基本路由
func Bind(e *gin.Engine) {
	for _, route := range Routes {
		if route.httpMethod == "GET" {
			e.GET(route.path, match(route.path, route))
		}
		if route.httpMethod == "POST" {
			e.POST(route.path, match(route.path, route))
		}
		if route.httpMethod == "PUT" {
			e.PUT(route.path, match(route.path, route))
		}
	}
}

// 根据path匹配对应的方法
func match(path string, route Route) gin.HandlerFunc {
	return func(c *gin.Context) {
		fields := strings.Split(path, "/")
		fmt.Println("fields,len(fields)=", fields, len(fields))
		if len(fields) < 3 {
			return
		}

		if len(Routes) > 0 {
			arguments := make([]reflect.Value, 1)
			arguments[0] = reflect.ValueOf(c) // *gin.Context
			//reflect.ValueOf(method).Method(0).Call(arguments)
			route.Method.Call(arguments)
		}
	}
}
