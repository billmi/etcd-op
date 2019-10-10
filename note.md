

```

------------------------

clientv3.WithPrevKV() //可获取上次的值  Put
fmt.Println("Revision:", putResp.Header.Revision)
    if putResp.PrevKv != nil {  
        fmt.Println("PrevValue:", string(putResp.PrevKv.Value))
}

------------------------

clientv3.WithCountOnly() //可以获取数量 Get
fmt.Println(getResp.Kvs, getResp.Count)

------------------------

clientv3.WithPrefix() //读取前缀 Get

------------------------


// 删除KV
if delResp, err = kv.Delete(context.TODO(), "/cron/jobs/job1", clientv3.WithFromKey(), 	clientv3.WithLimit(2)); err != nil {
		fmt.Println(err)
		return
}

// 被删除之前的value是什么
if len(delResp.PrevKvs) != 0 {
	for _, kvpair = range delResp.PrevKvs {
		fmt.Println("删除了:", string(kvpair.Key), string(kvpair.Value))
	}
}

------------------------

//watch模式监听
clientv3.WithRev(watchStartRevision)
watch模式下有通道会返回数据

------------------------
//Op可以做连贯操作
// 创建Op: operation
putOp = clientv3.OpPut("/cron/jobs/job8", "123123123")

// 执行OP
if opResp, err = kv.Do(context.TODO(), putOp); err != nil {
	fmt.Println(err)
	return
}

getOp = clientv3.OpGet("/cron/jobs/job8")
// 执行OP
if opResp, err = kv.Do(context.TODO(), getOp); err != nil {
fmt.Println(err)
	return
}

// kv.Do(op)
// kv.Put
// kv.Get
// kv.Delete

// 打印
fmt.Println("数据Revision:", opResp.Get().Kvs[0].ModRevision)	// create rev == mod rev
fmt.Println("数据value:", string(opResp.Get().Kvs[0].Value))

------------------------

//****** 这段代码比较重要 ******//
package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"fmt"
	"context"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		lease clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
		keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
		keepResp *clientv3.LeaseKeepAliveResponse
		ctx context.Context
		cancelFunc context.CancelFunc
		kv clientv3.KV
		txn clientv3.Txn
		txnResp *clientv3.TxnResponse
	)

	// 客户端配置
	config = clientv3.Config{
		Endpoints: []string{"xxxxxxx:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// lease实现锁自动过期:
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
	ctx, cancelFunc = context.WithCancel(context.TODO())

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
			case keepResp = <- keepRespChan:
				if keepRespChan == nil {
					fmt.Println("租约已经失效了")
					goto END
				} else {	// 每秒会续租一次, 所以就会受到一次应答
					fmt.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
	END:
	}()

	//  if 不存在key， then 设置它, else 抢锁失败
	kv = clientv3.NewKV(client)

	// 创建事务
	txn = kv.Txn(context.TODO())

	// 定义事务

	// 如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job9", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/lock/job9")) // 否则抢锁失败

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return // 没有问题
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
//****** 这段代码比较重要 ******//





```


 
 