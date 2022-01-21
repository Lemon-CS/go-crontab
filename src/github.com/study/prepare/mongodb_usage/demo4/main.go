package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// jobName过滤条件
type FindByJobName struct {
	JobName string `bson:"jobName"` // JobName赋值为job10
}

func main() {

	var (
		err        error
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		record     *LogRecord
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

	// 4, 按照jobName字段过滤, 想找出jobName=job10, 找出5条
	cond := &FindByJobName{JobName: "job10"}

	// 5, 查询（过滤 +翻页参数）
	findOptions := new(options.FindOptions)
	findOptions.SetLimit(2)
	cursor, err := collection.Find(context.TODO(), cond, findOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 延迟释放游标
	defer cursor.Close(context.TODO())

	// 6, 遍历结果集
	for cursor.Next(context.TODO()) {
		// 定义一个日志对象
		record = &LogRecord{}

		// 反序列化bson到对象
		err := cursor.Decode(record)
		if err != nil {
			fmt.Println(err)
			return
		}
		// 把日志行打印出来
		fmt.Println(*record)
	}

}
