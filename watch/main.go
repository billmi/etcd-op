package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"os"
	"time"
)

const (
	NewLeaseErr  = 101
	LeasTtlErr   = 102
	KeepAliveErr = 103
	PutErr       = 104
	GetErr       = 105
	RevokeErr    = 106
)

func main() {

	var conf = clientv3.Config{
		Endpoints:   []string{"xxxxxxx:2379"},
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(conf)
	defer client.Close()
	if err != nil {
		fmt.Printf("conn fail:\n", err.Error())
		os.Exit(NewLeaseErr)
	}

	lease := clientv3.NewLease(client)

	leaseResp, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Printf("Grant fail:%s\n", err.Error())
		os.Exit(LeasTtlErr)
	}

	//设置续租
	leaseID := leaseResp.ID
	leaseRespChan, err := lease.KeepAlive(context.TODO(), leaseID)
	if err != nil {
		fmt.Printf("Keep fail:%s\n", err.Error())
		os.Exit(KeepAliveErr)
	}

	//监听租约
	go func() {
		for {
			select {
			case leaseKeepResp := <-leaseRespChan:
				if leaseKeepResp == nil {
					fmt.Printf("keep close\n")
					return
				} else {
					fmt.Printf("keep succ\n")
					goto END
				}
			}
		END:
			fmt.Println("go to ! \r\n")
		}
	}()

	//listen key
	//ctx1, _ := context.WithTimeout(context.TODO(),20)
	go func() {
		wc := client.Watch(context.TODO(), "/job/v3/1", clientv3.WithPrevKV())
		for v := range wc {
			for _, e := range v.Events {
				fmt.Printf("type:%v kv:%v  prevKey:%v \n ", e.Type, string(e.Kv.Key), e.PrevKv)
			}
		}
	}()

	kv := clientv3.NewKV(client)
	//leaseID to PUT
	putResp, err := kv.Put(context.TODO(), "/job/v3/1", "koock", clientv3.WithLease(leaseID))
	//putResp, err := kv.Put(context.TODO(), "/job/v3/1", "new value", )
	if err != nil {
		fmt.Printf("Put fail：%s", err.Error())
		os.Exit(PutErr)
	}
	fmt.Printf("%v\n", putResp.Header)

	//续约监听不退出

	//time.Sleep(2 * time.Second)
	//_, err = lease.Revoke(context.TODO(), leaseID)

	for {
		getResp, err := kv.Get(context.TODO(), "/job/v3/1")
		if err != nil {
			fmt.Printf("Get fail：%s", err.Error())
			os.Exit(GetErr)
		}
		fmt.Println("\r\n ============ \r\n")
		fmt.Printf("%s", "End !")
		fmt.Printf("%v", getResp.Kvs)
		time.Sleep(1 * time.Second)
	}

	//if err != nil {
	//	fmt.Printf("Revoke fail :%s\n", err.Error())
	//	os.Exit(RevokeErr)
	//}
	//fmt.Printf("Revoke Succ")
	//getResp, err := kv.Get(context.TODO(), "/job/v3/1")
	//if err != nil {
	//	fmt.Printf("Get fail：%s", err.Error())
	//	os.Exit(GetErr)
	//}
	//fmt.Printf("%s", "End !")
	//fmt.Printf("%v", getResp.Kvs)
	//time.Sleep(20 * time.Second)

}
