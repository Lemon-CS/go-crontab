package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// 任务的执行时间点
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

// 一条日志
type LogRecord struct {
	JobName   string    `bson:"jobName"`   // 任务名
	Command   string    `bson:"command"`   // shell命令
	Err       string    `bson:"err"`       // 脚本错误
	Content   string    `bson:"content"`   // 脚本输出
	TimePoint TimePoint `bson:"timePoint"` // 执行时间点
}

func main() {

	var (
		err        error
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
	)

	// 1, 建立连接
	// 设置客户端连接配置
	opt := new(options.ClientOptions)
	credential := options.Credential{
		Username: "admin",
		Password: "123456",
	}
	opt.SetAuth(credential)
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions, opt)

	if err != nil {
		fmt.Println("连接失败！")
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("ping 失败")
	}

	// 2, 选择数据库my_db
	database = client.Database("cron")

	// 3, 选择表my_collection
	collection = database.Collection("log")

	// 4, 插入记录(bson)
	record := &LogRecord{
		JobName:   "job10",
		Command:   "echo hello",
		Err:       "",
		Content:   "hello",
		TimePoint: TimePoint{StartTime: time.Now().Unix(), EndTime: time.Now().Unix() + 10},
	}

	logArr := []interface{}{record, record, record}

	result, err := collection.InsertMany(context.TODO(), logArr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 推特很早的时候开源的，tweet的ID
	// snowflake: 毫秒/微秒的当前时间 + 机器的ID + 当前毫秒/微秒内的自增ID(每当毫秒变化了, 会重置成0，继续自增）
	for _, insertId := range result.InsertedIDs {
		// 拿着interface{}， 反射成objectID
		docId := insertId.(primitive.ObjectID)
		fmt.Println("自增ID:", docId.Hex())

	}

}
