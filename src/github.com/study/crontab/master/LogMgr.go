package master

import (
	"context"
	"fmt"
	"go-crontab/src/github.com/study/crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var client *mongo.Client

	// 设置客户端连接配置
	opt := new(options.ClientOptions)
	credential := options.Credential{
		Username: "admin",
		Password: "123456",
	}
	opt.SetAuth(credential)
	opt.SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout) * time.Millisecond)

	clientOptions := options.Client().ApplyURI(G_config.MongodbUri)
	client, err = mongo.Connect(context.TODO(), clientOptions, opt)

	if err != nil {
		fmt.Println("连接失败！")
	}

	//   选择db和collection
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}

	return
}

// 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)

	// len(logArr)
	logArr = make([]*common.JobLog, 0)

	// 过滤条件
	filter = &common.JobLogFilter{
		JobName: name,
	}

	// 按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{
		SortOrder: -1,
	}

	// 查询
	// 查询（过滤 +翻页参数）
	findOptions := new(options.FindOptions)
	findOptions.SetSort(logSort)
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findOptions); err != nil {
		return
	}

	// 延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法
		}

		logArr = append(logArr, jobLog)

	}
	return
}
