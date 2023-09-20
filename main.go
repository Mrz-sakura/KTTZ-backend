package main

import (
	_ "app-bff/app/controller"
	"app-bff/pkg/config"
	"app-bff/route"
	"log"
)

func main() {

	err := config.InitializeConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	//加载路由
	route.Run(route.InitRouter()) //":8000"
}
