package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println("config app inited")
}

func InitMysql() {
	//自定义日志模板，打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢Sql阈值
			LogLevel:                  logger.Info, // 级别
			IgnoreRecordNotFoundError: true,        // 彩色打印
		},
	)

	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("config mysql inited")
	// user := models.UserBasic{}
	// DB.Find(&user)
	// fmt.Println(user)
}

func InitRedis() {
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.dns"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConns"),
	})
	pong, err := Red.Ping(Red.Context()).Result()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("config redis inited", pong)
	}
}

const (
	PublishKey = "websocket"
)

// 发布消息到redis
func Pubilsh(ctx context.Context, channel string, msg string) error {
	var err error
	fmt.Println("发送消息：", msg)
	err = Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// 订阅redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	fmt.Println("收到消息：", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("收到消息：", msg.Payload)
	return msg.Payload, err
}
