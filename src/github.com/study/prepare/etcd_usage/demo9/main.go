/*
分布式乐观锁
*/
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {

	var (
		config         clientv3.Config
		client         *clientv3.Client
		err            error
		kv             clientv3.KV
		lease          clientv3.Lease
		leaseId        clientv3.LeaseID
		keepResp       *clientv3.LeaseKeepAliveResponse
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		leaseGrantResp *clientv3.LeaseGrantResponse
		txnResp        *clientv3.TxnResponse
	)

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// lease实现锁自动过期:以及续约
	// op操作
	// txn事务: if else then

	// 1, 上锁 (创建租约, 自动续租, 拿着租约去抢占一个key)
	lease = clientv3.NewLease(client)

	// 申请一个5秒的租约
	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
		return
	}

	// 拿到租约的ID
	leaseId = leaseGrantResp.ID

	// 准备一个用于取消自动续租的context
	ctx, cancelFunc := context.WithCancel(context.TODO())

	// 确保函数退出后, 自动续租会停止
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)

	// 5秒后会取消自动续租
	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		fmt.Println(err)
		return
	}

	// 处理续约应答的协程
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepRespChan == nil {
					fmt.Println("租约已经失效了")
					goto END
				} else { // 每秒会续租一次, 所以就会受到一次应答
					fmt.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
	END:
	}()

	//  if 不存在key， then 设置它, else 抢锁失败
	kv = clientv3.NewKV(client)

	txn := kv.Txn(context.TODO())

	// 定义事务

	// 如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job9", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/lock/job9")) // 否则抢锁失败

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	// 判断是否抢到了锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	// 2, 处理业务
	fmt.Println("处理任务")
	time.Sleep(5 * time.Second)

	// 3, 释放锁(取消自动续租, 释放租约)
	// defer 会把租约释放掉, 关联的KV就被删除了

}
