# go-crontab整体架构和设计
## 实现目标：
实现一个分布式crontab系统。用户可以通过前端页面配置任务和cron表达式和命令来执行定时任务，相比较linux自带的crontab来说，本项目可以方便看到执行结果，且分布式部署可以避免单点问题，用户不用登陆到各个机器去配置任务，操作方便。同时用户可以通过页面查看任务执行的情况。当然，目前做的还比较简单，对任务的执行时间没有超时机制，但提供了手动的删除和强杀正在执行的任务操作。

## 最终效果
![](https://img-blog.csdnimg.cn/20190408114017412.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L29YaWFvQnVEaW5n,size_16,color_FFFFFF,t_70)

![](https://img-blog.csdnimg.cn/20190408114017412.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L29YaWFvQnVEaW5n,size_16,color_FFFFFF,t_70)

## 整体架构图


1. 客户端请求无状态的Master集群，将任务保存到Etcd中，Master可以添加、查询任务，查询任务的执行日志
2. 然后Etcd将任务同步到Worker集群，所有的worker都拿到全部的任务列表
3. 通过分布式乐观锁互斥的控制多个worker争抢一个任务
4. 然后将任务执行的日志保存在MongoDB中

## 系统架构
主要分为master和worker两个角色。通过etcd来作为服务发现和分布式锁的实现。MongoDB作为数据量存储日志信息，方便查询执行结果。同时也可以通过本地log日志查看模块的执行情况。
master通过跟前端交互获取用户的任务操作信息，通过与etcd交互和mongodb交互来完成建立、删除、编辑、强杀、查看健康woker节点以及查看日志等功能。
woker通过监控etcd的节点变化来执行任务的执行、强杀等操作，同时通过etcd来实现自身服务的注册功能以及吧执行结果写入MongoDB作为日志存储。

- 利用etcd同步全量任务列表到所有的worker节点
- 每个worker独立调度全量任务，无需和Master产生直接的RPC，避免网络故障
- 每个worker利用分布式锁抢占，解决并发调度相同任务的问题

![](https://img-blog.csdnimg.cn/20190408114226494.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L29YaWFvQnVEaW5n,size_16,color_FFFFFF,t_70)

## Master功能

- 任务管理HTTP接口：新建、修改、查看、删除任务
- 任务日志HTTP接口：查看任务执行历史日志
- 任务控制HTTP接口：提供强制结束任务的接口
- 实现web管理页面，前后端分离

![](https://img-blog.csdnimg.cn/20190408115015775.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L29YaWFvQnVEaW5n,size_16,color_FFFFFF,t_70)

### 任务管理

Etcd结构

```json
/cron/jobs/任务名 -> {
  name,    // 任务名
  command, // shell命令
  cronExpr // cron表达式
}
```

### 任务日志

MongoDB结构

```json
{
	JobName   string    `bson:"jobName"`   // 任务名
	Command   string    `bson:"command"`   // shell命令
	Err       string    `bson:"err"`       // 脚本错误
	Content   string    `bson:"content"`   // 脚本输出
	TimePoint TimePoint `bson:"timePoint"` // 执行时间点
}
```

请求MongoDB，按任务名查看最近的执行日志

### 任务控制

1. 向etcd中写入

```go
/cron/killer/任务名 -> ""
```

2. worker会监听`/cron/killer/`目录下的put修改操作
3. Master将要结束的任务名put在`/cron/killer/`目录下，触发worker立即结束shell任务



## Worker功能

### 任务同步

监听etcd中`/cron/jobs/`目录的变化，有变化就说明有添加或者修改任务

### 任务调度

基于cron表达式计算，触发过期任务

### 任务执行

协程池并发执行多任务，基于etcd分布式锁抢占

### 日志捕获

捕获任务执行输出，并保存到MongoDB

![](https://img-blog.csdnimg.cn/20190408114948343.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L29YaWFvQnVEaW5n,size_16,color_FFFFFF,t_70)



### 监听协程

- 利用watch API，监听`/cron/jobs/`和`/cron/killer/`目录的变化
- 将变化事件通过channel推送给调度协程，更新内存中的任务信息

### 调度协程

- 监听任务变更event，更细内存中维护的任务列表
- 检查cron表达式，扫描到期任务，交给执行协程运行
- 监听任务控制event，强制中断正在执行的子进程
- 监听任务执行result，更新内存中任务状态，投递执行日志

### 执行协程

- 在etcd中抢占分布式乐观锁：`/cron/lock/任务名`
- 抢占成功则通过Command类执行shell任务
- 捕获Command输出并等待子进程结束，将执行结果投递给调度协程

### 日志协程

- 监听调度协程发来的执行日志，放入一个batch中
- 对新batch启动定时器，超时未满自动提交
- 若batch被放满，那么就立即提交，并取消自动提交定时器



## 后续优化

有很多地方有待优化，比如

- 任务执行时间的限制，可以支持配置任务执行的最大时长，超过强杀。
- master目前虽然支持多机部署但是没有主从机制，可以实现master的选主机制，防止并发问题。只有主才能执行etcd 的"写入操作"
- 代码结构上有一定的冗余，可以通过复用以实现精简

