package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// startTime小于某时间
// {"$lt": timestamp}
type TimeBeforeCond struct {
	Before int64 `bson:"$lt"`
}

// {"timePoint.startTime": {"$lt": timestamp} }
type DeleteCond struct {
	beforeCond TimeBeforeCond `bson:"timePoint.startTime"`
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

	// 4, 要删除开始时间早于当前时间的所有日志($lt是less than)
	//  delete({"timePoint.startTime": {"$lt": 当前时间}})
	delCond := &DeleteCond{beforeCond: TimeBeforeCond{Before: time.Now().Unix()}}
	// 执行删除
	result, err := collection.DeleteMany(context.TODO(), delCond)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("删除的行数:", result.DeletedCount)

}
