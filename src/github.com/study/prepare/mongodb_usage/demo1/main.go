package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	var (
		err        error
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
	)

	// 1, 建立连接
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		fmt.Println("连接失败！")
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("ping 失败")
	}

	// 2, 选择数据库my_db
	database = client.Database("my_db")

	// 3, 选择表my_collection
	collection = database.Collection("my_collection")

	fmt.Println(collection.Name())

}
