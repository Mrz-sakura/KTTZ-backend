package mod

import (
	"app-bff/pkg/config"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func GetRedisClient() (*redis.Client, error) {

	ctx := context.Background()

	// 初始化 Redis 客户端
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GetString("redis.host"), config.GetInt("redis.port")), // Redis 服务器地址和端口
		DB:       0,                                                                                 // 使用默认数据库
		Password: config.GetString("redis.password"),
	}

	client := redis.NewClient(opts)

	// 尝试连接到 Redis 服务器
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Could not connect to Redis: ", err)
		return nil, err
	}

	// 记得在程序结束前关闭 Redis 客户端
	//defer client.Close()

	return client, nil
}
